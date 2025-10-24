package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/errors"
)

// GlobalConfig 全局配置
type GlobalConfig struct {
	Version   string                  `json:"version"`
	Modules   map[string]ModuleConfig `json:"modules"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}

// ModuleConfig 模块配置接口
type ModuleConfig interface {
	Validate() error
	GetName() string
	GetVersion() string
}

// ConfigManager 配置管理器
type ConfigManager struct {
	configPath string
	config     *GlobalConfig
	mu         sync.RWMutex
}

var (
	globalConfig *ConfigManager
	once         sync.Once
)

// InitGlobalConfig 初始化全局配置
func InitGlobalConfig(configPath string) error {
	var err error
	once.Do(func() {
		globalConfig = &ConfigManager{
			configPath: configPath,
		}
		err = globalConfig.Load()
	})
	return err
}

// GetGlobalConfig 获取全局配置管理器
func GetGlobalConfig() *ConfigManager {
	return globalConfig
}

// Load 加载配置
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查配置文件是否存在
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// 创建默认配置
		logger.Infof("配置文件不存在，创建默认配置: %s", cm.configPath)
		return cm.createDefaultConfigUnlocked()
	}

	// 读取配置文件
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return errors.NewError(errors.ErrCodeConfigLoad, fmt.Sprintf("读取配置文件失败: %v", err))
	}

	// 解析配置
	var config GlobalConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return errors.NewError(errors.ErrCodeConfigParse, fmt.Sprintf("解析配置文件失败: %v", err))
	}

	cm.config = &config
	return nil
}

// Save 保存配置
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.saveUnlocked()
}

// GetModuleConfig 获取模块配置
func (cm *ConfigManager) GetModuleConfig(moduleName string) (ModuleConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return nil, errors.NewError(errors.ErrCodeConfigLoad, "配置未初始化")
	}

	moduleData, exists := cm.config.Modules[moduleName]
	if !exists {
		return nil, errors.NewError(errors.ErrCodeConfigLoad, fmt.Sprintf("模块配置不存在: %s", moduleName))
	}

	return moduleData, nil
}

// SetModuleConfig 设置模块配置
func (cm *ConfigManager) SetModuleConfig(moduleName string, config ModuleConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		cm.config = &GlobalConfig{
			Version: "1.0.0",
			Modules: make(map[string]ModuleConfig),
		}
	}

	cm.config.Modules[moduleName] = config
	cm.config.UpdatedAt = time.Now()

	return cm.saveUnlocked()
}

// createDefaultConfigUnlocked 创建默认配置（不加锁版本）
func (cm *ConfigManager) createDefaultConfigUnlocked() error {
	cm.config = &GlobalConfig{
		Version:   "1.0.0",
		Modules:   make(map[string]ModuleConfig),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 创建默认模块配置
	cm.config.Modules["ai"] = &AIConfig{
		Name:    "ai",
		Version: "1.0.0",
		SDWebUI: SDWebUIConfig{
			URL:     "http://127.0.0.1:7860",
			APIKey:  "",
			Timeout: 60,
		},
		KohyaSS: KohyaSSConfig{
			URL:     "http://127.0.0.1:7861",
			Timeout: 60,
		},
		OpenAI: OpenAIConfig{
			APIKeys: []string{},
			Timeout: 30,
		},
	}

	cm.config.Modules["crawler"] = &CrawlerConfig{
		Name:    "crawler",
		Version: "1.0.0",
		Pixiv: PixivConfig{
			Cookie:     "",
			UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			Accept:     "text/html,application/xhtml+xml,application/xml;q=0.9",
			Delay:      2,
			RetryCount: 3,
			Timeout:    30,
		},
	}

	cm.config.Modules["logger"] = &LoggerConfig{
		Name:    "logger",
		Version: "1.0.0",
		Level:   "info",
		Format:  "text",
		Output:  "stdout",
	}

	return cm.saveUnlocked()
}

// saveUnlocked 保存配置（不加锁版本）
func (cm *ConfigManager) saveUnlocked() error {
	// 确保配置目录存在
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewError(errors.ErrCodeConfigSave, fmt.Sprintf("创建配置目录失败: %v", err))
	}

	// 序列化配置
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return errors.NewError(errors.ErrCodeConfigSave, fmt.Sprintf("序列化配置失败: %v", err))
	}

	// 写入文件
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return errors.NewError(errors.ErrCodeConfigSave, fmt.Sprintf("写入配置文件失败: %v", err))
	}

	return nil
}
