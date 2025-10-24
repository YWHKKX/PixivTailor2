package tests

import (
	"path/filepath"
	"testing"

	"pixiv-tailor/backend/internal/ai"
	"pixiv-tailor/backend/internal/config"
	"pixiv-tailor/backend/internal/crawler"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/models"
)

// TestConfigSystem 测试配置系统
func TestConfigSystem(t *testing.T) {
	t.Run("InitGlobalConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("初始化配置失败: %v", err)
		}

		// 验证配置管理器已初始化
		if config.GetGlobalConfig() == nil {
			t.Error("配置管理器未初始化")
		}
	})

	t.Run("LoadDefaultConfig", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("初始化配置失败: %v", err)
		}

		// 验证默认配置已加载
		aiConfig, err := config.GetModuleConfig("ai")
		if err != nil {
			t.Fatalf("获取AI模块配置失败: %v", err)
		}

		if aiConfig.GetName() != "ai" {
			t.Errorf("期望AI模块名称 'ai'，实际: %s", aiConfig.GetName())
		}
	})

	t.Run("SetAndGetConfigValue", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("初始化配置失败: %v", err)
		}

		// 设置配置值
		testValue := "test_value"
		err = config.SetConfigValue("test.key", testValue)
		if err != nil {
			t.Fatalf("设置配置值失败: %v", err)
		}

		// 获取配置值
		value, err := config.GetConfigValue("test.key")
		if err != nil {
			t.Fatalf("获取配置值失败: %v", err)
		}

		if value != testValue {
			t.Errorf("期望配置值 '%s'，实际: %v", testValue, value)
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// 创建临时配置文件
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		// 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("初始化配置失败: %v", err)
		}

		// 验证配置
		err = config.ValidateConfig()
		if err != nil {
			t.Errorf("配置验证失败: %v", err)
		}
	})
}

// TestAIModules 测试AI模块
func TestAIModules(t *testing.T) {
	t.Run("Generator", func(t *testing.T) {
		// 创建生成器
		generator, err := ai.NewGenerator()
		if err != nil {
			t.Fatalf("创建生成器失败: %v", err)
		}

		// 创建生成请求
		request := &models.GenerateRequest{
			Model:     "test_model",
			Prompt:    "test prompt",
			BatchSize: 1,
			Steps:     10,
			Width:     512,
			Height:    512,
		}

		// 测试生成图像
		images, err := generator.GenerateImages(request)
		if err != nil {
			t.Errorf("生成图像失败: %v", err)
		}

		if len(images) != request.BatchSize {
			t.Errorf("期望生成 %d 张图像，实际: %d", request.BatchSize, len(images))
		}
	})

	t.Run("Trainer", func(t *testing.T) {
		// 创建训练器
		trainer, err := ai.NewTrainer()
		if err != nil {
			t.Fatalf("创建训练器失败: %v", err)
		}

		// 创建训练请求
		request := &models.TrainRequest{
			Name:           "test_model",
			PretrainedPath: "test_pretrained.safetensors",
			InputDir:       "test_input",
			OutputDir:      "test_output",
			Epochs:         5,
			BatchSize:      2,
			LearningRate:   0.0001,
		}

		// 测试训练模型
		result, err := trainer.TrainModel(request)
		if err != nil {
			t.Errorf("训练模型失败: %v", err)
		}

		if result == nil {
			t.Error("训练结果为空")
		}
	})

	t.Run("Tagger", func(t *testing.T) {
		// 创建标签器
		tagger, err := ai.NewTagger()
		if err != nil {
			t.Fatalf("创建标签器失败: %v", err)
		}

		// 创建标签请求
		request := &models.TagRequest{
			InputDir:  "test_input",
			OutputDir: "test_output",
			Analyzer:  "wd14tagger",
			Limit:     10,
		}

		// 测试生成标签
		taggedImages, err := tagger.GenerateTags(request)
		if err != nil {
			t.Errorf("生成标签失败: %v", err)
		}

		if len(taggedImages) == 0 {
			t.Error("没有生成标签图像")
		}
	})

	t.Run("Classifier", func(t *testing.T) {
		// 创建分类器
		classifier, err := ai.NewClassifier()
		if err != nil {
			t.Fatalf("创建分类器失败: %v", err)
		}

		// 创建分类请求
		request := &models.ClassifyRequest{
			Input:   []string{"1girl", "anime", "beautiful"},
			Output:  "test_output.json",
			APIKeys: []string{"test_key"},
			Limit:   100,
		}

		// 测试分类标签
		categories, err := classifier.ClassifyTags(request)
		if err != nil {
			t.Errorf("分类标签失败: %v", err)
		}

		if len(categories) == 0 {
			t.Error("没有生成分类结果")
		}
	})
}

