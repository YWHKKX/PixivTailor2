package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	configs  map[string]interface{}
	mutex    sync.RWMutex
	filePath string
}

// NewConfigManager 创建配置管理器
func NewConfigManager(filePath string) *ConfigManager {
	cm := &ConfigManager{
		configs:  make(map[string]interface{}),
		filePath: filePath,
	}

	// 加载现有配置
	cm.LoadConfig()
	return cm
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查配置文件是否存在
	if _, err := os.Stat(cm.filePath); os.IsNotExist(err) {
		// 创建默认配置
		cm.configs = cm.getDefaultConfigs()
		return cm.SaveConfig()
	}

	// 读取配置文件
	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON配置
	if err := json.Unmarshal(data, &cm.configs); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig() error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 确保目录存在
	dir := filepath.Dir(cm.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(cm.configs, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(cm.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig(module string) (interface{}, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if config, exists := cm.configs[module]; exists {
		return config, nil
	}

	return nil, fmt.Errorf("模块 %s 的配置不存在", module)
}

// SetConfig 设置配置
func (cm *ConfigManager) SetConfig(module string, config interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.configs[module] = config
	return cm.SaveConfig()
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(module string, updates map[string]interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 获取现有配置
	existingConfig, exists := cm.configs[module]
	if !exists {
		existingConfig = make(map[string]interface{})
	}

	// 转换为map
	configMap, ok := existingConfig.(map[string]interface{})
	if !ok {
		configMap = make(map[string]interface{})
	}

	// 更新配置
	for key, value := range updates {
		configMap[key] = value
	}

	cm.configs[module] = configMap
	return cm.SaveConfig()
}

// GetAllConfigs 获取所有配置
func (cm *ConfigManager) GetAllConfigs() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 返回配置的副本
	result := make(map[string]interface{})
	for k, v := range cm.configs {
		result[k] = v
	}
	return result
}

// getDefaultConfigs 获取默认配置
func (cm *ConfigManager) getDefaultConfigs() map[string]interface{} {
	return map[string]interface{}{
		"pixiv": map[string]interface{}{
			"user_agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"accept":          "application/json, text/plain, */*",
			"accept_language": "zh-CN,zh;q=0.9,en;q=0.8",
			"referer":         "https://www.pixiv.net/",
			"timeout":         30,
			"retry_count":     3,
			"default_delay":   2,
			"proxy": map[string]interface{}{
				"enabled": false,
				"url":     "http://127.0.0.1:7890",
			},
		},
		"ai": map[string]interface{}{
			"sd_webui_url":    "http://127.0.0.1:7860",
			"wd14_tagger_url": "http://127.0.0.1:7861",
			"kohya_ss_url":    "http://127.0.0.1:7862",
			"openai_api_keys": []string{},
			"default_timeout": 60,
		},
		"download": map[string]interface{}{
			"save_dir":       "data/images",
			"max_concurrent": 5,
			"retry_count":    3,
			"timeout":        30,
		},
		"cache": map[string]interface{}{
			"cache_dir":       "data/cache",
			"max_size":        "1GB",
			"expire_duration": "24h",
		},
		"system": map[string]interface{}{
			"log_level":     "info",
			"max_log_files": 10,
			"log_file_size": "10MB",
		},
	}
}
