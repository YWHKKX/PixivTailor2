package repository

import (
	"fmt"
	"time"

	"pixiv-tailor/backend/pkg/models"

	"gorm.io/gorm"
)

// GenerationConfigStorage 配置文件存储接口
type GenerationConfigStorage interface {
	CreateConfig(config *models.GenerationConfig) error
	GetConfig(id string) (*models.GenerationConfig, error)
	GetConfigByName(name string) (*models.GenerationConfig, error)
	UpdateConfig(config *models.GenerationConfig) error
	DeleteConfig(id string) error
	ListConfigs(page, pageSize int, category, query string, tags []string, sortBy, sortOrder string) ([]*models.GenerationConfig, int64, error)
	GetCategories() ([]string, error)
	GetDefaultConfig() (*models.GenerationConfig, error)
	SetDefaultConfig(id string) error
	IncrementUsageCount(id string) error
	SearchConfigs(query string, limit int) ([]*models.GenerationConfig, error)
	ImportConfig(config *models.GenerationConfig) error
	ExportConfigs(ids []string) ([]*models.GenerationConfig, error)
}

// generationConfigStorage 配置文件存储实现
type generationConfigStorage struct {
	db *gorm.DB
}

// NewGenerationConfigStorage 创建配置文件存储实例
func NewGenerationConfigStorage(db *gorm.DB) GenerationConfigStorage {
	return &generationConfigStorage{db: db}
}

// CreateConfig 创建配置
func (s *generationConfigStorage) CreateConfig(config *models.GenerationConfig) error {
	// 如果设置为默认配置，先取消其他默认配置
	if config.IsDefault {
		if err := s.db.Model(&models.GenerationConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("取消其他默认配置失败: %v", err)
		}
	}

	return s.db.Create(config).Error
}

// GetConfig 获取配置
func (s *generationConfigStorage) GetConfig(id string) (*models.GenerationConfig, error) {
	var config models.GenerationConfig
	err := s.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetConfigByName 根据名称获取配置
func (s *generationConfigStorage) GetConfigByName(name string) (*models.GenerationConfig, error) {
	var config models.GenerationConfig
	err := s.db.Where("name = ?", name).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateConfig 更新配置
func (s *generationConfigStorage) UpdateConfig(config *models.GenerationConfig) error {
	// 如果设置为默认配置，先取消其他默认配置
	if config.IsDefault {
		if err := s.db.Model(&models.GenerationConfig{}).Where("is_default = ? AND id != ?", true, config.ID).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("取消其他默认配置失败: %v", err)
		}
	}

	return s.db.Save(config).Error
}

// DeleteConfig 删除配置
func (s *generationConfigStorage) DeleteConfig(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.GenerationConfig{}).Error
}

