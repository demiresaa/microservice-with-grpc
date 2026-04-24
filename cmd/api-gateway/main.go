package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/gateway/config"
	gwmiddleware "github.com/suleymankursatdemir/ecommerce-platform/internal/gateway/middleware"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/gateway/proxy"
	"github.com/suleymankursatdemir/ecommerce-platform/pkg/logger"
	appmiddleware "github.com/suleymankursatdemir/ecommerce-platform/pkg/middleware"
)

func main() {
	log := logger.New("api-gateway")

	cfg := config.Load()

	r := chi.NewRouter()

	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Recovery(log))
	r.Use(appmiddleware.Logging(log))
	r.Use(appmiddleware.CORS)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(30 * time.Second))

	r.Get("/gateway/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"api-gateway"}`))
	})

	registerRoutes(r, cfg, log)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("api gateway starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("gateway routes configured")
	for _, route := range cfg.Routes {
		authLabel := "public"
		if route.AuthRequired {
			authLabel = "protected"
		}
		log.Info("route", "path", route.Path, "target", route.Target, "auth", authLabel, "rate_limit", route.RateLimitRPS)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("shutting down", "signal", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("api gateway stopped gracefully")
}

func registerRoutes(r chi.Router, cfg *config.GatewayConfig, log *slog.Logger) {
	for _, route := range cfg.Routes {
		rp := proxy.NewReverseProxy(route.Target, route.StripPrefix, log)
		if rp == nil {
			log.Error("skipping route, failed to create proxy", "path", route.Path, "target", route.Target)
			continue
		}

		rateLimiter := gwmiddleware.NewRateLimiter(route.RateLimitRPS)

		handler := http.Handler(rp)

		if route.AuthRequired {
			handler = appmiddleware.JWTAuth(cfg.JWTSecret)(handler)
		}

		handler = rateLimiter.Middleware(handler)

		path := route.Path
		if !strings.HasSuffix(path, "/*") && !strings.HasSuffix(path, "/") {
			path += "/*"
		}

		r.Handle(path, handler)
	}
}
