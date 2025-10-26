package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"pixiv-tailor/backend/pkg/paths"
)

// 扩展现有的AIConfig结构体
type ExtendedAIConfig struct {
	AIConfig
	WD14Tagger WD14TaggerConfig `json:"wd14tagger"`
	Timeout    int              `json:"timeout"`
	RetryCount int              `json:"retry_count"`
	RetryDelay int              `json:"retry_delay"`
	MaxWorkers int              `json:"max_workers"`
	BatchSize  int              `json:"batch_size"`
	QueueSize  int              `json:"queue_size"`
}

// 扩展SDWebUIConfig
type ExtendedSDWebUIConfig struct {
	SDWebUIConfig
	RetryCount     int                    `json:"retry_count"`
	DefaultModel   string                 `json:"default_model"`
	DefaultSampler string                 `json:"default_sampler"`
	DefaultSteps   int                    `json:"default_steps"`
	DefaultCFG     float64                `json:"default_cfg"`
	DefaultWidth   int                    `json:"default_width"`
	DefaultHeight  int                    `json:"default_height"`
	MaxBatchSize   int                    `json:"max_batch_size"`
	MaxSteps       int                    `json:"max_steps"`
	MaxCFG         float64                `json:"max_cfg"`
	MaxWidth       int                    `json:"max_width"`
	MaxHeight      int                    `json:"max_height"`
	Models         []string               `json:"models"`
	Samplers       []string               `json:"samplers"`
	Options        map[string]interface{} `json:"options"`
}

// 扩展KohyaSSConfig
type ExtendedKohyaSSConfig struct {
	KohyaSSConfig
	APIURL           string                 `json:"api_url"`
	RetryCount       int                    `json:"retry_count"`
	DefaultEpochs    int                    `json:"default_epochs"`
	DefaultBatchSize int                    `json:"default_batch_size"`
	DefaultLR        float64                `json:"default_lr"`
	MaxEpochs        int                    `json:"max_epochs"`
	MaxBatchSize     int                    `json:"max_batch_size"`
	MaxLR            float64                `json:"max_lr"`
	MinLR            float64                `json:"min_lr"`
	Options          map[string]interface{} `json:"options"`
}

// 扩展OpenAIConfig
type ExtendedOpenAIConfig struct {
	OpenAIConfig
	Model            string                 `json:"model"`
	RetryCount       int                    `json:"retry_count"`
	MaxTokens        int                    `json:"max_tokens"`
	Temperature      float64                `json:"temperature"`
	TopP             float64                `json:"top_p"`
	FrequencyPenalty float64                `json:"frequency_penalty"`
	PresencePenalty  float64                `json:"presence_penalty"`
	RateLimit        int                    `json:"rate_limit"`
	Options          map[string]interface{} `json:"options"`
}

// WD14TaggerConfig WD14Tagger配置
type WD14TaggerConfig struct {
	ModelPath  string                 `json:"model_path"`
	Threshold  float64                `json:"threshold"`
	BatchSize  int                    `json:"batch_size"`
	MaxTags    int                    `json:"max_tags"`
	SkipTags   []string               `json:"skip_tags"`
	ExtendTags []string               `json:"extend_tags"`
	TagOrder   string                 `json:"tag_order"`
	SaveType   string                 `json:"save_type"`
	Options    map[string]interface{} `json:"options"`
}

// Config 完整配置结构
type Config struct {
	System   SystemConfig   `json:"system"`
	Modules  ModulesConfig  `json:"modules"`
	User     UserConfig     `json:"user"`
	Metadata MetadataConfig `json:"metadata"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Paths       PathsConfig       `json:"paths"`
	Security    SecurityConfig    `json:"security"`
	Performance PerformanceConfig `json:"performance"`
	Debug       DebugConfig       `json:"debug"`
}

// ModulesConfig 模块配置
type ModulesConfig struct {
	AI      AIConfig      `json:"ai"`
	Crawler CrawlerConfig `json:"crawler"`
	Logger  LoggerConfig  `json:"logger"`
	Utils   UtilsConfig   `json:"utils"`
}

// 其他配置结构体（简化版）
type PathsConfig struct {
	BaseDir   string `json:"base_dir"`
	ConfigDir string `json:"config_dir"`
	DataDir   string `json:"data_dir"`
	LogDir    string `json:"log_dir"`
	TempDir   string `json:"temp_dir"`
	CacheDir  string `json:"cache_dir"`
	BackupDir string `json:"backup_dir"`
	ModelsDir string `json:"models_dir"`
	ImagesDir string `json:"images_dir"`
	TagsDir   string `json:"tags_dir"`
	PosesDir  string `json:"poses_dir"`
}

type SecurityConfig struct {
	EncryptionKey string   `json:"encryption_key"`
	AllowedHosts  []string `json:"allowed_hosts"`
	MaxFileSize   int64    `json:"max_file_size"`
	MaxMemory     int64    `json:"max_memory"`
	RateLimit     int      `json:"rate_limit"`
}

type PerformanceConfig struct {
	MaxWorkers      int `json:"max_workers"`
	QueueSize       int `json:"queue_size"`
	CacheSize       int `json:"cache_size"`
	BatchSize       int `json:"batch_size"`
	Timeout         int `json:"timeout"`
	RetryCount      int `json:"retry_count"`
	RetryDelay      int `json:"retry_delay"`
	CleanupInterval int `json:"cleanup_interval"`
}

type DebugConfig struct {
	Enabled   bool   `json:"enabled"`
	LogLevel  string `json:"log_level"`
	Profiling bool   `json:"profiling"`
	Trace     bool   `json:"trace"`
	Verbose   bool   `json:"verbose"`
	DryRun    bool   `json:"dry_run"`
}

type UtilsConfig struct {
	Name       string                 `json:"name"`
	Version    string                 `json:"version"`
	Enabled    bool                   `json:"enabled"`
	File       map[string]interface{} `json:"file"`
	String     map[string]interface{} `json:"string"`
	Network    map[string]interface{} `json:"network"`
	Cache      map[string]interface{} `json:"cache"`
	Timeout    int                    `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
	RetryDelay int                    `json:"retry_delay"`
	MaxWorkers int                    `json:"max_workers"`
}

