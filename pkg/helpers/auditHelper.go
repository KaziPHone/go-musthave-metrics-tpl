package helpers

import (
	"net/http"
	"strings"
	"sync"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

// reuse buffer for GetMetrics to avoid allocations
var metricsBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 10)
	},
}

func GetClientIP(r *http.Request) string {
	// Get IP without allocation - use strings.Index instead of strings.Split
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// Find first comma to extract first IP
			commaIdx := strings.Index(ip, ",")
			if commaIdx > 0 {
				ip = ip[:commaIdx]
			}
			// Trim spaces in-place
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
	// Split at last colon for port
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}
	return ip
}

func GetMetrics(metrics []models.Metrics) []string {
	buf := metricsBufferPool.Get().([]string)
	buf = buf[:0] // Clear the buffer
	
	for _, v := range metrics {
		buf = append(buf, v.ID)
	}
	
	// Return a copy to avoid the caller modifying the pooled buffer
	result := make([]string, len(buf))
	copy(result, buf)
	metricsBufferPool.Put(buf)
	
	return result
}
