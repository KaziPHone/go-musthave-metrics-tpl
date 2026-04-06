// Package models предоставляет модель данных для метрик.
//
// Модель поддерживает два типа метрик:
//   - Gauge: плавающая точка (float64)
//   - Counter: целое число (int64)
//
// Delta и Value объявлены через указатели, чтобы отличать значение "0"
// от не заданного значения (nil), что важно для корректной сериализации JSON.
package models

const (
	// Counter тип метрики-счетчика (целое число).
	Counter = "counter"
	// Gauge тип метрики-датчика (плавающая точка).
	Gauge = "gauge"
)

// Metrics представляет собой метрику с идентификатором, типом и значением.
//
// Метрика может быть типа Gauge (с полем Value) или Counter (с полем Delta).
// Использование указателей для Delta и Value позволяет отличать значение "0"
// от не заданного значения (nil), что важно для корректной сериализации JSON.
//
// Примеры:
//
//	// Gauge метрика
//	gauge := Metrics{
//	    ID:    "response_time",
//	    MType: "gauge",
//	    Value: &[]float64{0.123}[0],
//	}
//
//	// Counter метрика
//	counter := Metrics{
//	    ID:    "request_count",
//	    MType: "counter",
//	    Delta: &[]int64{100}[0],
//	}
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
