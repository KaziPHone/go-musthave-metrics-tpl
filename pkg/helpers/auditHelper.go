package helpers

import (
	"net/http"
	"strings"
	"sync"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

// повторное использование буфера для GetMetrics, чтобы избежать выделения памяти
var metricsBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 10)
	},
}

func GetClientIP(r *http.Request) string {
	// Получаем IP без выделения памяти - используем strings.Index вместо strings.Split
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// Находим первую запятую для извлечения первого IP
			commaIdx := strings.Index(ip, ",")
			if commaIdx > 0 {
				ip = ip[:commaIdx]
			}
			// Обрезаем пробелы на месте
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
	// Разделяем по последнему двоеточию для порта
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}
	return ip
}

func GetMetrics(metrics []models.Metrics) []string {
	buf := metricsBufferPool.Get().([]string)
	buf = buf[:0] // Очищаем буфер

	for _, v := range metrics {
		buf = append(buf, v.ID)
	}

	// Возвращаем копию, чтобы вызывающий не мог изменить буфер пула
	result := make([]string, len(buf))
	copy(result, buf)
	metricsBufferPool.Put(buf)

	return result
}
