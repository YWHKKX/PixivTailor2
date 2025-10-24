package service

import (
	"runtime"
	"time"
)

// SystemService 系统服务接口
type SystemService interface {
	GetSystemInfo() (*SystemInfo, error)
	GetHealthStatus() (*HealthStatus, error)
	GetMetrics() (*SystemMetrics, error)
}

// SystemInfo 系统信息
type SystemInfo struct {
	Version      string    `json:"version"`
	BuildTime    time.Time `json:"build_time"`
	GoVersion    string    `json:"go_version"`
	Platform     string    `json:"platform"`
	Architecture string    `json:"architecture"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	Memory     *MemoryMetrics `json:"memory"`
	CPU        *CPUMetrics    `json:"cpu"`
	Goroutines int            `json:"goroutines"`
	Uptime     time.Duration  `json:"uptime"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	NumCPU       int `json:"num_cpu"`
	NumGoroutine int `json:"num_goroutine"`
}

// systemServiceImpl 系统服务实现
type systemServiceImpl struct {
	startTime time.Time
}

// NewSystemService 创建系统服务实例
func NewSystemService() SystemService {
	return &systemServiceImpl{
		startTime: time.Now(),
	}
}

// GetSystemInfo 获取系统信息
func (s *systemServiceImpl) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		Version:      "1.0.0",
		BuildTime:    time.Now(),
		GoVersion:    runtime.Version(),
		Platform:     runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	return info, nil
}

// GetHealthStatus 获取健康状态
func (s *systemServiceImpl) GetHealthStatus() (*HealthStatus, error) {
	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services: map[string]string{
			"database": "healthy",
			"grpc":     "healthy",
			"storage":  "healthy",
		},
	}

	return status, nil
}

// GetMetrics 获取系统指标
func (s *systemServiceImpl) GetMetrics() (*SystemMetrics, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := &SystemMetrics{
		Memory: &MemoryMetrics{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
		CPU: &CPUMetrics{
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		Goroutines: runtime.NumGoroutine(),
		Uptime:     time.Since(s.startTime),
	}

	return metrics, nil
}
