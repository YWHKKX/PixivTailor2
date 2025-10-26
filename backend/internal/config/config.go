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

// EmptyModuleConfig 空模块配置，用于未知模块
type EmptyModuleConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Validate 验证配置
func (e *EmptyModuleConfig) Validate() error {
	return nil
}

// GetName 获取模块名称
func (e *EmptyModuleConfig) GetName() string {
	return e.Name
}

// GetVersion 获取模块版本
func (e *EmptyModuleConfig) GetVersion() string {
	return e.Version
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
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return errors.NewError(errors.ErrCodeConfigParse, fmt.Sprintf("解析配置文件失败: %v", err))
	}

	// 创建配置对象
	config := &GlobalConfig{
		Version:   "1.0.0",
		Modules:   make(map[string]ModuleConfig),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 解析版本
	if v, ok := rawConfig["version"].(string); ok {
		config.Version = v
	}

	// 解析模块配置
	if modules, ok := rawConfig["modules"].(map[string]interface{}); ok {
		for moduleName, moduleData := range modules {
			moduleConfig, err := cm.parseModuleConfig(moduleName, moduleData)
			if err != nil {
				return errors.NewError(errors.ErrCodeConfigParse, fmt.Sprintf("解析模块 %s 配置失败: %v", moduleName, err))
			}
			config.Modules[moduleName] = moduleConfig
		}
	}

	cm.config = config
	return nil
}

// parseModuleConfig 解析模块配置
func (cm *ConfigManager) parseModuleConfig(moduleName string, moduleData interface{}) (ModuleConfig, error) {
	// 将模块数据转换为JSON字节
	moduleBytes, err := json.Marshal(moduleData)
	if err != nil {
		return nil, fmt.Errorf("序列化模块数据失败: %v", err)
	}

	// 根据模块名称创建对应的配置对象
	switch moduleName {
	case "ai":
		var aiConfig AIConfig
		if err := json.Unmarshal(moduleBytes, &aiConfig); err != nil {
			return nil, fmt.Errorf("解析AI配置失败: %v", err)
		}
		return &aiConfig, nil
	case "crawler":
		var crawlerConfig CrawlerConfig
		if err := json.Unmarshal(moduleBytes, &crawlerConfig); err != nil {
			return nil, fmt.Errorf("解析爬虫配置失败: %v", err)
		}
		return &crawlerConfig, nil
	case "logger":
		var loggerConfig LoggerConfig
		if err := json.Unmarshal(moduleBytes, &loggerConfig); err != nil {
			return nil, fmt.Errorf("解析日志配置失败: %v", err)
		}
		return &loggerConfig, nil
	default:
		// 对于未知模块，返回一个空的配置对象
		logger.Warnf("未知的模块类型: %s，跳过", moduleName)
		return &EmptyModuleConfig{Name: moduleName, Version: "1.0.0"}, nil
	}
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
		Level:   map[string]interface{}{"default": "info"},
		Format:  map[string]interface{}{"type": "text"},
		Output:  map[string]interface{}{"type": "stdout"},
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
