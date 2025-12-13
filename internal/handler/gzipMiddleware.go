package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"

	hlp "github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
	"github.com/rs/zerolog/log"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func GzipRequestMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		originalBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read compressed body", http.StatusBadRequest)
			log.Print(err)
			return
		}
		ctx := context.WithValue(r.Context(), hlp.OriginalBodyKey, originalBody)
		r = r.WithContext(ctx)

		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" && contentType != "text/html" {
			http.Error(w, "Unsupported content type with gzip", http.StatusUnsupportedMediaType)
			return
		}

		gzReader, err := gzip.NewReader(bytes.NewReader(originalBody))
		if err != nil {
			http.Error(w, "Ошибка обработки gzipped-запроса", http.StatusBadRequest)
			log.Print(err)
			return
		}
		defer gzReader.Close()

		bodyBytes, err := io.ReadAll(gzReader)
		if err != nil {
			http.Error(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
			log.Print(err)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		h.ServeHTTP(w, r)
	})
}

func GzipResponseMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
