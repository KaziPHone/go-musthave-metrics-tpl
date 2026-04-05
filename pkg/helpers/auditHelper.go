// Package helpers предоставляет вспомогательные функции для работы с метриками.
//
// Основные функции:
//   - GetClientIP: получение IP-адреса клиента из заголовков
//   - GetMetrics: извлечение имен метрик из среза для аудита
//   - CalcSHA256Hash, CalcSHA256HashBuffer: вычисление хэша SHA256
//   - IsBadShaRequest: проверка целостности запроса по хэшу
package helpers

import (
	"net/http"
	"strings"
	"sync"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

// metricsBufferPool повторно использует буфер для GetMetrics для избежания выделения памяти.
var metricsBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 10)
	},
}

// GetClientIP получает IP-адрес клиента из HTTP-запроса.
//
// Приоритет проверки заголовков:
//   1. X-Real-IP
//   2. X-Forwarded-For (первый IP в списке)
//   3. RemoteAddr
//
// Обрезает порт из RemoteAddr если присутствует.
//
// Параметры:
//   - r: HTTP запрос
//
// Возвращает:
//   - string: IP-адрес клиента
func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			commaIdx := strings.Index(ip, ",")
			if commaIdx > 0 {
				ip = ip[:commaIdx]
			}
			for len(ip) > 0 && (ip[0] == ' ' || ip[0] == '\t') {
				ip = ip[1:]
			}
			for len(ip) > 0 && (ip[len(ip)-1] == ' ' || ip[len(ip)-1] == '\t') {
				ip = ip[:len(ip)-1]
			}
		}
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}
	return ip
}

// GetMetrics извлекает имена метрик из среза для аудита.
//
// Параметры:
//   - metrics: срез метрик
//
// Возвращает:
//   - []string: имена метрик
//
// Использует пул буферов для избежания выделения памяти.
func GetMetrics(metrics []models.Metrics) []string {
	buf := metricsBufferPool.Get().([]string)
	buf = buf[:0]

	for _, v := range metrics {
		buf = append(buf, v.ID)
	}

	result := make([]string, len(buf))
	copy(result, buf)
	metricsBufferPool.Put(buf)

	return result
}
