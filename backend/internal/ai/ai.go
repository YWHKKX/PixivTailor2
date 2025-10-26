package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"pixiv-tailor/backend/internal/config"
	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"
)

// ============================================================================
// SD WebUI 请求和响应结构体
// ============================================================================

// SDWebUIRequest SD WebUI API请求结构
type SDWebUIRequest struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Seed           int     `json:"seed"`
	SamplerName    string  `json:"sampler_name"`
	BatchSize      int     `json:"batch_size"`
	EnableHR       bool    `json:"enable_hr"`
}

// SDWebUIResponse SD WebUI API响应结构
type SDWebUIResponse struct {
	Images []string `json:"images"`
	Info   string   `json:"info"`
}

// SDWebUIInfo SD WebUI 生成信息
type SDWebUIInfo struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Steps          int     `json:"steps"`
	CFGScale       float64 `json:"cfg_scale"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	Seed           int     `json:"seed"`
	SamplerName    string  `json:"sampler_name"`
	BatchSize      int     `json:"batch_size"`
}

// ============================================================================
// 分类器接口和实现
// ============================================================================

// Classifier 分类器接口
type Classifier interface {
	// ClassifyTags 分类标签
	ClassifyTags(request *models.ClassifyRequest) (models.GlobalTags, error)

	// SaveClassification 保存分类结果
	SaveClassification(globalTags models.GlobalTags, outputPath string) error

	// GetCategories 获取标签分类
	GetCategories() (models.TagCategories, error)

	// ClassifySingleTag 分类单个标签
	ClassifySingleTag(tag string) (string, error)
}

// OpenAIClassifier OpenAI分类器
type OpenAIClassifier struct {
	apiKeys []string
	timeout int
}

// NewClassifier 创建新的分类器实例
func NewClassifier() (Classifier, error) {
	// 从配置文件加载AI配置
	aiConfig := config.GetAIConfig()

	return &OpenAIClassifier{
		apiKeys: aiConfig.OpenAI.APIKeys,
		timeout: aiConfig.OpenAI.Timeout,
	}, nil
}

// ClassifyTags 分类标签
func (c *OpenAIClassifier) ClassifyTags(request *models.ClassifyRequest) (models.GlobalTags, error) {
	// TODO: 实现标签分类逻辑
	return nil, nil
}

// SaveClassification 保存分类结果
func (c *OpenAIClassifier) SaveClassification(globalTags models.GlobalTags, outputPath string) error {
	// TODO: 实现分类结果保存逻辑
	return nil
}

// GetCategories 获取标签分类
func (c *OpenAIClassifier) GetCategories() (models.TagCategories, error) {
	// TODO: 实现标签分类获取逻辑
	return nil, nil
}

// ClassifySingleTag 分类单个标签
func (c *OpenAIClassifier) ClassifySingleTag(tag string) (string, error) {
	// TODO: 实现单个标签分类逻辑
	return "", nil
}

// ============================================================================
// 生成器接口和实现
// ============================================================================

// Generator AI生成器接口
type Generator interface {
	// GenerateImages 生成图像
	GenerateImages(request *models.GenerateRequest) ([]*models.GeneratedImage, error)

	// SaveImage 保存图像
	SaveImage(image *models.GeneratedImage, filepath string) error

	// GetModels 获取可用模型列表
	GetModels() ([]string, error)

	// GetLoras 获取可用LoRA模型列表
	GetLoras() ([]string, error)
}

// SDWebUIGenerator Stable Diffusion WebUI生成器
type SDWebUIGenerator struct {
	url     string
	apiKey  string
	timeout int
}

// NewGenerator 创建新的生成器实例
func NewGenerator() (Generator, error) {
	// 从配置文件加载AI配置
	aiConfig := config.GetAIConfig()

	return &SDWebUIGenerator{
		url:     aiConfig.SDWebUI.URL,
		apiKey:  "", // 暂时不使用API Key
		timeout: aiConfig.SDWebUI.Timeout,
	}, nil
}

// GenerateImages 生成图像
func (g *SDWebUIGenerator) GenerateImages(request *models.GenerateRequest) ([]*models.GeneratedImage, error) {
	// 构造SD WebUI请求
	sdRequest := SDWebUIRequest{
		Prompt:         request.Prompt,
		NegativePrompt: request.NegativePrompt,
		Steps:          request.Steps,
		CFGScale:       request.CFGScale,
		Width:          request.Width,
		Height:         request.Height,
		Seed:           request.Seed,
		SamplerName:    request.Sampler,
		BatchSize:      request.BatchSize,
		EnableHR:       request.EnableHR,
	}

	// 序列化请求
	requestBody, err := json.Marshal(sdRequest)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送HTTP请求到SD WebUI
	apiURL := g.url + "/sdapi/v1/txt2img"
	client := &http.Client{Timeout: time.Duration(g.timeout) * time.Second}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if g.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求SD WebUI失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SD WebUI返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var sdResponse SDWebUIResponse
	if err := json.NewDecoder(resp.Body).Decode(&sdResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 转换响应为GeneratedImage
	images := make([]*models.GeneratedImage, len(sdResponse.Images))
	for i, imageBase64 := range sdResponse.Images {
		// 转换LoraConfig为字符串切片
		loraNames := make([]string, len(request.Loras))
		for j, lora := range request.Loras {
			loraNames[j] = lora.Name
		}

		images[i] = &models.GeneratedImage{
			BaseModel: models.BaseModel{
				ID:        int64(i + 1),
				CreatedAt: time.Now(),
			},
			Prompt:         request.Prompt,
			NegativePrompt: request.NegativePrompt,
			Width:          request.Width,
			Height:         request.Height,
			Steps:          request.Steps,
			CFGScale:       request.CFGScale,
			Seed:           int64(request.Seed + i),
			Model:          request.Model,
			Loras:          loraNames,
			ImageURL:       imageBase64, // 存储base64编码的图像数据
		}
	}

	return images, nil
}

// SaveImage 保存图像
func (g *SDWebUIGenerator) SaveImage(image *models.GeneratedImage, savePath string) error {
	if image.ImageURL == "" {
		return fmt.Errorf("图像数据为空")
	}

	// 解码base64图像数据
	imageData, err := base64.StdEncoding.DecodeString(image.ImageURL)
	if err != nil {
		return fmt.Errorf("解码base64图像失败: %v", err)
	}

	// 使用路径管理器确保目录存在
	pathManager := paths.GetPathManager()
	if pathManager != nil {
		// 确保目录存在
		dir := filepath.Dir(savePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %v", err)
		}
	}

	// 写入文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	_, err = file.Write(imageData)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// GetModels 获取可用模型列表
func (g *SDWebUIGenerator) GetModels() ([]string, error) {
	// 发送HTTP请求到SD WebUI获取模型列表
	client := &http.Client{Timeout: time.Duration(g.timeout) * time.Second}
	req, err := http.NewRequest("GET", g.url+"/sdapi/v1/sd-models", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	if g.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求SD WebUI失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SD WebUI返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var models []struct {
		Title     string `json:"title"`
		ModelName string `json:"model_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取模型名称
	modelNames := make([]string, len(models))
	for i, model := range models {
		if model.Title != "" {
			modelNames[i] = model.Title
		} else {
			modelNames[i] = model.ModelName
		}
	}

	return modelNames, nil
}

