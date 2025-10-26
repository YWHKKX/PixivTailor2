package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"

	"github.com/google/uuid"
)

// GenerationConfigService 配置文件服务接口
type GenerationConfigService interface {
	CreateConfig(req *models.GenerationConfigRequest) (*models.GenerationConfigResponse, error)
	GetConfig(id string) (*models.GenerationConfigResponse, error)
	GetConfigByName(name string) (*models.GenerationConfigResponse, error)
	UpdateConfig(id string, req *models.GenerationConfigRequest) (*models.GenerationConfigResponse, error)
	DeleteConfig(id string) error
	ListConfigs(req *models.ConfigSearchRequest) (*models.ConfigListResponse, error)
	GetCategories() ([]string, error)
	GetDefaultConfig() (*models.GenerationConfigResponse, error)
	SetDefaultConfig(id string) error
	UseConfig(id string) (*models.GenerationConfigResponse, error)
	ImportConfig(req *models.ConfigImportRequest) (*models.GenerationConfigResponse, error)
	ExportConfigs(req *models.ConfigExportRequest) ([]*models.GenerationConfigResponse, error)
	// 文件系统配置管理方法
	UpdateConfigFile(id string, config map[string]interface{}) error
	CreateConfigFile(config map[string]interface{}) error
	DeleteConfigFile(id string) error
}

// generationConfigServiceImpl 配置文件服务实现
type generationConfigServiceImpl struct {
	storage repository.GenerationConfigStorage
}

// NewGenerationConfigService 创建配置文件服务实例
func NewGenerationConfigService(storage repository.GenerationConfigStorage) GenerationConfigService {
	return &generationConfigServiceImpl{storage: storage}
}

// CreateConfig 创建配置
func (s *generationConfigServiceImpl) CreateConfig(req *models.GenerationConfigRequest) (*models.GenerationConfigResponse, error) {
	// 配置管理已改为文件系统模式，不支持通过API创建配置
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请直接编辑 backend/data/configs/ 目录下的JSON文件")
}

// GetConfig 获取配置
func (s *generationConfigServiceImpl) GetConfig(id string) (*models.GenerationConfigResponse, error) {
	// 配置管理已改为文件系统模式，请使用文件系统API
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请使用文件系统API获取配置")
}

// GetConfigByName 根据名称获取配置
func (s *generationConfigServiceImpl) GetConfigByName(name string) (*models.GenerationConfigResponse, error) {
	// 配置管理已改为文件系统模式，请使用文件系统API
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请使用文件系统API获取配置")
}

// UpdateConfig 更新配置
func (s *generationConfigServiceImpl) UpdateConfig(id string, req *models.GenerationConfigRequest) (*models.GenerationConfigResponse, error) {
	// 配置管理已改为文件系统模式，请直接编辑JSON文件
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请直接编辑 backend/data/configs/ 目录下的JSON文件")
}

// DeleteConfig 删除配置
func (s *generationConfigServiceImpl) DeleteConfig(id string) error {
	// 配置管理已改为文件系统模式，请直接删除JSON文件
	return fmt.Errorf("配置管理已改为文件系统模式，请直接删除 backend/data/configs/ 目录下的JSON文件")
}

// ListConfigs 列出配置
func (s *generationConfigServiceImpl) ListConfigs(req *models.ConfigSearchRequest) (*models.ConfigListResponse, error) {
	// 配置管理已改为文件系统模式，请使用文件系统API
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请使用文件系统API获取配置列表")
}

// GetCategories 获取所有分类
func (s *generationConfigServiceImpl) GetCategories() ([]string, error) {
	// 配置管理已改为文件系统模式，请使用文件系统API
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请使用文件系统API获取分类")
}

// GetDefaultConfig 获取默认配置
func (s *generationConfigServiceImpl) GetDefaultConfig() (*models.GenerationConfigResponse, error) {
	// 配置管理已改为文件系统模式，请使用文件系统API
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请使用文件系统API获取默认配置")
}

