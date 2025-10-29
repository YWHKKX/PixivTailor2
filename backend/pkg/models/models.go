package models

import (
	"time"
)

// ============================================================================
// 通用模型
// ============================================================================

// BaseModel 基础模型
type BaseModel struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Pagination 分页信息
type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// ============================================================================
// 爬虫相关模型
// ============================================================================

// CrawlRequest 爬取请求
type CrawlRequest struct {
	Type     CrawlType `json:"type"`
	Query    string    `json:"query"`
	UserID   int       `json:"user_id,omitempty"`
	IllustID int       `json:"illust_id,omitempty"`
	Order    Order     `json:"order"`
	Mode     Mode      `json:"mode"`
	Limit    int       `json:"limit"`
	Delay    int       `json:"delay"`
}

// CrawlType 爬取类型
type CrawlType string

const (
	CrawlTypeTag    CrawlType = "tag"
	CrawlTypeUser   CrawlType = "user"
	CrawlTypeIllust CrawlType = "illust"
)

// Order 排序方式
type Order string

const (
	OrderDateD    Order = "date_d"
	OrderPopularD Order = "popular_d"
)

// Mode 模式
type Mode string

const (
	ModeSafe Mode = "safe"
	ModeR18  Mode = "r18"
	ModeAll  Mode = "all"
)

// PixivImage Pixiv图像信息
type PixivImage struct {
	BaseModel
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	AuthorID     int      `json:"author_id"`
	Tags         []string `json:"tags"`
	URL          string   `json:"url"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Bookmarks    int      `json:"bookmarks"`
	Views        int      `json:"views"`
	IsR18        bool     `json:"is_r18"`
}

// ============================================================================
// AI生成相关模型
// ============================================================================

// GenerateRequest 生成请求
type GenerateRequest struct {
	Model          string                 `json:"model"`
	Prompt         string                 `json:"prompt"`
	NegativePrompt string                 `json:"negative_prompt"`
	Loras          []LoraConfig           `json:"loras"`
	Poses          []PoseConfig           `json:"poses"`
	BatchSize      int                    `json:"batch_size"`
	BatchCount     int                    `json:"batch_count"`
	Steps          int                    `json:"steps"`
	CFGScale       float64                `json:"cfg_scale"`
	Width          int                    `json:"width"`
	Height         int                    `json:"height"`
	Seed           int                    `json:"seed"`
	Sampler        string                 `json:"sampler"`
	SavePath       string                 `json:"save_path"`
	EnableHR       bool                   `json:"enable_hr"`
	Options        map[string]interface{} `json:"options"`
}

// LoraConfig LoRA配置
type LoraConfig struct {
	Name        string   `json:"name"`                  // LoRA名称
	FullName    string   `json:"full_name,omitempty"`   // 完整名称（包含hash）
	Weight      float64  `json:"weight"`                // 权重
	Path        string   `json:"path,omitempty"`        // 文件路径
	Tags        []string `json:"tags,omitempty"`        // 关联标签
	UseMask     bool     `json:"use_mask,omitempty"`    // 是否使用遮罩
	Description string   `json:"description,omitempty"` // 描述
	// 新增字段用于权重写入关键词
	ExtendTags []string `json:"extend_tags,omitempty"` // 扩展标签（用于权重写入）
	LoraKey    string   `json:"lora_key,omitempty"`    // LoRA键名（用于权重写入）
}

// PoseConfig 姿态配置
type PoseConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// GeneratedImage 生成的图像
type GeneratedImage struct {
	BaseModel
	Prompt         string   `json:"prompt"`
	NegativePrompt string   `json:"negative_prompt"`
	Model          string   `json:"model"`
	Loras          []string `json:"loras"`
	ImageURL       string   `json:"image_url"`
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	Seed           int64    `json:"seed"`
	CFGScale       float64  `json:"cfg_scale"`
	Steps          int      `json:"steps"`
	Sampler        string   `json:"sampler"`
}

// ============================================================================
// 训练相关模型
// ============================================================================

// TrainRequest 训练请求
type TrainRequest struct {
	Name           string                 `json:"name"`
	PretrainedPath string                 `json:"pretrained_path"`
	InputDir       string                 `json:"input_dir"`
	OutputDir      string                 `json:"output_dir"`
	Epochs         int                    `json:"epochs"`
	BatchSize      int                    `json:"batch_size"`
	LearningRate   float64                `json:"learning_rate"`
	Tags           map[string]int         `json:"tags"`
	Prompts        string                 `json:"prompts"`
	Options        map[string]interface{} `json:"options"`
}

// TrainedModel 训练好的模型
type TrainedModel struct {
	BaseModel
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Path        string            `json:"path"`
	Size        int64             `json:"size"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}

