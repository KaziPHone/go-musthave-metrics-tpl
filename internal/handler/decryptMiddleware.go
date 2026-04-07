package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"

	cryptoPkg "github.com/KaziPHone/go-musthave-metrics-tpl/pkg/crypto"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
	"crypto/rsa"
	"github.com/rs/zerolog/log"
)

// DecryptRequestMiddleware расшифровывает входящее тело запроса с помощью переданного приватного ключа.
// Должен регистрироваться до GzipRequestMiddleware, чтобы далее распаковывающие/проверяющие
// middleware получили в контексте и в r.Body уже расшифрованные (gzip) байты.
func DecryptRequestMiddleware(priv *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if priv == nil {
				h.ServeHTTP(w, r)
				return
			}

			encBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				log.Error().Err(err).Msg("read request body error")
				return
			}

			if len(encBody) == 0 {
				r.Body = io.NopCloser(bytes.NewBuffer(encBody))
				ctx := context.WithValue(r.Context(), helpers.OriginalBodyKey, encBody)
				r = r.WithContext(ctx)
				h.ServeHTTP(w, r)
				return
			}

			plain, err := cryptoPkg.Decrypt(priv, encBody)
			if err != nil {
				http.Error(w, "failed to decrypt body", http.StatusBadRequest)
				log.Error().Err(err).Msg("decrypt request body error")
				return
			}

			// Кладём расшифрованные байты (должны быть gzip-содержимым) в контекст и в r.Body
			// для дальнейших middleware и обработчиков.
			ctx := context.WithValue(r.Context(), helpers.OriginalBodyKey, plain)
			r = r.WithContext(ctx)
			r.Body = io.NopCloser(bytes.NewBuffer(plain))

			h.ServeHTTP(w, r)
		})
	}
}

