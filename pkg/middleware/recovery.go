package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := GetRequestID(r.Context())

					logger.Error("panic recovered",
						"error", err,
						"request_id", requestID,
						"path", r.URL.Path,
						"method", r.Method,
						"stack", string(debug.Stack()),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error":      "internal server error",
						"request_id": requestID,
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
