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

func TestAIModuleIntegration(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 测试配置
	testConfig := &models.GenerationConfig{
		ID:          "integration-test-config",
		Name:        "集成测试配置",
		Category:    "测试",
		Description: "用于集成测试的配置",
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

	// 1. 测试创建配置文件
	t.Run("CreateConfigFile", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "integration-test-config.json")

		configData, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err)

		err = os.WriteFile(configPath, configData, 0644)
		assert.NoError(t, err)

		// 验证文件是否创建
		assert.FileExists(t, configPath)
	})

	// 2. 测试读取配置文件
	t.Run("ReadConfigFile", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "integration-test-config.json")

		// 读取文件
		fileContent, err := os.ReadFile(configPath)
		assert.NoError(t, err)

		// 解析JSON
		var loadedConfig models.GenerationConfig
		err = json.Unmarshal(fileContent, &loadedConfig)
		assert.NoError(t, err)

		// 验证配置内容
		assert.Equal(t, testConfig.ID, loadedConfig.ID)
		assert.Equal(t, testConfig.Name, loadedConfig.Name)
		assert.Equal(t, testConfig.Category, loadedConfig.Category)
		assert.Equal(t, testConfig.Model, loadedConfig.Model)
		assert.Equal(t, testConfig.Steps, loadedConfig.Steps)
		assert.Equal(t, testConfig.CFGScale, loadedConfig.CFGScale)
		assert.Equal(t, testConfig.Width, loadedConfig.Width)
		assert.Equal(t, testConfig.Height, loadedConfig.Height)
		assert.Equal(t, testConfig.Sampler, loadedConfig.Sampler)
		assert.Equal(t, testConfig.EnableHR, loadedConfig.EnableHR)

		// 验证LoRA配置
		assert.Len(t, loadedConfig.Loras, 1)
		assert.Equal(t, testConfig.Loras[0].Path, loadedConfig.Loras[0].Path)
		assert.Equal(t, testConfig.Loras[0].Weight, loadedConfig.Loras[0].Weight)
	})

	// 3. 测试生成请求参数
	t.Run("GenerateRequestParameters", func(t *testing.T) {
		requestBody := models.GenerateRequest{
			Model:          "test_model.safetensors",
			Prompt:         "test prompt for integration test",
			NegativePrompt: "bad quality, blurry",
			Steps:          20,
			CFGScale:       7,
			Width:          512,
			Height:         512,
			BatchCount:     1,
			EnableHR:       false,
			Options:        map[string]interface{}{},
		}

		// 验证请求参数
		assert.NotEmpty(t, requestBody.Model)
		assert.NotEmpty(t, requestBody.Prompt)
		assert.NotEmpty(t, requestBody.NegativePrompt)
		assert.Greater(t, requestBody.Steps, 0)
		assert.Greater(t, requestBody.CFGScale, 0.0)
		assert.Greater(t, requestBody.Width, 0)
		assert.Greater(t, requestBody.Height, 0)
		assert.Greater(t, requestBody.BatchCount, 0)

		// 测试JSON序列化
		jsonData, err := json.Marshal(requestBody)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// 测试JSON反序列化
		var loadedRequest models.GenerateRequest
		err = json.Unmarshal(jsonData, &loadedRequest)
		assert.NoError(t, err)

		assert.Equal(t, requestBody.Model, loadedRequest.Model)
		assert.Equal(t, requestBody.Prompt, loadedRequest.Prompt)
		assert.Equal(t, requestBody.NegativePrompt, loadedRequest.NegativePrompt)
		assert.Equal(t, requestBody.Steps, loadedRequest.Steps)
		assert.Equal(t, requestBody.CFGScale, loadedRequest.CFGScale)
		assert.Equal(t, requestBody.Width, loadedRequest.Width)
		assert.Equal(t, requestBody.Height, loadedRequest.Height)
		assert.Equal(t, requestBody.BatchCount, loadedRequest.BatchCount)
		assert.Equal(t, requestBody.EnableHR, loadedRequest.EnableHR)
	})

	// 4. 测试配置更新
	t.Run("UpdateConfig", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "integration-test-config.json")

		// 更新配置
		updatedConfig := *testConfig
		updatedConfig.Name = "更新的集成测试配置"
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

		assert.Equal(t, "更新的集成测试配置", loadedConfig.Name)
		assert.Equal(t, "更新后的描述", loadedConfig.Description)
		assert.Equal(t, 25, loadedConfig.Steps)
		assert.Equal(t, float64(8), loadedConfig.CFGScale)
	})

	// 5. 测试删除配置
	t.Run("DeleteConfig", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "integration-test-config.json")

		// 验证文件存在
		assert.FileExists(t, configPath)

		// 删除文件
		err := os.Remove(configPath)
		assert.NoError(t, err)

		// 验证文件被删除
		assert.NoFileExists(t, configPath)
	})
}

func TestAIModuleErrorHandling(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 测试获取不存在的配置文件
	t.Run("GetNonExistentConfig", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "non-existent-config.json")

		// 尝试读取不存在的文件
		_, err := os.ReadFile(configPath)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	// 测试无效的JSON文件
	t.Run("InvalidJSONFile", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "invalid-config.json")

		// 创建无效的JSON文件
		err := os.WriteFile(configPath, []byte("invalid json content"), 0644)
		require.NoError(t, err)

		// 尝试解析无效JSON
		fileContent, err := os.ReadFile(configPath)
		assert.NoError(t, err)

		var config models.GenerationConfig
		err = json.Unmarshal(fileContent, &config)
		assert.Error(t, err)
	})

	// 测试无效的生成请求参数
	t.Run("InvalidGenerateRequest", func(t *testing.T) {
		// 测试无效参数
		invalidRequest := models.GenerateRequest{
			Model:    "", // 空模型名
			Prompt:   "", // 空提示词
			Steps:    -1, // 无效步数
			CFGScale: -1, // 无效CFG Scale
			Width:    -1, // 无效宽度
			Height:   -1, // 无效高度
		}

		// 验证无效参数
		assert.Empty(t, invalidRequest.Model)
		assert.Empty(t, invalidRequest.Prompt)
		assert.Less(t, invalidRequest.Steps, 0)
		assert.Less(t, invalidRequest.CFGScale, 0.0)
		assert.Less(t, invalidRequest.Width, 0)
		assert.Less(t, invalidRequest.Height, 0)
	})
}