// GetLoras 获取可用LoRA模型列表
func (g *SDWebUIGenerator) GetLoras() ([]string, error) {
	// 发送HTTP请求到SD WebUI获取LoRA模型列表
	client := &http.Client{Timeout: time.Duration(g.timeout) * time.Second}
	req, err := http.NewRequest("GET", g.url+"/sdapi/v1/loras", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	if g.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求SD WebUI失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SD WebUI返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var loras []struct {
		Name  string `json:"name"`
		Alias string `json:"alias"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loras); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取LoRA名称
	loraNames := make([]string, len(loras))
	for i, lora := range loras {
		if lora.Name != "" {
			loraNames[i] = lora.Name
		} else {
			loraNames[i] = lora.Alias
		}
	}

	return loraNames, nil
}

// ============================================================================
// 标签器接口和实现
// ============================================================================

// Tagger 标签器接口
type Tagger interface {
	// GenerateTags 生成标签
	GenerateTags(request *models.TagRequest) ([]*models.TaggedImage, error)

	// SaveTags 保存标签
	SaveTags(taggedImage *models.TaggedImage, outputDir, saveType string) error

	// GetAvailableAnalyzers 获取可用的分析器
	GetAvailableAnalyzers() ([]string, error)

	// AnalyzeImage 分析单张图像
	AnalyzeImage(imagePath string, analyzer string) (*models.TaggedImage, error)
}

// WD14Tagger WD14Tagger实现
type WD14Tagger struct {
	url     string
	timeout int
}

// NewTagger 创建新的标签器实例
func NewTagger() (Tagger, error) {
	// 从配置文件加载AI配置
	aiConfig := config.GetAIConfig()

	return &WD14Tagger{
		url:     aiConfig.SDWebUI.URL, // 使用SD WebUI的URL
		timeout: aiConfig.SDWebUI.Timeout,
	}, nil
}

// GenerateTags 生成标签
func (t *WD14Tagger) GenerateTags(request *models.TagRequest) ([]*models.TaggedImage, error) {
	// TODO: 实现标签生成逻辑
	return nil, nil
}

// SaveTags 保存标签
func (t *WD14Tagger) SaveTags(taggedImage *models.TaggedImage, outputDir, saveType string) error {
	// TODO: 实现标签保存逻辑
	return nil
}

// GetAvailableAnalyzers 获取可用的分析器
func (t *WD14Tagger) GetAvailableAnalyzers() ([]string, error) {
	// TODO: 实现分析器列表获取逻辑
	return []string{"wd14tagger", "deepdanbooru"}, nil
}

// AnalyzeImage 分析单张图像
func (t *WD14Tagger) AnalyzeImage(imagePath string, analyzer string) (*models.TaggedImage, error) {
	// TODO: 实现单张图像分析逻辑
	return nil, nil
}

// ============================================================================
// 训练器接口和实现
// ============================================================================

// Trainer 训练器接口
type Trainer interface {
	// TrainModel 训练模型
	TrainModel(request *models.TrainRequest) (*models.TrainedModel, error)

	// GetTrainingStatus 获取训练状态
	GetTrainingStatus(modelName string) (*models.ProgressInfo, error)

	// StopTraining 停止训练
	StopTraining(modelName string) error

	// GetPretrainedModels 获取预训练模型列表
	GetPretrainedModels() ([]string, error)
}

// KohyaSSTrainer Kohya-ss训练器
type KohyaSSTrainer struct {
	url     string
	timeout int
}

// NewTrainer 创建新的训练器实例
func NewTrainer() (Trainer, error) {
	// 从配置文件加载AI配置
	aiConfig := config.GetAIConfig()

	return &KohyaSSTrainer{
		url:     aiConfig.KohyaSS.URL,
		timeout: aiConfig.KohyaSS.Timeout,
	}, nil
}

// TrainModel 训练模型
func (t *KohyaSSTrainer) TrainModel(request *models.TrainRequest) (*models.TrainedModel, error) {
	// TODO: 实现模型训练逻辑
	return nil, nil
}

// GetTrainingStatus 获取训练状态
func (t *KohyaSSTrainer) GetTrainingStatus(modelName string) (*models.ProgressInfo, error) {
	// TODO: 实现训练状态获取逻辑
	return nil, nil
}

// StopTraining 停止训练
func (t *KohyaSSTrainer) StopTraining(modelName string) error {
	// TODO: 实现训练停止逻辑
	return nil
}

// GetPretrainedModels 获取预训练模型列表
func (t *KohyaSSTrainer) GetPretrainedModels() ([]string, error) {
	// TODO: 实现预训练模型列表获取逻辑
	return nil, nil
}

// ============================================================================
// AI 模块统一管理
// ============================================================================

// AIManager AI模块管理器
type AIManager struct {
	classifier Classifier
	generator  Generator
	tagger     Tagger
	trainer    Trainer
}

// NewAIManager 创建AI管理器
func NewAIManager() (*AIManager, error) {
	classifier, err := NewClassifier()
	if err != nil {
		return nil, err
	}

	generator, err := NewGenerator()
	if err != nil {
		return nil, err
	}

	tagger, err := NewTagger()
	if err != nil {
		return nil, err
	}

	trainer, err := NewTrainer()
	if err != nil {
		return nil, err
	}

	return &AIManager{
		classifier: classifier,
		generator:  generator,
		tagger:     tagger,
		trainer:    trainer,
	}, nil
}

// GetClassifier 获取分类器
func (m *AIManager) GetClassifier() Classifier {
	return m.classifier
}

// GetGenerator 获取生成器
func (m *AIManager) GetGenerator() Generator {
	return m.generator
}

// GetTagger 获取标签器
func (m *AIManager) GetTagger() Tagger {
	return m.tagger
}

// GetTrainer 获取训练器
func (m *AIManager) GetTrainer() Trainer {
	return m.trainer
}

// NewAIService 创建AI服务实例
func NewAIService() (*AIManager, error) {
	return NewAIManager()
}
