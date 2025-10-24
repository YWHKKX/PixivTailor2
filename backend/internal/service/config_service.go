package service

import (
	"encoding/json"
	"fmt"
	"pixiv-tailor/backend/internal/repository"
)

// ConfigService 配置服务接口
type ConfigService interface {
	GetConfig(module string) (string, error)
	SetConfig(module, config string) error
	ExportConfig(modules []string) (string, error)
	ImportConfig(configData string) error
	ValidateConfig(module, config string) error
}

// configServiceImpl 配置服务实现
type configServiceImpl struct {
	storage repository.Storage
}

// NewConfigService 创建配置服务实例
func NewConfigService(storage repository.Storage) ConfigService {
	return &configServiceImpl{
		storage: storage,
	}
}

// GetConfig 获取配置
func (s *configServiceImpl) GetConfig(module string) (string, error) {
	if module == "" {
		return "", fmt.Errorf("模块名不能为空")
	}

	config, err := s.storage.GetConfig(module)
	if err != nil {
		return "", fmt.Errorf("获取配置失败: %v", err)
	}

	return config, nil
}

// SetConfig 设置配置
func (s *configServiceImpl) SetConfig(module, config string) error {
	if module == "" {
		return fmt.Errorf("模块名不能为空")
	}

	// 验证JSON格式
	if config != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(config), &configMap); err != nil {
			return fmt.Errorf("配置格式无效: %v", err)
		}
	}

	// 验证配置内容
	if err := s.ValidateConfig(module, config); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	if err := s.storage.SetConfig(module, config); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	return nil
}

// ExportConfig 导出配置
func (s *configServiceImpl) ExportConfig(modules []string) (string, error) {
	if len(modules) == 0 {
		// 导出所有模块配置
		modules = []string{"ai", "crawler", "logger", "storage"}
	}

	configMap := make(map[string]interface{})
	for _, module := range modules {
		config, err := s.GetConfig(module)
		if err != nil {
			return "", fmt.Errorf("获取模块 %s 配置失败: %v", module, err)
		}

		var moduleConfig map[string]interface{}
		if config != "" {
			if err := json.Unmarshal([]byte(config), &moduleConfig); err != nil {
				return "", fmt.Errorf("解析模块 %s 配置失败: %v", module, err)
			}
		} else {
			moduleConfig = make(map[string]interface{})
		}

		configMap[module] = moduleConfig
	}

	configData, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %v", err)
	}

	return string(configData), nil
}

// ImportConfig 导入配置
func (s *configServiceImpl) ImportConfig(configData string) error {
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configData), &configMap); err != nil {
		return fmt.Errorf("配置格式无效: %v", err)
	}

	for module, config := range configMap {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("序列化模块 %s 配置失败: %v", module, err)
		}

		if err := s.SetConfig(module, string(configBytes)); err != nil {
			return fmt.Errorf("导入模块 %s 配置失败: %v", module, err)
		}
	}

	return nil
}

// ValidateConfig 验证配置
func (s *configServiceImpl) ValidateConfig(module, config string) error {
	if module == "" {
		return fmt.Errorf("模块名不能为空")
	}

	// 根据模块类型进行特定验证
	switch module {
	case "ai":
		return s.validateAIConfig(config)
	case "crawler":
		return s.validateCrawlerConfig(config)
	case "logger":
		return s.validateLoggerConfig(config)
	case "storage":
		return s.validateStorageConfig(config)
	default:
		// 对于未知模块，只验证JSON格式
		if config != "" {
			var configMap map[string]interface{}
			if err := json.Unmarshal([]byte(config), &configMap); err != nil {
				return fmt.Errorf("配置格式无效: %v", err)
			}
		}
	}

	return nil
}

// validateAIConfig 验证AI模块配置
func (s *configServiceImpl) validateAIConfig(config string) error {
	if config == "" {
		return nil // 空配置是允许的
	}

	var aiConfig map[string]interface{}
	if err := json.Unmarshal([]byte(config), &aiConfig); err != nil {
		return fmt.Errorf("AI配置格式无效: %v", err)
	}

	// 验证必需字段
	if apiURL, exists := aiConfig["api_url"]; exists {
		if apiURLStr, ok := apiURL.(string); !ok || apiURLStr == "" {
			return fmt.Errorf("API URL不能为空")
		}
	}

	return nil
}

// validateCrawlerConfig 验证爬虫模块配置
func (s *configServiceImpl) validateCrawlerConfig(config string) error {
	if config == "" {
		return nil
	}

	var crawlerConfig map[string]interface{}
	if err := json.Unmarshal([]byte(config), &crawlerConfig); err != nil {
		return fmt.Errorf("爬虫配置格式无效: %v", err)
	}

	// 验证延迟设置
	if delay, exists := crawlerConfig["delay"]; exists {
		if delayFloat, ok := delay.(float64); !ok || delayFloat < 0 {
			return fmt.Errorf("延迟时间必须是非负数")
		}
	}

	return nil
}

// validateLoggerConfig 验证日志模块配置
func (s *configServiceImpl) validateLoggerConfig(config string) error {
	if config == "" {
		return nil
	}

	var loggerConfig map[string]interface{}
	if err := json.Unmarshal([]byte(config), &loggerConfig); err != nil {
		return fmt.Errorf("日志配置格式无效: %v", err)
	}

	// 验证日志级别
	if level, exists := loggerConfig["level"]; exists {
		validLevels := []string{"debug", "info", "warn", "error", "fatal"}
		if levelStr, ok := level.(string); !ok || !contains(validLevels, levelStr) {
			return fmt.Errorf("日志级别必须是: debug, info, warn, error, fatal")
		}
	}

	return nil
}

// validateStorageConfig 验证存储模块配置
func (s *configServiceImpl) validateStorageConfig(config string) error {
	if config == "" {
		return nil
	}

	var storageConfig map[string]interface{}
	if err := json.Unmarshal([]byte(config), &storageConfig); err != nil {
		return fmt.Errorf("存储配置格式无效: %v", err)
	}

	// 验证数据库路径
	if dbPath, exists := storageConfig["db_path"]; exists {
		if dbPathStr, ok := dbPath.(string); !ok || dbPathStr == "" {
			return fmt.Errorf("数据库路径不能为空")
		}
	}

	return nil
}