// ListConfigs 列出配置
func (s *generationConfigStorage) ListConfigs(page, pageSize int, category, query string, tags []string, sortBy, sortOrder string) ([]*models.GenerationConfig, int64, error) {
	var configs []*models.GenerationConfig
	var total int64

	queryBuilder := s.db.Model(&models.GenerationConfig{})

	// 分类过滤
	if category != "" {
		queryBuilder = queryBuilder.Where("category = ?", category)
	}

	// 搜索查询
	if query != "" {
		queryBuilder = queryBuilder.Where("name LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%")
	}

	// 标签过滤
	if len(tags) > 0 {
		for _, tag := range tags {
			queryBuilder = queryBuilder.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	// 计算总数
	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC"
	if sortBy != "" {
		order := "ASC"
		if sortOrder == "desc" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", sortBy, order)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := queryBuilder.Order(orderBy).Offset(offset).Limit(pageSize).Find(&configs).Error
	if err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}

// GetCategories 获取所有分类
func (s *generationConfigStorage) GetCategories() ([]string, error) {
	var categories []string
	err := s.db.Model(&models.GenerationConfig{}).Distinct("category").Where("category != ''").Pluck("category", &categories).Error
	return categories, err
}

// GetDefaultConfig 获取默认配置
func (s *generationConfigStorage) GetDefaultConfig() (*models.GenerationConfig, error) {
	var config models.GenerationConfig
	err := s.db.Where("is_default = ?", true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// SetDefaultConfig 设置默认配置
func (s *generationConfigStorage) SetDefaultConfig(id string) error {
	// 先取消所有默认配置
	if err := s.db.Model(&models.GenerationConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return fmt.Errorf("取消其他默认配置失败: %v", err)
	}

	// 设置新的默认配置
	return s.db.Model(&models.GenerationConfig{}).Where("id = ?", id).Update("is_default", true).Error
}

// IncrementUsageCount 增加使用次数
func (s *generationConfigStorage) IncrementUsageCount(id string) error {
	now := time.Now()
	return s.db.Model(&models.GenerationConfig{}).Where("id = ?", id).Updates(map[string]interface{}{
		"usage_count": gorm.Expr("usage_count + 1"),
		"last_used":   &now,
	}).Error
}

// SearchConfigs 搜索配置
func (s *generationConfigStorage) SearchConfigs(query string, limit int) ([]*models.GenerationConfig, error) {
	var configs []*models.GenerationConfig
	err := s.db.Where("name LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%").
		Order("usage_count DESC, created_at DESC").
		Limit(limit).
		Find(&configs).Error
	return configs, err
}

// ImportConfig 导入配置
func (s *generationConfigStorage) ImportConfig(config *models.GenerationConfig) error {
	// 检查是否已存在同名配置
	existing, err := s.GetConfigByName(config.Name)
	if err == nil && existing != nil {
		// 更新现有配置
		config.ID = existing.ID
		config.CreatedAt = existing.CreatedAt
		config.UpdatedAt = time.Now()
		return s.UpdateConfig(config)
	}

	// 创建新配置
	return s.CreateConfig(config)
}

// ExportConfigs 导出配置
func (s *generationConfigStorage) ExportConfigs(ids []string) ([]*models.GenerationConfig, error) {
	var configs []*models.GenerationConfig
	err := s.db.Where("id IN ?", ids).Find(&configs).Error
	return configs, err
}

// ConvertToResponse 转换为响应格式
func ConvertToResponse(config *models.GenerationConfig) *models.GenerationConfigResponse {
	return &models.GenerationConfigResponse{
		ID:             config.ID,
		Name:           config.Name,
		Description:    config.Description,
		Category:       config.Category,
		Prompt:         config.Prompt,
		NegativePrompt: config.NegativePrompt,
		Steps:          config.Steps,
		CFGScale:       config.CFGScale,
		Width:          config.Width,
		Height:         config.Height,
		Seed:           config.Seed,
		Model:          config.Model,
		Sampler:        config.Sampler,
		BatchSize:      config.BatchSize,
		EnableHR:       config.EnableHR,
		Loras:          config.Loras,

		// VAE配置
		VAE: config.VAE,

		// 高分辨率修复参数
		HiresSteps:        config.HiresSteps,
		DenoisingStrength: config.DenoisingStrength,
		Upscaler:          config.Upscaler,
		UpscaleBy:         config.UpscaleBy,

		// 其他参数
		RestoreFaces: config.RestoreFaces,
		Tiling:       config.Tiling,
		ClipSkip:     config.ClipSkip,
		Eta:          config.Eta,
		ENSD:         config.ENSD,

		// 输出设置
		SaveImages:    config.SaveImages,
		SaveGrid:      config.SaveGrid,
		SendImages:    config.SendImages,
		DoNotSaveGrid: config.DoNotSaveGrid,

		OtherParams: config.OtherParams,
		IsDefault:   config.IsDefault,
		UsageCount:  config.UsageCount,
		LastUsed:    config.LastUsed,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}
}

// ConvertSDWebUIConfig 转换SD WebUI配置
func ConvertSDWebUIConfig(config map[string]interface{}) *models.SDWebUIConfig {
	sdConfig := &models.SDWebUIConfig{
		Prompt:            getString(config, "prompt"),
		NegativePrompt:    getString(config, "negative_prompt"),
		Steps:             getInt(config, "steps", 20),
		CFGScale:          getFloat64(config, "cfg_scale", 7.0),
		Width:             getInt(config, "width", 512),
		Height:            getInt(config, "height", 512),
		Seed:              getInt64(config, "seed", -1),
		Sampler:           getString(config, "sampler"),
		BatchSize:         getInt(config, "batch_size", 1),
		BatchCount:        getInt(config, "batch_count", 1),
		EnableHR:          getBool(config, "enable_hr", false),
		HiresSteps:        getInt(config, "hires_steps", 0),
		DenoisingStrength: getFloat64(config, "denoising_strength", 0.7),
		Upscaler:          getString(config, "upscaler"),
		UpscaleBy:         getFloat64(config, "upscale_by", 2.0),
		Model:             getString(config, "model"),
		VAE:               getString(config, "vae"),
		RestoreFaces:      getBool(config, "restore_faces", false),
		Tiling:            getBool(config, "tiling", false),
		ClipSkip:          getInt(config, "clip_skip", 1),
		Eta:               getFloat64(config, "eta", 0),
		ENSD:              getFloat64(config, "ensd", 0),
		SaveImages:        getBool(config, "save_images", true),
		SaveGrid:          getBool(config, "save_grid", false),
		SendImages:        getBool(config, "send_images", true),
		DoNotSaveGrid:     getBool(config, "do_not_save_grid", false),
	}

	// 解析LoRA配置
	if loras, ok := config["loras"].([]interface{}); ok {
		for _, lora := range loras {
			if loraMap, ok := lora.(map[string]interface{}); ok {
				sdConfig.Loras = append(sdConfig.Loras, models.LoraConfig{
					Name:   getString(loraMap, "name"),
					Weight: getFloat64(loraMap, "weight", 1.0),
				})
			}
		}
	}

	return sdConfig
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return defaultValue
}

func getInt64(m map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return defaultValue
}

func getFloat64(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultValue
}

func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}
