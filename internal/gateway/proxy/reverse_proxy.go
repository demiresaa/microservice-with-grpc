package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type ReverseProxy struct {
	target      *url.URL
	proxy       *httputil.ReverseProxy
	logger      *slog.Logger
	stripPrefix string
}

func NewReverseProxy(target string, stripPrefix string, logger *slog.Logger) *ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		logger.Error("failed to parse target URL", "target", target, "error", err)
		return nil
	}

	p := httputil.NewSingleHostReverseProxy(targetURL)

	defaultDirector := p.Director
	p.Director = func(req *http.Request) {
		defaultDirector(req)

		if stripPrefix != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, stripPrefix)
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
		}

		req.Host = targetURL.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
	}

	p.ModifyResponse = func(resp *http.Response) error {
		logger.Info("proxy response",
			"method", resp.Request.Method,
			"path", resp.Request.URL.Path,
			"target", targetURL.String(),
			"status", resp.StatusCode,
		)
		return nil
	}

	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("proxy error",
			"method", r.Method,
			"path", r.URL.Path,
			"target", targetURL.String(),
			"error", err,
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"service unavailable","message":"backend service is not responding"}`))
	}

	return &ReverseProxy{
		target:      targetURL,
		proxy:       p,
		logger:      logger,
		stripPrefix: stripPrefix,
	}
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rp.proxy.ServeHTTP(w, r)
	rp.logger.Info("request proxied",
		"method", r.Method,
		"path", r.URL.Path,
		"target", rp.target.String(),
		"duration", time.Since(start).String(),
	)
}