type UserConfig struct {
	Preferences    map[string]interface{} `json:"preferences"`
	CustomSettings map[string]interface{} `json:"custom_settings"`
	Language       string                 `json:"language"`
	Theme          string                 `json:"theme"`
	Timezone       string                 `json:"timezone"`
}

type MetadataConfig struct {
	Version    string `json:"version"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	LastBackup string `json:"last_backup"`
	Checksum   string `json:"checksum"`
}

var (
	globalAIConfig *ExtendedAIConfig
)

// LoadAIConfig 加载AI配置
func LoadAIConfig() (*ExtendedAIConfig, error) {
	if globalAIConfig != nil {
		return globalAIConfig, nil
	}

	// 获取路径管理器
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return nil, fmt.Errorf("路径管理器未初始化")
	}

	// 读取配置文件
	configPath := pathManager.GetMainConfigPath()
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 创建扩展的AI配置
	extendedConfig := &ExtendedAIConfig{
		AIConfig: config.Modules.AI,
		// 设置默认值
		Timeout:    30,
		RetryCount: 3,
		RetryDelay: 1,
		MaxWorkers: 4,
		BatchSize:  4,
		QueueSize:  100,
	}

	// 应用环境变量覆盖
	applyEnvOverrides(extendedConfig)

	globalAIConfig = extendedConfig
	return globalAIConfig, nil
}

// applyEnvOverrides 应用环境变量覆盖
func applyEnvOverrides(aiConfig *ExtendedAIConfig) {
	// SD WebUI配置覆盖
	if url := os.Getenv("SD_WEBUI_URL"); url != "" {
		aiConfig.SDWebUI.URL = url
	}
	if timeout := os.Getenv("SD_WEBUI_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			aiConfig.SDWebUI.Timeout = t
		}
	}

	// OpenAI配置覆盖
	if apiKeys := os.Getenv("OPENAI_API_KEYS"); apiKeys != "" {
		aiConfig.OpenAI.APIKeys = strings.Split(apiKeys, ",")
	}
	if timeout := os.Getenv("OPENAI_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			aiConfig.OpenAI.Timeout = t
		}
	}

	// Kohya-ss配置覆盖
	if url := os.Getenv("KOHYA_SS_URL"); url != "" {
		aiConfig.KohyaSS.URL = url
	}
	if timeout := os.Getenv("KOHYA_SS_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			aiConfig.KohyaSS.Timeout = t
		}
	}

	// WD14Tagger配置覆盖
	if modelPath := os.Getenv("WD14TAGGER_MODEL_PATH"); modelPath != "" {
		aiConfig.WD14Tagger.ModelPath = modelPath
	}
}

// GetAIConfig 获取AI配置（单例模式）
func GetAIConfig() *ExtendedAIConfig {
	if globalAIConfig == nil {
		config, err := LoadAIConfig()
		if err != nil {
			// 返回默认配置
			return getDefaultAIConfig()
		}
		return config
	}
	return globalAIConfig
}

// getDefaultAIConfig 获取默认AI配置
func getDefaultAIConfig() *ExtendedAIConfig {
	return &ExtendedAIConfig{
		AIConfig: AIConfig{
			Name:    "ai",
			Version: "1.0.0",
			SDWebUI: SDWebUIConfig{
				URL:     "http://127.0.0.1:7860",
				APIKey:  "",
				Timeout: 300,
			},
			KohyaSS: KohyaSSConfig{
				URL:     "http://127.0.0.1:7861",
				Timeout: 3600,
			},
			OpenAI: OpenAIConfig{
				APIKeys: []string{},
				Timeout: 60,
			},
		},
		WD14Tagger: WD14TaggerConfig{
			ModelPath:  "",
			Threshold:  0.35,
			BatchSize:  1,
			MaxTags:    100,
			SkipTags:   []string{"low_quality", "blurry"},
			ExtendTags: []string{"high_quality", "detailed"},
			TagOrder:   "character",
			SaveType:   "txt",
			Options:    make(map[string]interface{}),
		},
		Timeout:    30,
		RetryCount: 3,
		RetryDelay: 1,
		MaxWorkers: 4,
		BatchSize:  4,
		QueueSize:  100,
	}
}

// ReloadAIConfig 重新加载AI配置
func ReloadAIConfig() error {
	globalAIConfig = nil
	_, err := LoadAIConfig()
	return err
}
