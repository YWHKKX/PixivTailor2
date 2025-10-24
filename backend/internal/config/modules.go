package config

import (
	"fmt"
)

// ============================================================================
// AI 模块配置
// ============================================================================

// AIConfig AI模块配置
type AIConfig struct {
	Name    string        `json:"name"`
	Version string        `json:"version"`
	SDWebUI SDWebUIConfig `json:"sd_webui"`
	KohyaSS KohyaSSConfig `json:"kohya_ss"`
	OpenAI  OpenAIConfig  `json:"openai"`
}

// SDWebUIConfig Stable Diffusion WebUI配置
type SDWebUIConfig struct {
	URL     string `json:"url"`
	APIKey  string `json:"api_key"`
	Timeout int    `json:"timeout"`
}

// KohyaSSConfig Kohya-ss配置
type KohyaSSConfig struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKeys []string `json:"api_keys"`
	Timeout int      `json:"timeout"`
}

// Validate 验证AI配置
func (c *AIConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("AI模块名称不能为空")
	}
	if c.SDWebUI.URL == "" {
		return fmt.Errorf("SDWebUI URL不能为空")
	}
	if len(c.OpenAI.APIKeys) == 0 {
		return fmt.Errorf("OpenAI API密钥不能为空")
	}
	return nil
}

// GetName 获取模块名称
func (c *AIConfig) GetName() string {
	return c.Name
}

// GetVersion 获取模块版本
func (c *AIConfig) GetVersion() string {
	return c.Version
}

// ============================================================================
// 爬虫模块配置
// ============================================================================

// CrawlerConfig 爬虫模块配置
type CrawlerConfig struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Pixiv   PixivConfig `json:"pixiv"`
}

// PixivConfig Pixiv配置
type PixivConfig struct {
	Cookie     string `json:"cookie"`
	UserAgent  string `json:"user_agent"`
	Accept     string `json:"accept"`
	Delay      int    `json:"delay"`
	RetryCount int    `json:"retry_count"`
	Timeout    int    `json:"timeout"`
}

// Validate 验证爬虫配置
func (c *CrawlerConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("爬虫模块名称不能为空")
	}
	if c.Pixiv.UserAgent == "" {
		return fmt.Errorf("User-Agent不能为空")
	}
	if c.Pixiv.Delay < 0 {
		return fmt.Errorf("延迟时间不能为负数")
	}
	return nil
}

// GetName 获取模块名称
func (c *CrawlerConfig) GetName() string {
	return c.Name
}

// GetVersion 获取模块版本
func (c *CrawlerConfig) GetVersion() string {
	return c.Version
}

// ============================================================================
// 日志模块配置
// ============================================================================

// LoggerConfig 日志模块配置
type LoggerConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Level   string `json:"level"`
	Format  string `json:"format"`
	Output  string `json:"output"`
}

// Validate 验证日志配置
func (c *LoggerConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("日志模块名称不能为空")
	}
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	for _, level := range validLevels {
		if c.Level == level {
			return nil
		}
	}
	return fmt.Errorf("无效的日志级别: %s", c.Level)
}

// GetName 获取模块名称
func (c *LoggerConfig) GetName() string {
	return c.Name
}

// GetVersion 获取模块版本
func (c *LoggerConfig) GetVersion() string {
	return c.Version
}

// ============================================================================
// 配置工厂
// ============================================================================

// ConfigFactory 配置工厂
type ConfigFactory struct{}

// NewConfigFactory 创建配置工厂
func NewConfigFactory() *ConfigFactory {
	return &ConfigFactory{}
}

// CreateModuleConfig 创建模块配置
func (cf *ConfigFactory) CreateModuleConfig(moduleName string) (ModuleConfig, error) {
	switch moduleName {
	case "ai":
		return &AIConfig{
			Name:    "ai",
			Version: "1.0.0",
		}, nil
	case "crawler":
		return &CrawlerConfig{
			Name:    "crawler",
			Version: "1.0.0",
		}, nil
	case "logger":
		return &LoggerConfig{
			Name:    "logger",
			Version: "1.0.0",
		}, nil
	default:
		return nil, fmt.Errorf("未知的模块类型: %s", moduleName)
	}
}

// GetSupportedModules 获取支持的模块列表
func (cf *ConfigFactory) GetSupportedModules() []string {
	return []string{"ai", "crawler", "logger"}
}
