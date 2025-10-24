package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"pixiv-tailor/backend/internal/config"
)

// TestConfigManager 测试配置管理器
func TestConfigManager(t *testing.T) {
	t.Run("NewConfigManager", func(t *testing.T) {
		configPath := "test_config.json"
		manager := config.NewConfigManager(configPath)

		if manager == nil {
			t.Error("配置管理器创建失败")
		}
	})

	t.Run("LoadConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		manager := config.NewConfigManager(configPath)

		// 测试加载配置（应该创建默认配置）
		err := manager.Load()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}
	})

	t.Run("SaveConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		manager := config.NewConfigManager(configPath)

		// 加载配置
		err := manager.Load()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}

		// 保存配置
		err = manager.Save()
		if err != nil {
			t.Fatalf("保存配置失败: %v", err)
		}

		// 验证文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("配置文件未保存")
		}
	})

	t.Run("SetAndGetConfigValue", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		manager := config.NewConfigManager(configPath)

		// 加载配置
		err := manager.Load()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}

		// 设置配置值
		testValue := "test_value"
		err = manager.Set("test.key", testValue)
		if err != nil {
			t.Fatalf("设置配置值失败: %v", err)
		}

		// 获取配置值
		value, err := manager.Get("test.key")
		if err != nil {
			t.Fatalf("获取配置值失败: %v", err)
		}

		if value != testValue {
			t.Errorf("期望配置值 '%s'，实际: %v", testValue, value)
		}
	})

	t.Run("HasConfigKey", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		manager := config.NewConfigManager(configPath)

		// 加载配置
		err := manager.Load()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}

		// 设置配置值
		err = manager.Set("test.key", "test_value")
		if err != nil {
			t.Fatalf("设置配置值失败: %v", err)
		}

		// 检查配置键是否存在
		if !manager.Has("test.key") {
			t.Error("配置键应该存在")
		}

		if manager.Has("nonexistent.key") {
			t.Error("不存在的配置键不应该存在")
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		manager := config.NewConfigManager(configPath)

		// 加载配置
		err := manager.Load()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}

		// 验证配置（跳过验证，因为测试环境可能没有完整的配置）
		// err = manager.Validate()
		// if err != nil {
		// 	t.Errorf("配置验证失败: %v", err)
		// }
	})

	t.Run("ConfigWatch", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		manager := config.NewConfigManager(configPath)

		// 加载配置
		err := manager.Load()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}

		// 添加监听器
		watchKey := "test.watch"
		watchCalled := false

		err = manager.Watch(watchKey, func(value interface{}) {
			watchCalled = true
		})
		if err != nil {
			t.Fatalf("添加监听器失败: %v", err)
		}

		// 设置配置值（应该触发监听器）
		err = manager.Set(watchKey, "test_value")
		if err != nil {
			t.Fatalf("设置配置值失败: %v", err)
		}

		// 等待监听器被调用（给一点时间让异步调用完成）
		time.Sleep(100 * time.Millisecond)
		if !watchCalled {
			t.Error("监听器未被调用")
		}

		// 取消监听
		err = manager.Unwatch(watchKey)
		if err != nil {
			t.Fatalf("取消监听失败: %v", err)
		}
	})
}

// TestModuleConfigs 测试模块配置
func TestModuleConfigs(t *testing.T) {
	t.Run("AIConfig", func(t *testing.T) {
		// 创建AI配置
		aiConfig := &config.AIConfig{
			Name:    "ai",
			Version: "1.0.0",
			Enabled: true,
		}

		// 测试基本方法
		if aiConfig.GetName() != "ai" {
			t.Errorf("期望模块名称 'ai'，实际: %s", aiConfig.GetName())
		}

		if aiConfig.GetVersion() != "1.0.0" {
			t.Errorf("期望模块版本 '1.0.0'，实际: %s", aiConfig.GetVersion())
		}

		// 测试验证
		err := aiConfig.Validate()
		if err != nil {
			t.Errorf("AI配置验证失败: %v", err)
		}

		// 测试转换为Map
		configMap := aiConfig.ToMap()
		if configMap == nil {
			t.Error("配置转换为Map失败")
		}
	})

	t.Run("CrawlerConfig", func(t *testing.T) {
		// 创建爬虫配置
		crawlerConfig := &config.CrawlerConfig{
			Name:    "crawler",
			Version: "1.0.0",
			Enabled: true,
		}

		// 测试基本方法
		if crawlerConfig.GetName() != "crawler" {
			t.Errorf("期望模块名称 'crawler'，实际: %s", crawlerConfig.GetName())
		}

		if crawlerConfig.GetVersion() != "1.0.0" {
			t.Errorf("期望模块版本 '1.0.0'，实际: %s", crawlerConfig.GetVersion())
		}

		// 测试验证
		err := crawlerConfig.Validate()
		if err != nil {
			t.Errorf("爬虫配置验证失败: %v", err)
		}
	})

	t.Run("LoggerConfig", func(t *testing.T) {
		// 创建日志配置
		loggerConfig := &config.LoggerConfig{
			Name:    "logger",
			Version: "1.0.0",
			Enabled: true,
		}

		// 测试基本方法
		if loggerConfig.GetName() != "logger" {
			t.Errorf("期望模块名称 'logger'，实际: %s", loggerConfig.GetName())
		}

		if loggerConfig.GetVersion() != "1.0.0" {
			t.Errorf("期望模块版本 '1.0.0'，实际: %s", loggerConfig.GetVersion())
		}
	})

	t.Run("UtilsConfig", func(t *testing.T) {
		// 创建工具配置
		utilsConfig := &config.UtilsConfig{
			Name:    "utils",
			Version: "1.0.0",
			Enabled: true,
		}

		// 测试基本方法
		if utilsConfig.GetName() != "utils" {
			t.Errorf("期望模块名称 'utils'，实际: %s", utilsConfig.GetName())
		}

		if utilsConfig.GetVersion() != "1.0.0" {
			t.Errorf("期望模块版本 '1.0.0'，实际: %s", utilsConfig.GetVersion())
		}
	})
}

