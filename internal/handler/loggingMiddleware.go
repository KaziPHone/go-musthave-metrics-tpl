package handlers

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type responseRecorder struct {
	http.ResponseWriter     // Исходный ResponseWriter
	statusCode          int // Код состояния
	responseSize        int // Размер ответа
	bodyBuf             bytes.Buffer
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	n, err := rr.ResponseWriter.Write(b)
	rr.responseSize += n
	rr.bodyBuf.Write(b)
	return n, err
}

func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Читаем тело запроса
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close() // закрываем оригинальный поток

		// Восстанавливаем тело запроса обратно в поток
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		h.ServeHTTP(recorder, r)

		log.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Int("status_code", recorder.statusCode).
			Int("response_size_bytes", recorder.responseSize).
			Dur("request_duration_ms", time.Since(start)).
			Str("req_body", decompressed(bodyBytes, r)).
			Str("resp_body", decompressed(recorder.bodyBuf.Bytes(), r)).
			//Str("body", string(bodyBytes)).
			//Str("resp_body", recorder.bodyBuf.String()).
			Msg("")
	})
}

func decompressed(value []byte, r *http.Request) string {
	if r.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(bytes.NewBuffer(value))
		if err != nil {
			log.Error().Err(err).Msg("Ошибка декомпрессии gzip")
		}
		defer reader.Close()

		decompressedBody, err := io.ReadAll(reader)
		if err != nil {
			log.Error().Err(err).Msg("Ошибка чтения декомпрессированных данных")
		}
		return string(decompressedBody)
	} else {
		return string(value)
	}
}