// SetDefaultConfig 设置默认配置
func (s *generationConfigServiceImpl) SetDefaultConfig(id string) error {
	// 配置管理已改为文件系统模式，请直接编辑JSON文件
	return fmt.Errorf("配置管理已改为文件系统模式，请直接编辑 backend/data/configs/ 目录下的JSON文件")
}

// UseConfig 使用配置（增加使用次数）
func (s *generationConfigServiceImpl) UseConfig(id string) (*models.GenerationConfigResponse, error) {
	// 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return nil, fmt.Errorf("Path manager not initialized")
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 查找配置文件
	files, err := filepath.Glob(filepath.Join(configsDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("Failed to read config directory: %v", err)
	}

	var targetFile string
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		var fileConfig map[string]interface{}
		if err := json.Unmarshal(content, &fileConfig); err != nil {
			continue
		}

		if fileID, ok := fileConfig["id"].(string); ok && fileID == id {
			targetFile = file
			break
		}
	}

	if targetFile == "" {
		return nil, fmt.Errorf("Config file not found for ID: %s", id)
	}

	// 读取配置文件
	content, err := ioutil.ReadFile(targetFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %v", err)
	}

	var configData map[string]interface{}
	if err := json.Unmarshal(content, &configData); err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %v", err)
	}

	// 更新使用次数
	useCount, _ := configData["use_count"].(float64)
	configData["use_count"] = useCount + 1
	configData["last_used"] = time.Now().Format(time.RFC3339)

	// 保存更新后的配置
	configDataBytes, err := json.MarshalIndent(configData, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal config: %v", err)
	}

	if err := ioutil.WriteFile(targetFile, configDataBytes, 0644); err != nil {
		return nil, fmt.Errorf("Failed to write config file: %v", err)
	}

	// 转换为响应格式
	response := &models.GenerationConfigResponse{
		ID:             configData["id"].(string),
		Name:           configData["name"].(string),
		Description:    configData["description"].(string),
		Category:       configData["category"].(string),
		Prompt:         configData["prompt"].(string),
		NegativePrompt: configData["negative_prompt"].(string),
		Steps:          int(configData["steps"].(float64)),
		CFGScale:       configData["cfg_scale"].(float64),
		Width:          int(configData["width"].(float64)),
		Height:         int(configData["height"].(float64)),
		Seed:           int64(configData["seed"].(float64)),
		Model:          configData["model"].(string),
		Sampler:        configData["sampler"].(string),
		BatchSize:      int(configData["batch_size"].(float64)),
		EnableHR:       configData["enable_hr"].(bool),
		IsDefault:      configData["is_default"].(bool),
		UsageCount:     int(configData["use_count"].(float64)),
	}

	// 处理时间字段
	if createdAt, ok := configData["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			response.CreatedAt = t
		}
	}
	if updatedAt, ok := configData["updated_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			response.UpdatedAt = t
		}
	}
	if lastUsed, ok := configData["last_used"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastUsed); err == nil {
			response.LastUsed = &t
		}
	}

	// 处理其他可选字段
	if vae, ok := configData["vae"].(string); ok {
		response.VAE = vae
	}
	if hiresSteps, ok := configData["hires_steps"].(float64); ok {
		response.HiresSteps = int(hiresSteps)
	}
	if denoisingStrength, ok := configData["denoising_strength"].(float64); ok {
		response.DenoisingStrength = denoisingStrength
	}
	if upscaler, ok := configData["upscaler"].(string); ok {
		response.Upscaler = upscaler
	}
	if upscaleBy, ok := configData["upscale_by"].(float64); ok {
		response.UpscaleBy = upscaleBy
	}
	if restoreFaces, ok := configData["restore_faces"].(bool); ok {
		response.RestoreFaces = restoreFaces
	}
	if tiling, ok := configData["tiling"].(bool); ok {
		response.Tiling = tiling
	}
	if clipSkip, ok := configData["clip_skip"].(float64); ok {
		response.ClipSkip = int(clipSkip)
	}
	if eta, ok := configData["eta"].(float64); ok {
		response.Eta = eta
	}
	if ensd, ok := configData["ensd"].(float64); ok {
		response.ENSD = ensd
	}
	if saveImages, ok := configData["save_images"].(bool); ok {
		response.SaveImages = saveImages
	}
	if saveGrid, ok := configData["save_grid"].(bool); ok {
		response.SaveGrid = saveGrid
	}
	if sendImages, ok := configData["send_images"].(bool); ok {
		response.SendImages = sendImages
	}
	if doNotSaveGrid, ok := configData["do_not_save_grid"].(bool); ok {
		response.DoNotSaveGrid = doNotSaveGrid
	}
	if otherParams, ok := configData["other_params"].(string); ok {
		response.OtherParams = otherParams
	}

	// 处理LoRA配置
	if lorasData, ok := configData["loras"].([]interface{}); ok {
		loras := make([]models.LoraConfig, 0, len(lorasData))
		for _, loraData := range lorasData {
			if loraMap, ok := loraData.(map[string]interface{}); ok {
				lora := models.LoraConfig{}

				// 基本字段
				if name, ok := loraMap["name"].(string); ok {
					lora.Name = name
				}
				if fullName, ok := loraMap["full_name"].(string); ok {
					lora.FullName = fullName
				}
				if path, ok := loraMap["path"].(string); ok {
					lora.Path = path
				}
				if weight, ok := loraMap["weight"].(float64); ok {
					lora.Weight = weight
				}
				if description, ok := loraMap["description"].(string); ok {
					lora.Description = description
				}
				if useMask, ok := loraMap["use_mask"].(bool); ok {
					lora.UseMask = useMask
				}

				// 处理标签数组
				if tagsData, ok := loraMap["tags"].([]interface{}); ok {
					tags := make([]string, 0, len(tagsData))
					for _, tag := range tagsData {
						if tagStr, ok := tag.(string); ok {
							tags = append(tags, tagStr)
						}
					}
					lora.Tags = tags
				}

				// 处理扩展标签数组（用于权重写入）
				if extendTagsData, ok := loraMap["extend_tags"].([]interface{}); ok {
					extendTags := make([]string, 0, len(extendTagsData))
					for _, tag := range extendTagsData {
						if tagStr, ok := tag.(string); ok {
							extendTags = append(extendTags, tagStr)
						}
					}
					lora.ExtendTags = extendTags
				}

				// 处理LoRA键名
				if loraKey, ok := loraMap["lora_key"].(string); ok {
					lora.LoraKey = loraKey
				}

				loras = append(loras, lora)
			}
		}
		response.Loras = loras
	}

	// 处理LoRA权重写入关键词
	response.Prompt = s.buildLoraPrompt(response.Prompt, response.Loras)

	return response, nil
}

