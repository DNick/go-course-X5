package handler

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"time"
)

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := rand.Text()
			w.Header().Set("X-Request-ID", requestID)
			r = r.WithContext(context.WithValue(r.Context(), "request_id", requestID))

			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(recorder, r)

			logger.Info("request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", recorder.status),
				slog.Duration("duration", time.Since(start)),
				slog.String("request_id", requestID),
			)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					requestID, _ := r.Context().Value("request_id").(string)
					logger.Error("panic recovered",
						slog.Any("panic", recovered),
						slog.String("request_id", requestID),
					)
					writeJSON(w, http.StatusInternalServerError, errorResponse{
						Error: "internal server error",
						Code:  "INTERNAL_SERVER_ERROR",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}


type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
