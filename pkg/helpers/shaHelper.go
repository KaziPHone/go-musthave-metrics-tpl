// Package helpers предоставляет вспомогательные функции для работы с метриками.
//
// Основные функции:
//   - GetClientIP: получение IP-адреса клиента из заголовков
//   - GetMetrics: извлечение имен метрик из среза для аудита
//   - CalcSHA256Hash, CalcSHA256HashBuffer: вычисление хэша SHA256
//   - IsBadShaRequest: проверка целостности запроса по хэшу
package helpers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"
)

// hasherPool повторно использует экземпляр хэшера для избежания выделения памяти.
var hasherPool = sync.Pool{
	New: func() interface{} {
		return sha256.New()
	},
}

// CalcSHA256HashBuffer вычисляет SHA256 хэш из буфера.
//
// Параметры:
//   - data: буфер с данными
//
// Возвращает:
//   - string: SHA256 хэш в виде шестнадцатеричной строки
func CalcSHA256HashBuffer(data bytes.Buffer) string {
	hasher := hasherPool.Get().(interface {
		Reset()
		io.Writer
		Sum([]byte) []byte
	})
	defer hasherPool.Put(hasher)

	hasher.Reset()
	io.Copy(hasher, &data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// CalcSHA256Hash вычисляет SHA256 хэш из среза байт.
//
// Параметры:
//   - data: данные для хэширования
//
// Возвращает:
//   - string: SHA256 хэш в виде шестнадцатеричной строки
func CalcSHA256Hash(data []byte) string {
	h := hasherPool.Get().(interface {
		Reset()
		io.Writer
		Sum([]byte) []byte
	})
	defer hasherPool.Put(h)

	h.Reset()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// IsBadShaRequest проверяет, что хэш запроса совпадает с вычисленным.
//
// Параметры:
//   - data: исходные данные запроса
//   - hash: ожидаемый хэш из заголовка
//
// Возвращает:
//   - bool: true если хэш не совпадает и не пустой
func IsBadShaRequest(data []byte, hash string) bool {
	return CalcSHA256Hash(data) != hash && hash != ""
}
