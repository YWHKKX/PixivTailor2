package models

import "time"

// CharacterProfile 角色特征配置文件
type CharacterProfile struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`                   // 角色名称
	Description string             `json:"description"`            // 角色描述
	Tags        []string           `json:"tags"`                   // 关键特征标签
	TagWeights  map[string]float64 `json:"tag_weights"`            // 标签权重
	SourceFiles []string           `json:"source_files,omitempty"` // 来源tag文件（已废弃，不使用）
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ExtractCharacterTags 提取角色标签统计
type ExtractCharacterTags struct {
	TagFrequency map[string]int     `json:"tag_frequency"` // 标签出现频率
	TagWeights   map[string]float64 `json:"tag_weights"`   // 标签平均权重
	TotalFiles   int                `json:"total_files"`   // 文件总数
}

// CharacterProfileRequest 创建角色特征请求
type CharacterProfileRequest struct {
	Name         string   `json:"name"`          // 角色名称
	Description  string   `json:"description"`   // 角色描述
	SourceDir    string   `json:"source_dir"`    // 来源tag目录
	MinFrequency int      `json:"min_frequency"` // 最小出现频率
	MinWeight    float64  `json:"min_weight"`    // 最小权重阈值
	MaxTags      int      `json:"max_tags"`      // 最大标签数量
	CoreTags     []string `json:"core_tags"`     // 核心关键词（如 character_name）
	IncludeTags  []string `json:"include_tags"`  // 强制包含的标签
	ExcludeTags  []string `json:"exclude_tags"`  // 要排除的标签
}
