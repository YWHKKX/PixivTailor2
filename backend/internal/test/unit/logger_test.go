package tests

import (
	"testing"

	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/errors"
)

// TestPredefinedErrors 测试预定义错误
func TestPredefinedErrors(t *testing.T) {
	t.Run("ErrInvalidParam", func(t *testing.T) {
		err := errors.ErrInvalidParam

		if err.Code != errors.ErrCodeInvalidParam {
			t.Errorf("期望错误代码 %d，实际: %d", errors.ErrCodeInvalidParam, err.Code)
		}

		if err.Message != "无效参数" {
			t.Errorf("期望错误消息 '无效参数'，实际: %s", err.Message)
		}
	})

	t.Run("ErrFileNotFound", func(t *testing.T) {
		err := errors.ErrFileNotFound

		if err.Code != errors.ErrCodeFileNotFound {
			t.Errorf("期望错误代码 %d，实际: %d", errors.ErrCodeFileNotFound, err.Code)
		}

		if err.Message != "文件不存在" {
			t.Errorf("期望错误消息 '文件不存在'，实际: %s", err.Message)
		}
	})

	t.Run("ErrNetworkError", func(t *testing.T) {
		err := errors.ErrNetworkError

		if err.Code != errors.ErrCodeNetworkError {
			t.Errorf("期望错误代码 %d，实际: %d", errors.ErrCodeNetworkError, err.Code)
		}

		if err.Message != "网络错误" {
			t.Errorf("期望错误消息 '网络错误'，实际: %s", err.Message)
		}
	})

	t.Run("ErrTimeout", func(t *testing.T) {
		err := errors.ErrTimeout

		if err.Code != errors.ErrCodeTimeout {
			t.Errorf("期望错误代码 %d，实际: %d", errors.ErrCodeTimeout, err.Code)
		}

		if err.Message != "请求超时" {
			t.Errorf("期望错误消息 '请求超时'，实际: %s", err.Message)
		}
	})
}

// BenchmarkErrorWrapping 错误包装性能测试
func BenchmarkErrorWrapping(b *testing.B) {
	originalErr := errors.NewError(errors.ErrCodeFileNotFound, "文件不存在")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = errors.WrapError(errors.ErrCodeConfigLoad, "配置加载失败", originalErr)
	}
}

// BenchmarkStructuredLogging 结构化日志性能测试
func BenchmarkStructuredLogging(b *testing.B) {
	// 初始化日志系统
	logger.Init(false)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithField("user_id", i).WithField("operation", "test").Info("性能测试日志")
	}
}