// ImportConfig 导入配置
func (s *generationConfigServiceImpl) ImportConfig(req *models.ConfigImportRequest) (*models.GenerationConfigResponse, error) {
	// 配置管理已改为文件系统模式，请直接编辑JSON文件
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请直接编辑 backend/data/configs/ 目录下的JSON文件")
}

// ExportConfigs 导出配置
func (s *generationConfigServiceImpl) ExportConfigs(req *models.ConfigExportRequest) ([]*models.GenerationConfigResponse, error) {
	// 配置管理已改为文件系统模式，请直接复制JSON文件
	return nil, fmt.Errorf("配置管理已改为文件系统模式，请直接复制 backend/data/configs/ 目录下的JSON文件")
}

// UpdateConfigFile 更新配置文件
func (s *generationConfigServiceImpl) UpdateConfigFile(id string, config map[string]interface{}) error {
	// 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return fmt.Errorf("Path manager not initialized")
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 查找配置文件
	files, err := filepath.Glob(filepath.Join(configsDir, "*.json"))
	if err != nil {
		return fmt.Errorf("Failed to read config directory: %v", err)
	}

	var targetFile string
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		var fileConfig map[string]interface{}
		if err := json.Unmarshal(content, &fileConfig); err != nil {
			continue
		}

		if fileID, ok := fileConfig["id"].(string); ok && fileID == id {
			targetFile = file
			break
		}
	}

	if targetFile == "" {
		return fmt.Errorf("Config file not found for ID: %s", id)
	}

	// 更新配置数据
	config["id"] = id
	config["updated_at"] = time.Now().Format(time.RFC3339)

	// 写入文件
	configData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}

	if err := ioutil.WriteFile(targetFile, configData, 0644); err != nil {
		return fmt.Errorf("Failed to write config file: %v", err)
	}

	return nil
}

