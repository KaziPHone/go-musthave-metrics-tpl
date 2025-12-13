package handlers

import (
	"bytes"
	"net/http"

	hlp "github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
)

type shaRw struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (s *shaRw) Write(p []byte) (int, error) {
	return s.buf.Write(p) // ← только записываем в буфер, НЕ прокидываем дальше
}

// WriteHeader тоже перехватываем, чтобы не отправлять заголовки преждевременно
func (s *shaRw) WriteHeader(statusCode int) {
	// Ничего не делаем — откладываем отправку заголовков
}

func ShaMiddleware(key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				h.ServeHTTP(w, r)
				return
			}

			originalBody, ok := r.Context().Value(hlp.OriginalBodyKey).([]byte)
			if !ok {
				http.Error(w, "original_body not found or not []byte in context", http.StatusInternalServerError)
				return
			}

			if hlp.IsBadShaRequest(originalBody, r.Header.Get("HashSHA256")) {
				http.Error(w, "bad request sha", http.StatusBadRequest)
				return
			}

			var buf bytes.Buffer
			crw := &shaRw{ResponseWriter: w, buf: &buf}

			// Выполняем обработчик, но он пишет в буфер, а не сразу в w
			h.ServeHTTP(crw, r)

			// Теперь вычисляем хеш
			digest := hlp.CalcSHA256HashBuffer(buf)

			// Устанавливаем заголовок
			w.Header().Set("HashSHA256", digest)

			// Отправляем статус (если не был отправлен)
			// Если нужно — можно отслеживать statusCode через обёртку
			w.WriteHeader(http.StatusOK) // ← ИЛИ отслеживайте код через кастомный Writer

			// Отправляем тело
			w.Write(buf.Bytes())
		})
	}
}
