package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/segmentio/kafka-go"
)

type HealthChecker struct {
	db          *sql.DB
	kafkaBroker string
}

func NewHealthChecker(db *sql.DB, kafkaBroker string) *HealthChecker {
	return &HealthChecker{
		db:          db,
		kafkaBroker: kafkaBroker,
	}
}

// Health godoc
// @Summary      Servis canli mi kontrol et
// @Description  Servisin ayakta olup olmadigini kontrol eder (liveness)
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /health [get]
func (h *HealthChecker) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}

// Ready godoc
// @Summary      Servis hazir mi kontrol et
// @Description  Veritabani ve Kafka baglantilarini kontrol eder (readiness)
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      503 {object} map[string]interface{}
// @Router       /ready [get]
func (h *HealthChecker) Ready(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	checks := make(map[string]string)

	if err := h.checkDB(r.Context()); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		status = http.StatusServiceUnavailable
	} else {
		checks["database"] = "healthy"
	}

	if err := h.checkKafka(); err != nil {
		checks["kafka"] = "unhealthy: " + err.Error()
		status = http.StatusServiceUnavailable
	} else {
		checks["kafka"] = "healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"status": statusCodeToText(status),
		"checks": checks,
	})
}

func (h *HealthChecker) checkDB(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return h.db.PingContext(pingCtx)
}

func (h *HealthChecker) checkKafka() error {
	dialer := &kafka.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "tcp", h.kafkaBroker)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func statusCodeToText(code int) string {
	if code == http.StatusOK {
		return "ready"
	}
	return "not ready"
}

func RegisterRoutes(r chi.Router, hc *HealthChecker) {
	r.Get("/health", hc.Health)
	r.Get("/ready", hc.Ready)
}