// TestCrawlerModule 测试爬虫模块
func TestCrawlerModule(t *testing.T) {
	t.Run("CreateCrawler", func(t *testing.T) {
		// 创建爬虫
		crawler, err := crawler.NewCrawler()
		if err != nil {
			t.Fatalf("创建爬虫失败: %v", err)
		}

		if crawler == nil {
			t.Error("爬虫实例为空")
		}
	})

	t.Run("CrawlByTag", func(t *testing.T) {
		// 创建爬虫
		crawler, err := crawler.NewCrawler()
		if err != nil {
			t.Fatalf("创建爬虫失败: %v", err)
		}

		// 测试按标签爬取
		images, err := crawler.CrawlByTag("test", "date_d", "safe", 10)
		if err != nil {
			t.Errorf("按标签爬取失败: %v", err)
		}

		if len(images) == 0 {
			t.Error("没有爬取到图像")
		}
	})

	t.Run("CrawlByUser", func(t *testing.T) {
		// 创建爬虫
		crawler, err := crawler.NewCrawler()
		if err != nil {
			t.Fatalf("创建爬虫失败: %v", err)
		}

		// 测试按用户爬取
		images, err := crawler.CrawlByUser(12345, 10)
		if err != nil {
			t.Errorf("按用户爬取失败: %v", err)
		}

		if len(images) == 0 {
			t.Error("没有爬取到图像")
		}
	})

	t.Run("CrawlByIllust", func(t *testing.T) {
		// 创建爬虫
		crawler, err := crawler.NewCrawler()
		if err != nil {
			t.Fatalf("创建爬虫失败: %v", err)
		}

		// 测试按插画ID爬取
		image, err := crawler.CrawlByIllust(12345)
		if err != nil {
			t.Errorf("按插画ID爬取失败: %v", err)
		}

		if image == nil {
			t.Error("没有爬取到图像")
		}
	})

	t.Run("DownloadImage", func(t *testing.T) {
		// 创建爬虫
		crawler, err := crawler.NewCrawler()
		if err != nil {
			t.Fatalf("创建爬虫失败: %v", err)
		}

		// 创建临时目录
		tempDir := t.TempDir()
		savePath := filepath.Join(tempDir, "test_image.jpg")

		// 测试下载图像
		err = crawler.DownloadImage("https://example.com/test.jpg", savePath)
		if err != nil {
			t.Errorf("下载图像失败: %v", err)
		}
	})
}

