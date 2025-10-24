package service

import (
	"fmt"
	"pixiv-tailor/backend/internal/repository"
)

// DataService 数据服务接口
type DataService interface {
	GetCrawlResults(page, pageSize int32, tags []string, author string) ([]*repository.CrawlResult, int, error)
	GetGeneratedImages(page, pageSize int32, model string) ([]*repository.GeneratedImage, int, error)
	GetTrainedModels(page, pageSize int32, modelType string) ([]*repository.TrainedModel, int, error)
	AddCrawlResult(result *repository.CrawlResult) error
	AddGeneratedImage(image *repository.GeneratedImage) error
	AddTrainedModel(model *repository.TrainedModel) error
}

// dataServiceImpl 数据服务实现
type dataServiceImpl struct {
	storage repository.Storage
}

// NewDataService 创建数据服务实例
func NewDataService(storage repository.Storage) DataService {
	return &dataServiceImpl{
		storage: storage,
	}
}

// GetCrawlResults 获取爬取结果
func (s *dataServiceImpl) GetCrawlResults(page, pageSize int32, tags []string, author string) ([]*repository.CrawlResult, int, error) {
	if pageSize <= 0 {
		pageSize = 20 // 默认限制
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制
	}
	if page <= 0 {
		page = 1
	}

	offset := int((page - 1) * pageSize)
	results, err := s.storage.GetCrawlResults(int(pageSize), offset)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := s.storage.CountCrawlResults()
	if err != nil {
		return results, 0, err
	}

	return results, total, nil
}

// GetGeneratedImages 获取生成的图像
func (s *dataServiceImpl) GetGeneratedImages(page, pageSize int32, model string) ([]*repository.GeneratedImage, int, error) {
	if pageSize <= 0 {
		pageSize = 20 // 默认限制
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制
	}
	if page <= 0 {
		page = 1
	}

	offset := int((page - 1) * pageSize)
	images, err := s.storage.GetGeneratedImages(int(pageSize), offset)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := s.storage.CountGeneratedImages()
	if err != nil {
		return images, 0, err
	}

	return images, total, nil
}

// GetTrainedModels 获取训练的模型
func (s *dataServiceImpl) GetTrainedModels(page, pageSize int32, modelType string) ([]*repository.TrainedModel, int, error) {
	if pageSize <= 0 {
		pageSize = 20 // 默认限制
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制
	}
	if page <= 0 {
		page = 1
	}

	offset := int((page - 1) * pageSize)
	models, err := s.storage.GetTrainedModels(int(pageSize), offset)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	total, err := s.storage.CountTrainedModels()
	if err != nil {
		return models, 0, err
	}

	return models, total, nil
}

// AddCrawlResult 添加爬取结果
func (s *dataServiceImpl) AddCrawlResult(result *repository.CrawlResult) error {
	if result == nil {
		return fmt.Errorf("爬取结果不能为空")
	}

	if result.ID == "" {
		return fmt.Errorf("爬取结果ID不能为空")
	}

	if result.URL == "" {
		return fmt.Errorf("爬取结果URL不能为空")
	}

	// TODO: 实现具体的数据库插入逻辑
	// 这里需要扩展Storage接口来支持插入操作

	return nil
}

// AddGeneratedImage 添加生成的图像
func (s *dataServiceImpl) AddGeneratedImage(image *repository.GeneratedImage) error {
	if image == nil {
		return fmt.Errorf("生成的图像不能为空")
	}

	if image.ID == "" {
		return fmt.Errorf("生成的图像ID不能为空")
	}

	if image.ImageURL == "" {
		return fmt.Errorf("生成的图像URL不能为空")
	}

	// TODO: 实现具体的数据库插入逻辑
	// 这里需要扩展Storage接口来支持插入操作

	return nil
}

// AddTrainedModel 添加训练的模型
func (s *dataServiceImpl) AddTrainedModel(model *repository.TrainedModel) error {
	if model == nil {
		return fmt.Errorf("训练的模型不能为空")
	}

	if model.ID == "" {
		return fmt.Errorf("训练的模型ID不能为空")
	}

	if model.Name == "" {
		return fmt.Errorf("训练的模型名称不能为空")
	}

	if model.Path == "" {
		return fmt.Errorf("训练的模型路径不能为空")
	}

	// TODO: 实现具体的数据库插入逻辑
	// 这里需要扩展Storage接口来支持插入操作

	return nil
}