// ============================================================================
// 标签相关模型
// ============================================================================

// TagRequest 标签请求
type TagRequest struct {
	InputDir   interface{} `json:"input_dir"` // 可以是字符串或字符串数组
	OutputDir  string      `json:"output_dir"`
	Analyzer   string      `json:"analyzer"`
	Model      string      `json:"model,omitempty"` // 模型名称（可选）
	SkipTags   []string    `json:"skip_tags"`
	ExtendTags []string    `json:"extend_tags"`
	TagOrder   string      `json:"tag_order"`
	SaveType   string      `json:"save_type"`
	Limit      int         `json:"limit"`
}

// GetInputDirs 获取输入目录列表（统一返回数组）
func (t *TagRequest) GetInputDirs() []string {
	switch v := t.InputDir.(type) {
	case string:
		if v == "" {
			return []string{}
		}
		return []string{v}
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, dir := range v {
			if dirStr, ok := dir.(string); ok && dirStr != "" {
				result = append(result, dirStr)
			}
		}
		return result
	case []string:
		result := make([]string, 0, len(v))
		for _, dir := range v {
			if dir != "" {
				result = append(result, dir)
			}
		}
		return result
	default:
		return []string{}
	}
}

// TaggedImage 已标签的图像
type TaggedImage struct {
	BaseModel
	ImagePath string            `json:"image_path"`
	Tags      []Tag             `json:"tags"`
	Analyzer  string            `json:"analyzer"`
	Metadata  map[string]string `json:"metadata"`
}

// Tag 标签
type Tag struct {
	Name      string  `json:"name"`
	Score     float64 `json:"score"`
	Category  string  `json:"category"`
	IsGeneral bool    `json:"is_general"`
}

// ============================================================================
// 分类相关模型
// ============================================================================

// ClassifyRequest 分类请求
type ClassifyRequest struct {
	Input    []string `json:"input"`
	Output   string   `json:"output"`
	APIKeys  []string `json:"api_keys"`
	Limit    int      `json:"limit"`
	ShowTags bool     `json:"show_tags"`
}

// GlobalTags 全局标签
type GlobalTags map[string][]string

// TagCategories 标签分类
type TagCategories map[string][]string

// ============================================================================
// 任务相关模型
// ============================================================================

// Task 任务
type Task struct {
	BaseModel
	Name         string            `json:"name"`
	Type         TaskType          `json:"type"`
	Status       TaskStatus        `json:"status"`
	Parameters   map[string]string `json:"parameters"`
	Progress     int32             `json:"progress"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

// TaskType 任务类型
type TaskType string

const (
	TaskTypeCrawl    TaskType = "crawl"
	TaskTypeGenerate TaskType = "generate"
	TaskTypeTrain    TaskType = "train"
	TaskTypeTag      TaskType = "tag"
	TaskTypeClassify TaskType = "classify"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// ProgressInfo 进度信息
type ProgressInfo struct {
	TaskID      string    `json:"task_id"`
	Progress    int32     `json:"progress"`
	CurrentStep string    `json:"current_step"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// ============================================================================
// 系统相关模型
// ============================================================================

// SystemStatus 系统状态
type SystemStatus struct {
	Version string            `json:"version"`
	Status  string            `json:"status"`
	Uptime  int64             `json:"uptime"`
	Modules map[string]string `json:"modules"`
	Metrics SystemMetrics     `json:"metrics"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    int64   `json:"memory_usage"`
	DiskUsage      int64   `json:"disk_usage"`
	ActiveTasks    int32   `json:"active_tasks"`
	CompletedTasks int32   `json:"completed_tasks"`
	FailedTasks    int32   `json:"failed_tasks"`
}