// TestDataModels 测试数据模型
func TestDataModels(t *testing.T) {
	t.Run("PixivImage", func(t *testing.T) {
		// 创建Pixiv图像
		image := &models.PixivImage{
			ID:        12345,
			Title:     "Test Image",
			URL:       "https://example.com/test.jpg",
			Author:    "Test Author",
			AuthorID:  67890,
			Tags:      []string{"test", "anime"},
			Width:     512,
			Height:    512,
			Bookmarks: 100,
			Views:     1000,
		}

		if image.ID != 12345 {
			t.Errorf("期望ID 12345，实际: %d", image.ID)
		}

		if image.Title != "Test Image" {
			t.Errorf("期望标题 'Test Image'，实际: %s", image.Title)
		}
	})

	t.Run("GenerateRequest", func(t *testing.T) {
		// 创建生成请求
		request := &models.GenerateRequest{
			Model:          "test_model",
			Prompt:         "test prompt",
			NegativePrompt: "test negative",
			BatchSize:      4,
			Steps:          20,
			CFGScale:       7.0,
			Width:          512,
			Height:         512,
			Seed:           12345,
			Sampler:        "DPM++ 2M Karras",
		}

		if request.Model != "test_model" {
			t.Errorf("期望模型 'test_model'，实际: %s", request.Model)
		}

		if request.BatchSize != 4 {
			t.Errorf("期望批处理大小 4，实际: %d", request.BatchSize)
		}
	})

	t.Run("TrainRequest", func(t *testing.T) {
		// 创建训练请求
		request := &models.TrainRequest{
			Name:           "test_model",
			PretrainedPath: "test_pretrained.safetensors",
			InputDir:       "test_input",
			OutputDir:      "test_output",
			Epochs:         10,
			BatchSize:      2,
			LearningRate:   0.0001,
			Tags: map[string]int{
				"1girl": 5,
				"anime": 3,
			},
		}

		if request.Name != "test_model" {
			t.Errorf("期望模型名称 'test_model'，实际: %s", request.Name)
		}

		if request.Epochs != 10 {
			t.Errorf("期望训练轮数 10，实际: %d", request.Epochs)
		}
	})
}

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	t.Run("FullWorkflow", func(t *testing.T) {
		// 创建临时目录
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test_config.json")

		// 1. 初始化配置系统
		err := config.InitGlobalConfig(configPath)
		if err != nil {
			t.Fatalf("初始化配置失败: %v", err)
		}

		// 2. 初始化日志系统
		logger.Init(false)

		// 3. 测试配置访问
		_, err = config.GetModuleConfig("ai")
		if err != nil {
			t.Fatalf("获取AI模块配置失败: %v", err)
		}

		// 4. 测试AI模块
		generator, err := ai.NewGenerator()
		if err != nil {
			t.Fatalf("创建生成器失败: %v", err)
		}

		// 5. 测试爬虫模块
		crawler, err := crawler.NewCrawler()
		if err != nil {
			t.Fatalf("创建爬虫失败: %v", err)
		}

		// 6. 测试完整流程
		logger.Info("开始集成测试")

		// 模拟爬取图像
		images, err := crawler.CrawlByTag("test", "date_d", "safe", 5)
		if err != nil {
			t.Errorf("爬取图像失败: %v", err)
		}

		// 模拟生成图像
		request := &models.GenerateRequest{
			Model:     "test_model",
			Prompt:    "test prompt",
			BatchSize: 2,
			Steps:     10,
			Width:     512,
			Height:    512,
		}

		generatedImages, err := generator.GenerateImages(request)
		if err != nil {
			t.Errorf("生成图像失败: %v", err)
		}

		logger.Infof("爬取到 %d 张图像，生成了 %d 张图像", len(images), len(generatedImages))

		// 7. 验证配置保存
		legacyConfig := config.GetConfig()
		err = config.SaveConfig(legacyConfig, configPath)
		if err != nil {
			t.Errorf("保存配置失败: %v", err)
		}

		logger.Info("集成测试完成")
	})
}

// BenchmarkConfigAccess 配置访问性能测试
func BenchmarkConfigAccess(b *testing.B) {
	// 创建临时配置文件
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "bench_config.json")

	// 初始化配置系统
	err := config.InitGlobalConfig(configPath)
	if err != nil {
		b.Fatalf("初始化配置失败: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := config.GetConfigValue("modules")
		if err != nil {
			b.Errorf("获取配置值失败: %v", err)
		}
	}
}
