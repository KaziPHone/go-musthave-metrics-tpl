package helpers

import (
	"net/http"
	"strings"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// Берём первый IP, если несколько
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		}
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return strings.Split(ip, ":")[0]
}

func GetMetrics(metrics []models.Metrics) []string {
	modelsMetrics := []string{}
	for _, v := range metrics {
		modelsMetrics = append(modelsMetrics, v.ID)
	}
	return modelsMetrics
}
