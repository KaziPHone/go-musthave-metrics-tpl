package handlers

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type responseRecorder struct {
	http.ResponseWriter     // Исходный ResponseWriter
	statusCode          int // Код состояния
	responseSize        int // Размер ответа
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	n, err := rr.ResponseWriter.Write(b)
	rr.responseSize += n
	return n, err
}

func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusBadRequest,
		}

		h.ServeHTTP(recorder, r)

		log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Int("status_code", recorder.statusCode).
			Int("response_size_bytes", recorder.responseSize).
			Dur("request_duration_ms", time.Since(start)).
			Msg("")
	})
}
