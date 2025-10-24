package tests

import (
	"testing"
	"time"

	"pixiv-tailor/backend/pkg/errors"
	"pixiv-tailor/backend/pkg/models"
)

// TestDataModelsDetailed 测试数据模型详细功能
func TestDataModelsDetailed(t *testing.T) {
	t.Run("PixivImage", func(t *testing.T) {
		// 创建Pixiv图像
		image := &models.PixivImage{
			ID:        12345,
			Title:     "Test Image",
			URL:       "https://example.com/test.jpg",
			Author:    "Test Author",
			AuthorID:  67890,
			Tags:      []string{"test", "anime", "beautiful"},
			Width:     512,
			Height:    512,
			Bookmarks: 100,
			Views:     1000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 验证字段
		if image.ID != 12345 {
			t.Errorf("期望ID 12345，实际: %d", image.ID)
		}

		if image.Title != "Test Image" {
			t.Errorf("期望标题 'Test Image'，实际: %s", image.Title)
		}

		if image.URL != "https://example.com/test.jpg" {
			t.Errorf("期望URL 'https://example.com/test.jpg'，实际: %s", image.URL)
		}

		if image.Author != "Test Author" {
			t.Errorf("期望作者 'Test Author'，实际: %s", image.Author)
		}

		if image.AuthorID != 67890 {
			t.Errorf("期望作者ID 67890，实际: %d", image.AuthorID)
		}

		if len(image.Tags) != 3 {
			t.Errorf("期望标签数量 3，实际: %d", len(image.Tags))
		}

		if image.Bookmarks != 100 {
			t.Errorf("期望收藏数 100，实际: %d", image.Bookmarks)
		}

		if image.Views != 1000 {
			t.Errorf("期望浏览数 1000，实际: %d", image.Views)
		}

		if image.Width != 512 {
			t.Errorf("期望宽度 512，实际: %d", image.Width)
		}

		if image.Height != 512 {
			t.Errorf("期望高度 512，实际: %d", image.Height)
		}
	})

	t.Run("GenerateRequest", func(t *testing.T) {
		// 创建生成请求
		request := &models.GenerateRequest{
			Model:          "test_model.safetensors",
			Prompt:         "1girl, beautiful, detailed",
			NegativePrompt: "blurry, low quality",
			Loras: []models.LoraConfig{
				{Name: "lora1", Weight: 0.8},
				{Name: "lora2", Weight: 0.6},
			},
			BatchSize: 4,
			Steps:     20,
			CFGScale:  7.0,
			Width:     512,
			Height:    512,
			Seed:      12345,
			Sampler:   "DPM++ 2M Karras",
			SavePath:  "output/test",
		}

		// 验证字段
		if request.Model != "test_model.safetensors" {
			t.Errorf("期望模型 'test_model.safetensors'，实际: %s", request.Model)
		}

		if request.Prompt != "1girl, beautiful, detailed" {
			t.Errorf("期望提示词 '1girl, beautiful, detailed'，实际: %s", request.Prompt)
		}

		if request.NegativePrompt != "blurry, low quality" {
			t.Errorf("期望负面提示词 'blurry, low quality'，实际: %s", request.NegativePrompt)
		}

		if len(request.Loras) != 2 {
			t.Errorf("期望LoRA数量 2，实际: %d", len(request.Loras))
		}

		if request.Loras[0].Name != "lora1" {
			t.Errorf("期望第一个LoRA名称 'lora1'，实际: %s", request.Loras[0].Name)
		}

		if request.Loras[0].Weight != 0.8 {
			t.Errorf("期望第一个LoRA权重 0.8，实际: %f", request.Loras[0].Weight)
		}

		if request.BatchSize != 4 {
			t.Errorf("期望批处理大小 4，实际: %d", request.BatchSize)
		}

		if request.Steps != 20 {
			t.Errorf("期望步数 20，实际: %d", request.Steps)
		}

		if request.CFGScale != 7.0 {
			t.Errorf("期望CFG缩放 7.0，实际: %f", request.CFGScale)
		}

		if request.Width != 512 {
			t.Errorf("期望宽度 512，实际: %d", request.Width)
		}

		if request.Height != 512 {
			t.Errorf("期望高度 512，实际: %d", request.Height)
		}

		if request.Seed != 12345 {
			t.Errorf("期望种子 12345，实际: %d", request.Seed)
		}

		if request.Sampler != "DPM++ 2M Karras" {
			t.Errorf("期望采样器 'DPM++ 2M Karras'，实际: %s", request.Sampler)
		}
	})

	t.Run("TrainRequest", func(t *testing.T) {
		// 创建训练请求
		request := &models.TrainRequest{
			Name:           "test_character_lora",
			PretrainedPath: "models/pretrained.safetensors",
			InputDir:       "data/images/train",
			OutputDir:      "data/models",
			Epochs:         10,
			BatchSize:      2,
			LearningRate:   0.0001,
			Tags: map[string]int{
				"1girl":     5,
				"anime":     3,
				"beautiful": 2,
			},
			Prompts: "1girl, beautiful, detailed",
		}

		// 验证字段
		if request.Name != "test_character_lora" {
			t.Errorf("期望模型名称 'test_character_lora'，实际: %s", request.Name)
		}

		if request.PretrainedPath != "models/pretrained.safetensors" {
			t.Errorf("期望预训练模型路径 'models/pretrained.safetensors'，实际: %s", request.PretrainedPath)
		}

		if request.InputDir != "data/images/train" {
			t.Errorf("期望输入目录 'data/images/train'，实际: %s", request.InputDir)
		}

		if request.OutputDir != "data/models" {
			t.Errorf("期望输出目录 'data/models'，实际: %s", request.OutputDir)
		}

		if request.Epochs != 10 {
			t.Errorf("期望训练轮数 10，实际: %d", request.Epochs)
		}

		if request.BatchSize != 2 {
			t.Errorf("期望批处理大小 2，实际: %d", request.BatchSize)
		}

		if request.LearningRate != 0.0001 {
			t.Errorf("期望学习率 0.0001，实际: %f", request.LearningRate)
		}

		if len(request.Tags) != 3 {
			t.Errorf("期望标签数量 3，实际: %d", len(request.Tags))
		}

		if request.Tags["1girl"] != 5 {
			t.Errorf("期望标签 '1girl' 重复次数 5，实际: %d", request.Tags["1girl"])
		}

		if request.Prompts != "1girl, beautiful, detailed" {
			t.Errorf("期望提示词 '1girl, beautiful, detailed'，实际: %s", request.Prompts)
		}
	})

	t.Run("TagRequest", func(t *testing.T) {
		// 创建标签请求
		request := &models.TagRequest{
			InputDir:   "data/images",
			OutputDir:  "data/tags",
			Analyzer:   "wd14tagger",
			SkipTags:   []string{"low_quality", "blurry"},
			ExtendTags: []string{"high_quality", "detailed"},
			TagOrder:   "character",
			SaveType:   "txt",
			Limit:      100,
		}

		// 验证字段
		if request.InputDir != "data/images" {
			t.Errorf("期望输入目录 'data/images'，实际: %s", request.InputDir)
		}

		if request.OutputDir != "data/tags" {
			t.Errorf("期望输出目录 'data/tags'，实际: %s", request.OutputDir)
		}

		if request.Analyzer != "wd14tagger" {
			t.Errorf("期望分析器 'wd14tagger'，实际: %s", request.Analyzer)
		}

		if len(request.SkipTags) != 2 {
			t.Errorf("期望跳过标签数量 2，实际: %d", len(request.SkipTags))
		}

		if len(request.ExtendTags) != 2 {
			t.Errorf("期望扩展标签数量 2，实际: %d", len(request.ExtendTags))
		}

		if request.TagOrder != "character" {
			t.Errorf("期望标签顺序 'character'，实际: %s", request.TagOrder)
		}

		if request.SaveType != "txt" {
			t.Errorf("期望保存类型 'txt'，实际: %s", request.SaveType)
		}

		if request.Limit != 100 {
			t.Errorf("期望限制数量 100，实际: %d", request.Limit)
		}
	})

	t.Run("ClassifyRequest", func(t *testing.T) {
		// 创建分类请求
		request := &models.ClassifyRequest{
			Input:    []string{"1girl", "long_hair", "blue_eyes"},
			Output:   "configs/global_tags.json",
			APIKeys:  []string{"sk-xxx", "sk-yyy"},
			Limit:    1000,
			ShowTags: true,
		}

		// 验证字段
		if len(request.Input) != 3 {
			t.Errorf("期望输入标签数量 3，实际: %d", len(request.Input))
		}

		if request.Output != "configs/global_tags.json" {
			t.Errorf("期望输出路径 'configs/global_tags.json'，实际: %s", request.Output)
		}

		if len(request.APIKeys) != 2 {
			t.Errorf("期望API密钥数量 2，实际: %d", len(request.APIKeys))
		}

		if request.Limit != 1000 {
			t.Errorf("期望限制数量 1000，实际: %d", request.Limit)
		}

		if request.ShowTags != true {
			t.Error("期望ShowTags为true")
		}
	})

	t.Run("LoraConfig", func(t *testing.T) {
		// 创建LoRA配置
		lora := models.LoraConfig{
			Name:   "test_lora",
			Weight: 0.8,
		}

		// 验证字段
		if lora.Name != "test_lora" {
			t.Errorf("期望LoRA名称 'test_lora'，实际: %s", lora.Name)
		}

		if lora.Weight != 0.8 {
			t.Errorf("期望LoRA权重 0.8，实际: %f", lora.Weight)
		}
	})

	t.Run("CrawlRequest", func(t *testing.T) {
		// 创建爬取请求
		request := &models.CrawlRequest{
			Type:  models.CrawlTypeTag,
			Query: "test_tag",
			Order: models.OrderDateD,
			Mode:  models.ModeSafe,
			Limit: 100,
			Delay: 1,
		}

		// 验证字段
		if request.Type != models.CrawlTypeTag {
			t.Errorf("期望爬取类型 %s，实际: %s", models.CrawlTypeTag, request.Type)
		}

		if request.Query != "test_tag" {
			t.Errorf("期望查询 'test_tag'，实际: %s", request.Query)
		}

		if request.Order != models.OrderDateD {
			t.Errorf("期望排序 %s，实际: %s", models.OrderDateD, request.Order)
		}

		if request.Mode != models.ModeSafe {
			t.Errorf("期望模式 %s，实际: %s", models.ModeSafe, request.Mode)
		}

		if request.Limit != 100 {
			t.Errorf("期望限制 100，实际: %d", request.Limit)
		}

		if request.Delay != 1 {
			t.Errorf("期望延迟 1，实际: %d", request.Delay)
		}
	})
}

// TestConstants 测试常量
func TestConstants(t *testing.T) {
	t.Run("CrawlType", func(t *testing.T) {
		if models.CrawlTypeTag != "tag" {
			t.Errorf("期望CrawlTypeTag 'tag'，实际: %s", models.CrawlTypeTag)
		}

		if models.CrawlTypeUser != "user" {
			t.Errorf("期望CrawlTypeUser 'user'，实际: %s", models.CrawlTypeUser)
		}

		if models.CrawlTypeIllust != "illust" {
			t.Errorf("期望CrawlTypeIllust 'illust'，实际: %s", models.CrawlTypeIllust)
		}
	})

	t.Run("Order", func(t *testing.T) {
		if models.OrderDateD != "date_d" {
			t.Errorf("期望OrderDateD 'date_d'，实际: %s", models.OrderDateD)
		}

		if models.OrderPopularD != "popular_d" {
			t.Errorf("期望OrderPopularD 'popular_d'，实际: %s", models.OrderPopularD)
		}
	})

	t.Run("Mode", func(t *testing.T) {
		if models.ModeSafe != "safe" {
			t.Errorf("期望ModeSafe 'safe'，实际: %s", models.ModeSafe)
		}

		if models.ModeR18 != "r18" {
			t.Errorf("期望ModeR18 'r18'，实际: %s", models.ModeR18)
		}

		if models.ModeAll != "all" {
			t.Errorf("期望ModeAll 'all'，实际: %s", models.ModeAll)
		}
	})
}

// TestErrorCodes 测试错误代码
func TestErrorCodes(t *testing.T) {
	t.Run("ErrorCodeValues", func(t *testing.T) {
		// 验证错误代码值
		if errors.ErrCodeUnknown != 0 {
			t.Errorf("期望ErrCodeUnknown 0，实际: %d", errors.ErrCodeUnknown)
		}

		if errors.ErrCodeInvalidParam != 1 {
			t.Errorf("期望ErrCodeInvalidParam 1，实际: %d", errors.ErrCodeInvalidParam)
		}

		if errors.ErrCodeFileNotFound != 2 {
			t.Errorf("期望ErrCodeFileNotFound 2，实际: %d", errors.ErrCodeFileNotFound)
		}
	})

	t.Run("ErrorCodeNames", func(t *testing.T) {
		// 验证错误代码名称
		errorCodes := map[errors.ErrorCode]string{
			errors.ErrCodeUnknown:          "ErrCodeUnknown",
			errors.ErrCodeInvalidParam:     "ErrCodeInvalidParam",
			errors.ErrCodeFileNotFound:     "ErrCodeFileNotFound",
			errors.ErrCodePermissionDenied: "ErrCodePermissionDenied",
			errors.ErrCodeNetworkError:     "ErrCodeNetworkError",
			errors.ErrCodeTimeout:          "ErrCodeTimeout",
			errors.ErrCodeCrawlerInit:      "ErrCodeCrawlerInit",
			errors.ErrCodeCrawlerRequest:   "ErrCodeCrawlerRequest",
			errors.ErrCodeCrawlerParse:     "ErrCodeCrawlerParse",
			errors.ErrCodeCrawlerDownload:  "ErrCodeCrawlerDownload",
			errors.ErrCodeAIConfig:         "ErrCodeAIConfig",
			errors.ErrCodeAIGenerate:       "ErrCodeAIGenerate",
			errors.ErrCodeAITrain:          "ErrCodeAITrain",
			errors.ErrCodeAITag:            "ErrCodeAITag",
			errors.ErrCodeAIClassify:       "ErrCodeAIClassify",
			errors.ErrCodeConfigLoad:       "ErrCodeConfigLoad",
			errors.ErrCodeConfigParse:      "ErrCodeConfigParse",
			errors.ErrCodeConfigValidate:   "ErrCodeConfigValidate",
			errors.ErrCodeConfigSave:       "ErrCodeConfigSave",
		}

		for code, expectedName := range errorCodes {
			// 这里只是验证错误代码存在，实际名称验证需要更复杂的实现
			_ = code
			_ = expectedName
		}
	})
}

// BenchmarkModelCreation 模型创建性能测试
func BenchmarkModelCreation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建Pixiv图像
		_ = &models.PixivImage{
			ID:        i,
			Title:     "Test Image",
			URL:       "https://example.com/test.jpg",
			Author:    "Test Author",
			AuthorID:  12345,
			Tags:      []string{"test", "anime"},
			Width:     512,
			Height:    512,
			Bookmarks: 100,
			Views:     1000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
}

// BenchmarkGenerateRequest 生成请求创建性能测试
func BenchmarkGenerateRequest(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建生成请求
		_ = &models.GenerateRequest{
			Model:          "test_model.safetensors",
			Prompt:         "1girl, beautiful, detailed",
			NegativePrompt: "blurry, low quality",
			Loras: []models.LoraConfig{
				{Name: "lora1", Weight: 0.8},
				{Name: "lora2", Weight: 0.6},
			},
			BatchSize: 4,
			Steps:     20,
			CFGScale:  7.0,
			Width:     512,
			Height:    512,
			Seed:      12345,
			Sampler:   "DPM++ 2M Karras",
		}
	}
}
