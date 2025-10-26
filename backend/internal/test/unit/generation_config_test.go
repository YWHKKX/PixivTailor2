package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pixiv-tailor/backend/pkg/models"
)

func TestGenerationConfigFileSystem(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 测试配置
	config := &models.GenerationConfig{
		ID:          "test-config",
		Name:        "测试配置",
		Category:    "动漫",
		Description: "测试用配置",
		Model:       "test_model.safetensors",
		Loras: []models.LoraConfig{
			{
				Path:   "test_lora.safetensors",
				Weight: 0.8,
			},
		},
		Steps:    20,
		CFGScale: 7,
		Width:    512,
		Height:   512,
		Sampler:  "DPM++ 2M Karras",
		EnableHR: false,
	}

	// 1. 测试保存配置到文件
	t.Run("SaveConfigToFile", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "test-config.json")

		configData, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(configPath, configData, 0644)
		assert.NoError(t, err)

		// 验证文件是否创建
		assert.FileExists(t, configPath)
	})

	// 2. 测试从文件读取配置
	t.Run("LoadConfigFromFile", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "test-config.json")

		// 读取文件
		fileContent, err := os.ReadFile(configPath)
		assert.NoError(t, err)

		// 解析JSON
		var loadedConfig models.GenerationConfig
		err = json.Unmarshal(fileContent, &loadedConfig)
		assert.NoError(t, err)

		// 验证配置内容
		assert.Equal(t, config.ID, loadedConfig.ID)
		assert.Equal(t, config.Name, loadedConfig.Name)
		assert.Equal(t, config.Category, loadedConfig.Category)
		assert.Equal(t, config.Model, loadedConfig.Model)
		assert.Equal(t, config.Steps, loadedConfig.Steps)
		assert.Equal(t, config.CFGScale, loadedConfig.CFGScale)
		assert.Equal(t, config.Width, loadedConfig.Width)
		assert.Equal(t, config.Height, loadedConfig.Height)
		assert.Equal(t, config.Sampler, loadedConfig.Sampler)
		assert.Equal(t, config.EnableHR, loadedConfig.EnableHR)

		// 验证LoRA配置
		assert.Len(t, loadedConfig.Loras, 1)
		assert.Equal(t, config.Loras[0].Path, loadedConfig.Loras[0].Path)
		assert.Equal(t, config.Loras[0].Weight, loadedConfig.Loras[0].Weight)
	})

	// 3. 测试更新配置
	t.Run("UpdateConfig", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "test-config.json")

		// 更新配置
		updatedConfig := *config
		updatedConfig.Name = "更新的测试配置"
		updatedConfig.Description = "更新后的描述"
		updatedConfig.Steps = 25
		updatedConfig.CFGScale = 8

		// 保存更新后的配置
		configData, err := json.MarshalIndent(&updatedConfig, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(configPath, configData, 0644)
		assert.NoError(t, err)

		// 验证更新
		fileContent, err := os.ReadFile(configPath)
		assert.NoError(t, err)

		var loadedConfig models.GenerationConfig
		err = json.Unmarshal(fileContent, &loadedConfig)
		assert.NoError(t, err)

		assert.Equal(t, "更新的测试配置", loadedConfig.Name)
		assert.Equal(t, "更新后的描述", loadedConfig.Description)
		assert.Equal(t, 25, loadedConfig.Steps)
		assert.Equal(t, float64(8), loadedConfig.CFGScale)
	})

	// 4. 测试删除配置
	t.Run("DeleteConfig", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "test-config.json")

		// 验证文件存在
		assert.FileExists(t, configPath)

		// 删除文件
		err := os.Remove(configPath)
		assert.NoError(t, err)

		// 验证文件被删除
		assert.NoFileExists(t, configPath)
	})
}

func TestGenerationConfigValidation(t *testing.T) {
	// 测试配置验证
	t.Run("ValidConfig", func(t *testing.T) {
		config := &models.GenerationConfig{
			ID:       "valid-config",
			Name:     "有效配置",
			Category: "动漫",
			Model:    "test_model.safetensors",
			Steps:    20,
			CFGScale: 7,
			Width:    512,
			Height:   512,
			Sampler:  "DPM++ 2M Karras",
			EnableHR: false,
		}

		// 验证基本字段
		assert.NotEmpty(t, config.ID)
		assert.NotEmpty(t, config.Name)
		assert.NotEmpty(t, config.Model)
		assert.Greater(t, config.Steps, 0)
		assert.Greater(t, config.CFGScale, 0.0)
		assert.Greater(t, config.Width, 0)
		assert.Greater(t, config.Height, 0)
	})

	t.Run("ConfigWithLoRA", func(t *testing.T) {
		config := &models.GenerationConfig{
			ID:       "lora-config",
			Name:     "LoRA配置",
			Category: "动漫",
			Model:    "test_model.safetensors",
			Loras: []models.LoraConfig{
				{
					Path:   "anime_lora.safetensors",
					Weight: 0.8,
				},
				{
					Path:   "style_lora.safetensors",
					Weight: 0.6,
				},
			},
			Steps:    20,
			CFGScale: 7,
			Width:    512,
			Height:   512,
			Sampler:  "DPM++ 2M Karras",
			EnableHR: false,
		}

		// 验证LoRA配置
		assert.Len(t, config.Loras, 2)
		assert.Equal(t, "anime_lora.safetensors", config.Loras[0].Path)
		assert.Equal(t, 0.8, config.Loras[0].Weight)
		assert.Equal(t, "style_lora.safetensors", config.Loras[1].Path)
		assert.Equal(t, 0.6, config.Loras[1].Weight)
	})
}

func TestGenerationConfigJSON(t *testing.T) {
	// 测试JSON序列化和反序列化
	t.Run("JSONSerialization", func(t *testing.T) {
		config := &models.GenerationConfig{
			ID:          "json-test-config",
			Name:        "JSON测试配置",
			Category:    "测试",
			Description: "用于测试JSON序列化",
			Model:       "test_model.safetensors",
			Loras: []models.LoraConfig{
				{
					Path:   "test_lora.safetensors",
					Weight: 0.8,
				},
			},
			Steps:    20,
			CFGScale: 7,
			Width:    512,
			Height:   512,
			Sampler:  "DPM++ 2M Karras",
			EnableHR: false,
		}

		// 序列化为JSON
		jsonData, err := json.MarshalIndent(config, "", "  ")
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// 反序列化
		var loadedConfig models.GenerationConfig
		err = json.Unmarshal(jsonData, &loadedConfig)
		assert.NoError(t, err)

		// 验证数据完整性
		assert.Equal(t, config.ID, loadedConfig.ID)
		assert.Equal(t, config.Name, loadedConfig.Name)
		assert.Equal(t, config.Category, loadedConfig.Category)
		assert.Equal(t, config.Description, loadedConfig.Description)
		assert.Equal(t, config.Model, loadedConfig.Model)
		assert.Equal(t, config.Steps, loadedConfig.Steps)
		assert.Equal(t, config.CFGScale, loadedConfig.CFGScale)
		assert.Equal(t, config.Width, loadedConfig.Width)
		assert.Equal(t, config.Height, loadedConfig.Height)
		assert.Equal(t, config.Sampler, loadedConfig.Sampler)
		assert.Equal(t, config.EnableHR, loadedConfig.EnableHR)

		// 验证LoRA配置
		assert.Len(t, loadedConfig.Loras, 1)
		assert.Equal(t, config.Loras[0].Path, loadedConfig.Loras[0].Path)
		assert.Equal(t, config.Loras[0].Weight, loadedConfig.Loras[0].Weight)
	})
}