// CreateConfigFile 创建配置文件
func (s *generationConfigServiceImpl) CreateConfigFile(config map[string]interface{}) error {
	// 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return fmt.Errorf("Path manager not initialized")
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 确保目录存在
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		return fmt.Errorf("Failed to create config directory: %v", err)
	}

	// 生成ID和文件名
	id := uuid.New().String()
	config["id"] = id
	config["created_at"] = time.Now().Format(time.RFC3339)
	config["updated_at"] = time.Now().Format(time.RFC3339)

	// 生成文件名
	name, _ := config["name"].(string)
	if name == "" {
		name = "untitled"
	}

	// 清理文件名
	safeName := sanitizeFileName(name)
	fileName := fmt.Sprintf("%s-%s.json", safeName, id[:8])
	filePath := filepath.Join(configsDir, fileName)

	// 写入文件
	configData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}

	if err := ioutil.WriteFile(filePath, configData, 0644); err != nil {
		return fmt.Errorf("Failed to write config file: %v", err)
	}

	return nil
}

// DeleteConfigFile 删除配置文件
func (s *generationConfigServiceImpl) DeleteConfigFile(id string) error {
	// 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return fmt.Errorf("Path manager not initialized")
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 查找配置文件
	files, err := filepath.Glob(filepath.Join(configsDir, "*.json"))
	if err != nil {
		return fmt.Errorf("Failed to read config directory: %v", err)
	}

	var targetFile string
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		var fileConfig map[string]interface{}
		if err := json.Unmarshal(content, &fileConfig); err != nil {
			continue
		}

		if fileID, ok := fileConfig["id"].(string); ok && fileID == id {
			targetFile = file
			break
		}
	}

	if targetFile == "" {
		return fmt.Errorf("Config file not found for ID: %s", id)
	}

	// 删除文件
	if err := os.Remove(targetFile); err != nil {
		return fmt.Errorf("Failed to delete config file: %v", err)
	}

	return nil
}

// buildLoraPrompt 构建包含LoRA权重的提示词
func (s *generationConfigServiceImpl) buildLoraPrompt(originalPrompt string, loras []models.LoraConfig) string {
	if len(loras) == 0 {
		return originalPrompt
	}

	var loraString strings.Builder

	for _, lora := range loras {
		// 构建LoRA字符串，参考D:\GolangProject\PixivTailor的格式
		if lora.UseMask {
			// 使用遮罩模式
			if lora.LoraKey != "" {
				extendTags := strings.Join(lora.ExtendTags, ",")
				loraString.WriteString(fmt.Sprintf("(%s,%s),", lora.LoraKey, extendTags))
			}
		} else {
			// 标准LoRA模式：<lora:name:weight>tags,
			tags := strings.Join(lora.Tags, ",")
			extendTags := strings.Join(lora.ExtendTags, ",")

			// 组合标签
			var allTags strings.Builder
			if tags != "" {
				allTags.WriteString(tags)
			}
			if extendTags != "" {
				if allTags.Len() > 0 {
					allTags.WriteString(",")
				}
				allTags.WriteString(extendTags)
			}

			// 构建LoRA字符串
			if allTags.Len() > 0 {
				loraString.WriteString(fmt.Sprintf("<lora:%s:%.2f>%s,", lora.Name, lora.Weight, allTags.String()))
			} else {
				loraString.WriteString(fmt.Sprintf("<lora:%s:%.2f>,", lora.Name, lora.Weight))
			}
		}
	}

	// 将LoRA字符串添加到原始提示词前面
	if loraString.Len() > 0 {
		return loraString.String() + originalPrompt
	}

	return originalPrompt
}

// sanitizeFileName 清理文件名
func sanitizeFileName(name string) string {
	// 移除或替换非法字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
