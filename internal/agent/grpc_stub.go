package agent

import (
    models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

// sendViaGRPC — заглушка для отправки метрик через gRPC.
// По умолчанию (когда пакет собирается без тега 'grpc') ничего не делает
// и сообщает, что отправка не выполнена (false).
func sendViaGRPC(a *Agent, metrics []models.Metrics) (bool, error) {
    return false, nil
}
