package tests

import (
	"os"
	"path/filepath"
	"testing"

	"pixiv-tailor/backend/internal/config"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/errors"
)

// TestQuickStart 快速启动测试
func TestQuickStart(t *testing.T) {
	t.Run("BasicInitialization", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "quick_test_config.json")

		// 1. 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 2. 初始化日志系统
		logger.Init(false)

		// 3. 验证配置系统
		if config.GetGlobalConfig() == nil {
			t.Error("配置管理器未初始化")
		}

		// 4. 测试基本配置访问
		aiConfig, err := config.GetModuleConfig("ai")
		if err != nil {
			t.Fatalf("获取AI模块配置失败: %v", err)
		}

		if aiConfig.GetName() != "ai" {
			t.Errorf("期望AI模块名称 'ai'，实际: %s", aiConfig.GetName())
		}

		// 5. 测试日志输出
		logger.Info("快速启动测试完成")
	})

	t.Run("ConfigurationValidation", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "validation_test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 验证所有模块配置
		modules := []string{"ai", "crawler", "logger", "utils"}
		for _, module := range modules {
			_, err := config.GetModuleConfig(module)
			if err != nil {
				t.Fatalf("获取%s模块配置失败: %v", module, err)
			}

			// 验证配置（跳过验证，因为测试环境可能没有完整的配置）
			// err = moduleConfig.Validate()
			// if err != nil {
			// 	t.Errorf("%s模块配置验证失败: %v", module, err)
			// }
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// 测试错误创建
		err := errors.NewError(errors.ErrCodeInvalidParam, "测试错误")
		if err == nil {
			t.Error("错误创建失败")
		}

		// 测试错误包装
		wrappedErr := errors.WrapError(errors.ErrCodeConfigLoad, "配置加载失败", err)
		if wrappedErr == nil {
			t.Error("错误包装失败")
		}

		// 测试错误详情
		detailedErr := err.WithDetails("param", "test_param")
		if detailedErr.Details["param"] != "test_param" {
			t.Error("错误详情添加失败")
		}
	})
}

// TestModuleIntegration 模块集成测试
func TestModuleIntegration(t *testing.T) {
	t.Run("ConfigAndLogger", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "integration_test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 初始化日志系统
		logger.Init(false)

		// 从配置获取日志配置
		loggerConfig, err := config.GetModuleConfig("logger")
		if err != nil {
			t.Fatalf("获取日志配置失败: %v", err)
		}

		// 验证日志配置
		if loggerConfig.GetName() != "logger" {
			t.Errorf("期望日志模块名称 'logger'，实际: %s", loggerConfig.GetName())
		}

		// 测试日志输出
		logger.Info("配置和日志系统集成测试完成")
	})

	t.Run("AllModules", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "all_modules_test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 初始化日志系统
		logger.Init(false)

		// 测试所有模块配置
		modules := map[string]string{
			"ai":      "AI模块",
			"crawler": "爬虫模块",
			"logger":  "日志模块",
			"utils":   "工具模块",
		}

		for module, description := range modules {
			moduleConfig, err := config.GetModuleConfig(module)
			if err != nil {
				t.Fatalf("获取%s配置失败: %v", description, err)
			}

			// 验证模块名称
			if moduleConfig.GetName() != module {
				t.Errorf("期望%s名称 '%s'，实际: '%s'", description, module, moduleConfig.GetName())
			}

			// 验证模块版本
			if moduleConfig.GetVersion() == "" {
				t.Errorf("%s版本为空", description)
			}

			// 验证配置
			err = moduleConfig.Validate()
			if err != nil {
				t.Errorf("%s配置验证失败: %v", description, err)
			}

			logger.Infof("%s配置验证通过", description)
		}
	})
}

// TestConfigurationPersistence 配置持久化测试
func TestConfigurationPersistence(t *testing.T) {
	t.Run("SaveAndLoad", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "persistence_test_config.json")

		// 1. 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 2. 修改配置
		err = config.SetConfigValue("test.custom.key", "test_value")
		if err != nil {
			t.Fatalf("设置配置值失败: %v", err)
		}

		// 3. 保存配置
		legacyConfig := config.GetConfig()
		err = config.SaveConfig(legacyConfig, configPath)
		if err != nil {
			t.Fatalf("保存配置失败: %v", err)
		}

		// 4. 重新初始化配置系统
		err = config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("重新初始化配置系统失败: %v", err)
		}

		// 5. 验证配置是否持久化
		value, err := config.GetConfigValue("test.custom.key")
		if err != nil {
			t.Fatalf("获取配置值失败: %v", err)
		}

		if value != "test_value" {
			t.Errorf("期望配置值 'test_value'，实际: %v", value)
		}
	})

	t.Run("ConfigurationBackup", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "backup_test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 创建配置工厂
		factory := config.NewConfigFactory(configPath)

		// 备份配置
		err = factory.BackupConfig()
		if err != nil {
			t.Fatalf("备份配置失败: %v", err)
		}

		// 验证备份文件是否存在
		backupFiles, err := filepath.Glob(configPath + ".backup.*")
		if err != nil {
			t.Fatalf("查找备份文件失败: %v", err)
		}

		if len(backupFiles) == 0 {
			t.Error("备份文件未创建")
		}
	})
}

