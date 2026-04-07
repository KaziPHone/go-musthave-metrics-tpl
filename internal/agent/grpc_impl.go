//go:build grpc
// +build grpc

package agent

import (
    "context"
    "time"

    pb "github.com/KaziPHone/go-musthave-metrics-tpl/internal/proto"
    models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

// sendViaGRPC отправляет метрики на gRPC сервер, возвращает true если отправлено.
func sendViaGRPC(a *Agent, metrics []models.Metrics) (bool, error) {
    if a.GRPCAddress == "" {
        return false, nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    conn, err := grpc.DialContext(ctx, a.GRPCAddress, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        return false, err
    }
    defer conn.Close()

    client := pb.NewMetricsClient(conn)

    // Формируем pb-метрики
    req := &pb.UpdateMetricsRequest{}
    for _, m := range metrics {
        pm := &pb.Metric{Id: m.ID}
        if m.MType == models.Gauge {
            pm.Type = pb.Metric_GAUGE
            if m.Value != nil {
                pm.Value = *m.Value
            }
        } else {
            pm.Type = pb.Metric_COUNTER
            if m.Delta != nil {
                pm.Delta = *m.Delta
            }
        }
        req.Metrics = append(req.Metrics, pm)
    }

    // Добавляем metadata x-real-ip, если есть
    md := metadata.New(nil)
    if a.localIP != "" {
        md.Set("x-real-ip", a.localIP)
    }
    ctx = metadata.NewOutgoingContext(ctx, md)

    _, err = client.UpdateMetrics(ctx, req)
    if err != nil {
        return false, err
    }
    return true, nil
}
