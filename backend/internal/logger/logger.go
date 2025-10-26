package logger

import (
	"os"
	"path/filepath"

	"pixiv-tailor/backend/pkg/paths"

	"github.com/sirupsen/logrus"
)

// LogLevel 日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var (
	logger *logrus.Logger
	level  LogLevel = InfoLevel
)

// ensureLogger 确保 logger 已初始化
func ensureLogger() {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   false,
		})
		logger.SetOutput(os.Stdout)
	}
}

// Init 初始化日志系统
func Init(verbose bool) {
	logger = logrus.New()

	// 设置日志级别
	if verbose {
		logger.SetLevel(logrus.DebugLevel)
		level = DebugLevel
	} else {
		logger.SetLevel(logrus.InfoLevel)
		level = InfoLevel
	}

	// 设置日志格式
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})

	// 设置输出
	logger.SetOutput(os.Stdout)
}

// SetLevel 设置日志级别
func SetLevel(l LogLevel) {
	level = l
	switch l {
	case DebugLevel:
		logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		logger.SetLevel(logrus.FatalLevel)
	}
}

// SetOutput 设置日志输出
func SetOutput(output string) {
	ensureLogger()

	switch output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		// 使用路径管理器获取日志目录
		pathManager := paths.GetPathManager()
		var logFile string

		if pathManager != nil {
			logFile = pathManager.GetMainLogPath()
		} else {
			// 备用方案
			logDir := "logs"
			if err := os.MkdirAll(logDir, 0755); err != nil {
				logger.Errorf("创建日志目录失败: %v", err)
				return
			}
			logFile = filepath.Join(logDir, "pixiv-tailor.log")
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Errorf("打开日志文件失败: %v", err)
			return
		}

		logger.SetOutput(file)
	default:
		logger.SetOutput(os.Stdout)
	}
}

// Debug 调试日志
func Debug(args ...interface{}) {
	ensureLogger()
	if level <= DebugLevel {
		logger.Debug(args...)
	}
}

// Debugf 格式化调试日志
func Debugf(format string, args ...interface{}) {
	ensureLogger()
	if level <= DebugLevel {
		logger.Debugf(format, args...)
	}
}

// Info 信息日志
func Info(args ...interface{}) {
	ensureLogger()
	if level <= InfoLevel {
		logger.Info(args...)
	}
}

// Infof 格式化信息日志
func Infof(format string, args ...interface{}) {
	ensureLogger()
	if level <= InfoLevel {
		logger.Infof(format, args...)
	}
}

// Warn 警告日志
func Warn(args ...interface{}) {
	ensureLogger()
	if level <= WarnLevel {
		logger.Warn(args...)
	}
}

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) {
	ensureLogger()
	if level <= WarnLevel {
		logger.Warnf(format, args...)
	}
}

// Error 错误日志
func Error(args ...interface{}) {
	ensureLogger()
	if level <= ErrorLevel {
		logger.Error(args...)
	}
}

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) {
	ensureLogger()
	if level <= ErrorLevel {
		logger.Errorf(format, args...)
	}
}

// Fatal 致命错误日志
func Fatal(args ...interface{}) {
	ensureLogger()
	if level <= FatalLevel {
		logger.Fatal(args...)
	}
}

// Fatalf 格式化致命错误日志
func Fatalf(format string, args ...interface{}) {
	ensureLogger()
	if level <= FatalLevel {
		logger.Fatalf(format, args...)
	}
}

// WithField 添加字段
func WithField(key string, value interface{}) *logrus.Entry {
	ensureLogger()
	return logger.WithField(key, value)
}

// WithFields 添加多个字段
func WithFields(fields logrus.Fields) *logrus.Entry {
	ensureLogger()
	return logger.WithFields(fields)
}

// GetLogger 获取原始 logger 实例
func GetLogger() *logrus.Logger {
	ensureLogger()
	return logger
}
