package models

import (
	"time"
)

// GenerationConfig 图片生成配置
type GenerationConfig struct {
	ID             string       `json:"id" gorm:"primaryKey"`             // 配置ID
	CreatedAt      time.Time    `json:"created_at"`                       // 创建时间
	UpdatedAt      time.Time    `json:"updated_at"`                       // 更新时间
	Name           string       `json:"name" gorm:"uniqueIndex;not null"` // 配置名称
	Description    string       `json:"description"`                      // 配置描述
	Category       string       `json:"category"`                         // 配置分类
	Prompt         string       `json:"prompt"`                           // 正面提示词
	NegativePrompt string       `json:"negative_prompt"`                  // 负面提示词
	Steps          int          `json:"steps"`                            // 采样步数
	CFGScale       float64      `json:"cfg_scale"`                        // CFG Scale
	Width          int          `json:"width"`                            // 宽度
	Height         int          `json:"height"`                           // 高度
	Seed           int64        `json:"seed"`                             // 随机种子
	Model          string       `json:"model"`                            // 模型名称
	Sampler        string       `json:"sampler"`                          // 采样器
	BatchSize      int          `json:"batch_size"`                       // 批次大小
	EnableHR       bool         `json:"enable_hr"`                        // 高分辨率修复
	Loras          []LoraConfig `json:"loras" gorm:"type:text"`           // LoRA模型列表

	// VAE配置
	VAE string `json:"vae" gorm:"default:''"`

	// 高分辨率修复参数
	HiresSteps        int     `json:"hires_steps" gorm:"default:0"`
	DenoisingStrength float64 `json:"denoising_strength" gorm:"default:0.7"`
	Upscaler          string  `json:"upscaler" gorm:"default:'Latent'"`
	UpscaleBy         float64 `json:"upscale_by" gorm:"default:2.0"`

	// 其他参数
	RestoreFaces bool    `json:"restore_faces" gorm:"default:false"`
	Tiling       bool    `json:"tiling" gorm:"default:false"`
	ClipSkip     int     `json:"clip_skip" gorm:"default:2"`
	Eta          float64 `json:"eta" gorm:"default:0.0"`
	ENSD         float64 `json:"ensd" gorm:"default:31337"`

	// 输出设置
	SaveImages    bool `json:"save_images" gorm:"default:true"`
	SaveGrid      bool `json:"save_grid" gorm:"default:true"`
	SendImages    bool `json:"send_images" gorm:"default:true"`
	DoNotSaveGrid bool `json:"do_not_save_grid" gorm:"default:false"`

	OtherParams string     `json:"other_params" gorm:"type:text"` // 其他参数（JSON字符串）
	IsDefault   bool       `json:"is_default"`                    // 是否为默认配置
	UsageCount  int        `json:"usage_count" gorm:"default:0"`  // 使用次数
	LastUsed    *time.Time `json:"last_used"`                     // 最后使用时间
}

// GenerationConfigRequest 创建/更新配置请求
type GenerationConfigRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	IsDefault   bool                   `json:"is_default"`
	Config      map[string]interface{} `json:"config" binding:"required"`
	Tags        []string               `json:"tags"`
}

// GenerationConfigResponse 配置响应
type GenerationConfigResponse struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Category       string       `json:"category"`
	Prompt         string       `json:"prompt"`
	NegativePrompt string       `json:"negative_prompt"`
	Steps          int          `json:"steps"`
	CFGScale       float64      `json:"cfg_scale"`
	Width          int          `json:"width"`
	Height         int          `json:"height"`
	Seed           int64        `json:"seed"`
	Model          string       `json:"model"`
	Sampler        string       `json:"sampler"`
	BatchSize      int          `json:"batch_size"`
	EnableHR       bool         `json:"enable_hr"`
	Loras          []LoraConfig `json:"loras"`

	// VAE配置
	VAE string `json:"vae"`

	// 高分辨率修复参数
	HiresSteps        int     `json:"hires_steps"`
	HiresUpscaler     string  `json:"hires_upscaler"`
	HiresUpscale      float64 `json:"hires_upscale"`
	DenoisingStrength float64 `json:"denoising_strength"`
	Upscaler          string  `json:"upscaler"`
	UpscaleBy         float64 `json:"upscale_by"`

	// 其他参数
	RestoreFaces bool    `json:"restore_faces"`
	Tiling       bool    `json:"tiling"`
	ClipSkip     int     `json:"clip_skip"`
	Eta          float64 `json:"eta"`
	ENSD         float64 `json:"ensd"`

	// 输出设置
	SaveImages    bool `json:"save_images"`
	SaveGrid      bool `json:"save_grid"`
	SendImages    bool `json:"send_images"`
	DoNotSaveGrid bool `json:"do_not_save_grid"`

	OtherParams string     `json:"other_params"` // JSON字符串
	IsDefault   bool       `json:"is_default"`
	UsageCount  int        `json:"usage_count"`
	LastUsed    *time.Time `json:"last_used"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ConfigListResponse 配置列表响应
type ConfigListResponse struct {
	Configs    []*GenerationConfigResponse `json:"configs"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	Categories []string                    `json:"categories"`
}

// SDWebUIConfig SD WebUI配置结构
type SDWebUIConfig struct {
	// 基础参数
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Seed           int64   `json:"seed"`
	Sampler        string  `json:"sampler"`
	BatchSize      int     `json:"batch_size"`
	BatchCount     int     `json:"batch_count"`

	// 高分辨率修复
	EnableHR          bool    `json:"enable_hr"`
	HiresSteps        int     `json:"hires_steps"`
	DenoisingStrength float64 `json:"denoising_strength"`
	Upscaler          string  `json:"upscaler"`
	UpscaleBy         float64 `json:"upscale_by"`

	// 模型相关
	Model string       `json:"model"`
	VAE   string       `json:"vae"`
	Loras []LoraConfig `json:"loras"`

	// 其他参数
	RestoreFaces bool    `json:"restore_faces"`
	Tiling       bool    `json:"tiling"`
	ClipSkip     int     `json:"clip_skip"`
	Eta          float64 `json:"eta"`
	ENSD         float64 `json:"ensd"`

	// 输出设置
	SaveImages    bool `json:"save_images"`
	SaveGrid      bool `json:"save_grid"`
	SendImages    bool `json:"send_images"`
	DoNotSaveGrid bool `json:"do_not_save_grid"`
}

// ConfigImportRequest 配置导入请求
type ConfigImportRequest struct {
	ConfigName string                 `json:"config_name" binding:"required"`
	Config     map[string]interface{} `json:"config" binding:"required"`
	Category   string                 `json:"category"`
	Tags       []string               `json:"tags"`
}

// ConfigExportRequest 配置导出请求
type ConfigExportRequest struct {
	ConfigIDs []string `json:"config_ids" binding:"required"`
	Format    string   `json:"format"` // json, yaml
}

// ConfigSearchRequest 配置搜索请求
type ConfigSearchRequest struct {
	Query     string   `json:"query"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	Page      int      `json:"page"`
	PageSize  int      `json:"page_size"`
	SortBy    string   `json:"sort_by"`    // name, created_at, usage_count, last_used
	SortOrder string   `json:"sort_order"` // asc, desc
}