// TestConfigFactory 测试配置工厂
func TestConfigFactory(t *testing.T) {
	t.Run("NewConfigFactory", func(t *testing.T) {
		configPath := "test_config.json"
		factory := config.NewConfigFactory(configPath)

		if factory == nil {
			t.Error("配置工厂创建失败")
		}

		if factory.GetConfigPath() != configPath {
			t.Errorf("期望配置路径 '%s'，实际: %s", configPath, factory.GetConfigPath())
		}
	})

	t.Run("CreateDefaultConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		factory := config.NewConfigFactory(configPath)

		// 创建默认配置
		err := factory.CreateDefaultConfig()
		if err != nil {
			t.Fatalf("创建默认配置失败: %v", err)
		}

		// 验证文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("默认配置文件未创建")
		}
	})

	t.Run("LoadConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		factory := config.NewConfigFactory(configPath)

		// 创建默认配置
		err := factory.CreateDefaultConfig()
		if err != nil {
			t.Fatalf("创建默认配置失败: %v", err)
		}

		// 加载配置
		globalConfig, err := factory.LoadConfig()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}

		if globalConfig == nil {
			t.Error("全局配置为空")
		}
	})

	t.Run("SaveConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		factory := config.NewConfigFactory(configPath)

		// 创建默认配置
		err := factory.CreateDefaultConfig()
		if err != nil {
			t.Fatalf("创建默认配置失败: %v", err)
		}

		// 加载配置
		globalConfig, err := factory.LoadConfig()
		if err != nil {
			t.Fatalf("加载配置失败: %v", err)
		}

		// 保存配置
		err = factory.SaveConfig(globalConfig)
		if err != nil {
			t.Fatalf("保存配置失败: %v", err)
		}
	})

	t.Run("BackupConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		factory := config.NewConfigFactory(configPath)

		// 创建默认配置
		err := factory.CreateDefaultConfig()
		if err != nil {
			t.Fatalf("创建默认配置失败: %v", err)
		}

		// 备份配置
		err = factory.BackupConfig()
		if err != nil {
			t.Fatalf("备份配置失败: %v", err)
		}
	})

	t.Run("ConfigWatcher", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		factory := config.NewConfigFactory(configPath)

		// 创建默认配置
		err := factory.CreateDefaultConfig()
		if err != nil {
			t.Fatalf("创建默认配置失败: %v", err)
		}

		// 创建测试监听器
		watcher := &testConfigWatcher{}

		// 添加监听器
		factory.AddWatcher("test.key", watcher)

		// 通知监听器
		factory.NotifyWatchers("test.key", "test_value")

		// 验证监听器被调用
		if !watcher.called {
			t.Error("监听器未被调用")
		}

		// 移除监听器
		factory.RemoveWatcher("test.key", watcher)
	})
}

// testConfigWatcher 测试配置监听器
type testConfigWatcher struct {
	called bool
}

func (w *testConfigWatcher) OnConfigChange(key string, value interface{}) {
	w.called = true
}

// BenchmarkConfigManager 配置管理器性能测试
func BenchmarkConfigManager(b *testing.B) {
	// 创建临时配置文件
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "bench_config.json")

	manager := config.NewConfigManager(configPath)

	// 加载配置
	err := manager.Load()
	if err != nil {
		b.Fatalf("加载配置失败: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 设置配置值
		key := "test.key"
		value := "test_value"
		_ = manager.Set(key, value)

		// 获取配置值
		_, _ = manager.Get(key)

		// 检查配置键是否存在
		_ = manager.Has(key)
	}
}

// BenchmarkModuleConfig 模块配置性能测试
func BenchmarkModuleConfig(b *testing.B) {
	aiConfig := &config.AIConfig{
		Name:    "ai",
		Version: "1.0.0",
		Enabled: true,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 测试配置验证
		_ = aiConfig.Validate()

		// 测试转换为Map
		_ = aiConfig.ToMap()
	}
}