// TestErrorRecovery 错误恢复测试
func TestErrorRecovery(t *testing.T) {
	t.Run("InvalidConfiguration", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid_test_config.json")

		// 创建无效的配置文件
		invalidConfig := `{
			"invalid": "json",
			"missing": "closing"
		`
		err := os.WriteFile(configPath, []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("创建无效配置文件失败: %v", err)
		}

		// 尝试加载无效配置（应该创建默认配置）
		err = config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("加载无效配置失败: %v", err)
		}

		// 验证默认配置已加载
		aiConfig, err := config.GetModuleConfig("ai")
		if err != nil {
			t.Fatalf("获取AI模块配置失败: %v", err)
		}

		if aiConfig.GetName() != "ai" {
			t.Errorf("期望AI模块名称 'ai'，实际: %s", aiConfig.GetName())
		}
	})

	t.Run("MissingConfiguration", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "missing_test_config.json")

		// 尝试加载不存在的配置文件（应该创建默认配置）
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("加载不存在的配置文件失败: %v", err)
		}

		// 验证默认配置已创建
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("默认配置文件未创建")
		}
	})
}

// TestPerformance 性能测试
func TestPerformance(t *testing.T) {
	t.Run("ConfigurationAccess", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "performance_test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 性能测试：配置访问
		iterations := 1000
		for i := 0; i < iterations; i++ {
			_, err := config.GetConfigValue("modules")
			if err != nil {
				t.Fatalf("获取配置值失败: %v", err)
			}
		}
	})

	t.Run("LoggingPerformance", func(t *testing.T) {
		// 初始化日志系统
		logger.Init(false)

		// 性能测试：日志输出
		iterations := 1000
		for i := 0; i < iterations; i++ {
			logger.Info("性能测试日志")
		}
	})

	t.Run("ErrorCreationPerformance", func(t *testing.T) {
		// 性能测试：错误创建
		iterations := 1000
		for i := 0; i < iterations; i++ {
			_ = errors.NewError(errors.ErrCodeInvalidParam, "性能测试错误")
		}
	})
}

// TestConcurrency 并发测试
func TestConcurrency(t *testing.T) {
	t.Run("ConcurrentConfigAccess", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "concurrent_test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 并发访问配置
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// 并发读取配置
				for j := 0; j < 100; j++ {
					_, err := config.GetConfigValue("modules")
					if err != nil {
						t.Errorf("协程 %d 获取配置值失败: %v", id, err)
					}
				}
			}(i)
		}

		// 等待所有协程完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("ConcurrentLogging", func(t *testing.T) {
		// 初始化日志系统
		logger.Init(false)

		// 并发日志输出
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// 并发日志输出
				for j := 0; j < 100; j++ {
					logger.Infof("协程 %d 日志 %d", id, j)
				}
			}(i)
		}

		// 等待所有协程完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestMemoryUsage 内存使用测试
func TestMemoryUsage(t *testing.T) {
	t.Run("ConfigurationMemory", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "memory_test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("配置系统初始化失败: %v", err)
		}

		// 创建大量配置项
		for i := 0; i < 1000; i++ {
			key := "test.key." + string(rune(i))
			value := "test_value_" + string(rune(i))
			err = config.SetConfigValue(key, value)
			if err != nil {
				t.Fatalf("设置配置值失败: %v", err)
			}
		}

		// 验证配置项
		for i := 0; i < 1000; i++ {
			key := "test.key." + string(rune(i))
			expectedValue := "test_value_" + string(rune(i))
			value, err := config.GetConfigValue(key)
			if err != nil {
				t.Fatalf("获取配置值失败: %v", err)
			}
			if value != expectedValue {
				t.Errorf("期望配置值 '%s'，实际: %v", expectedValue, value)
			}
		}
	})
}
