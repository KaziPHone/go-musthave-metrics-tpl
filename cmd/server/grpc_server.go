//go:build grpc
// +build grpc

package main

import (
	"context"
	"net"
	"strings"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	pb "github.com/KaziPHone/go-musthave-metrics-tpl/internal/proto"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type metricsServer struct {
	pb.UnimplementedMetricsServer
	st      storage.IStorage
	trusted *net.IPNet
}

func (s *metricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	ips := md.Get("x-real-ip")
	var ipStr string
	if len(ips) > 0 {
		ipStr = strings.TrimSpace(ips[0])
	}
	if ipStr == "" {
		return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, status.Error(codes.PermissionDenied, "invalid ip")
	}
	if s.trusted != nil && !s.trusted.Contains(ip) {
		return nil, status.Error(codes.PermissionDenied, "ip not in trusted subnet")
	}

	var ms []model.Metrics
	for _, m := range req.GetMetrics() {
		var mv *float64
		var dv *int64
		if m.GetValue() != 0 {
			v := m.GetValue()
			mv = &v
		}
		if m.GetDelta() != 0 {
			d := m.GetDelta()
			dv = &d
		}
		mt := model.MetricType(m.GetType().String())
		mm := model.Metrics{ID: m.GetId(), MType: mt}
		if mv != nil {
			mm.Value = mv
		}
		if dv != nil {
			mm.Delta = dv
		}
		ms = append(ms, mm)
	}
	s.st.UpdatesMetrics(ms)

	return &pb.UpdateMetricsResponse{}, nil
}

func unaryAuthInterceptor(trusted *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trusted == nil {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		ips := md.Get("x-real-ip")
		if len(ips) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}
		ip := net.ParseIP(strings.TrimSpace(ips[0]))
		if ip == nil || !trusted.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "ip not in trusted subnet")
		}
		return handler(ctx, req)
	}
}

func StartGRPCServer(addr string, st storage.IStorage, trusted *net.IPNet) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(unaryAuthInterceptor(trusted)))
	pb.RegisterMetricsServer(srv, &metricsServer{st: st, trusted: trusted})
	go srv.Serve(lis)
	return srv, lis, nil
}
