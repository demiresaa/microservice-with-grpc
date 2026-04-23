package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseCapture struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (rc *responseCapture) WriteHeader(code int) {
	if !rc.wrote {
		rc.status = code
		rc.wrote = true
	}
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	if !rc.wrote {
		rc.status = http.StatusOK
		rc.wrote = true
	}
	return rc.ResponseWriter.Write(b)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rc := &responseCapture{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rc, r)

			duration := time.Since(start)

			requestID := GetRequestID(r.Context())

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rc.status,
				"duration_ms", duration.Milliseconds(),
				"request_id", requestID,
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
