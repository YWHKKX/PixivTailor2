package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pixiv-tailor/backend/pkg/models"
)

// AIServiceConfig AI服务配置
type AIServiceConfig struct {
	SDWebUIURL     string
	WD14TaggerURL  string
	KohyaSSURL     string
	OpenAIAPIKeys  []string
	DefaultTimeout int
}

// AIService AI服务接口
type AIService interface {
	// 图像生成
	GenerateImages(request *models.GenerateRequest) ([]*models.GeneratedImage, error)

	// 标签生成
	GenerateTags(request *models.TagRequest) ([]*models.TaggedImage, error)

	// 标签分类
	ClassifyTags(request *models.ClassifyRequest) (models.GlobalTags, error)

	// 模型训练
	TrainModel(request *models.TrainRequest) (*models.TrainedModel, error)

	// 获取训练状态
	GetTrainingStatus(modelName string) (*models.ProgressInfo, error)

	// 停止训练
	StopTraining(modelName string) error
}

// aiServiceImpl AI服务实现
type aiServiceImpl struct {
	config     AIServiceConfig
	httpClient HTTPClient
}

// NewAIService 创建AI服务
func NewAIService(config AIServiceConfig, httpClient HTTPClient) AIService {
	return &aiServiceImpl{
		config:     config,
		httpClient: httpClient,
	}
}

// GenerateImages 生成图像
func (a *aiServiceImpl) GenerateImages(request *models.GenerateRequest) ([]*models.GeneratedImage, error) {
	// 构造SD WebUI请求
	sdRequest := map[string]interface{}{
		"prompt":          request.Prompt,
		"negative_prompt": request.NegativePrompt,
		"steps":           request.Steps,
		"cfg_scale":       request.CFGScale,
		"width":           request.Width,
		"height":          request.Height,
		"seed":            request.Seed,
		"sampler_name":    request.Sampler,
		"batch_size":      request.BatchSize,
		"n_iter":          request.BatchCount,
		"enable_hr":       request.EnableHR,
		"save_images":     true,
		"save_grid":       false,
	}

	// 序列化请求
	requestBody, err := json.Marshal(sdRequest)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送请求到SD WebUI
	req, err := a.httpClient.CreateRequest("POST", a.config.SDWebUIURL+"/sdapi/v1/txt2img", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置超时
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.DoWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("请求SD WebUI失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := a.httpClient.ReadResponseBody(resp)
		return nil, fmt.Errorf("SD WebUI返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var sdResponse struct {
		Images []string `json:"images"`
		Info   string   `json:"info"`
	}

	body, err := a.httpClient.ReadResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if err := json.Unmarshal(body, &sdResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 转换响应为GeneratedImage
	images := make([]*models.GeneratedImage, len(sdResponse.Images))
	for i, imageBase64 := range sdResponse.Images {
		images[i] = &models.GeneratedImage{
			BaseModel: models.BaseModel{
				ID:        int64(time.Now().UnixNano() + int64(i)),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Prompt:         request.Prompt,
			NegativePrompt: request.NegativePrompt,
			Model:          request.Model,
			ImageURL:       fmt.Sprintf("data:image/png;base64,%s", imageBase64),
			Width:          request.Width,
			Height:         request.Height,
			Seed:           int64(request.Seed),
			CFGScale:       request.CFGScale,
			Steps:          request.Steps,
			Sampler:        request.Sampler,
		}
	}

	return images, nil
}

// GenerateTags 生成标签
func (a *aiServiceImpl) GenerateTags(request *models.TagRequest) ([]*models.TaggedImage, error) {
	// TODO: 实现WD14Tagger集成
	return nil, fmt.Errorf("标签生成功能暂未实现")
}

// ClassifyTags 分类标签
func (a *aiServiceImpl) ClassifyTags(request *models.ClassifyRequest) (models.GlobalTags, error) {
	// TODO: 实现OpenAI分类集成
	return nil, fmt.Errorf("标签分类功能暂未实现")
}

// TrainModel 训练模型
func (a *aiServiceImpl) TrainModel(request *models.TrainRequest) (*models.TrainedModel, error) {
	// TODO: 实现Kohya-ss训练集成
	return nil, fmt.Errorf("模型训练功能暂未实现")
}

// GetTrainingStatus 获取训练状态
func (a *aiServiceImpl) GetTrainingStatus(modelName string) (*models.ProgressInfo, error) {
	// TODO: 实现训练状态获取
	return nil, fmt.Errorf("训练状态获取功能暂未实现")
}

// StopTraining 停止训练
func (a *aiServiceImpl) StopTraining(modelName string) error {
	// TODO: 实现训练停止
	return fmt.Errorf("训练停止功能暂未实现")
}

// DefaultAIServiceConfig 默认AI服务配置
func DefaultAIServiceConfig() AIServiceConfig {
	return AIServiceConfig{
		SDWebUIURL:     "http://127.0.0.1:7860",
		WD14TaggerURL:  "http://127.0.0.1:7861",
		KohyaSSURL:     "http://127.0.0.1:7862",
		OpenAIAPIKeys:  []string{},
		DefaultTimeout: 60,
	}
}
