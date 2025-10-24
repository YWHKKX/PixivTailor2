package grpc

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthServer 健康检查服务器
type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	health *health.Server
}

// NewHealthServer 创建健康检查服务器
func NewHealthServer() *HealthServer {
	healthServer := health.NewServer()
	return &HealthServer{
		health: healthServer,
	}
}

// Check 健康检查
func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return s.health.Check(ctx, req)
}

// Watch 健康检查监控
func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return s.health.Watch(req, stream)
}

// SetServingStatus 设置服务状态
func (s *HealthServer) SetServingStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	s.health.SetServingStatus(service, status)
}

// HTTPHealthHandler HTTP 健康检查处理器
func HTTPHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}
