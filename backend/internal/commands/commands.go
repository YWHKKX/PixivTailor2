package commands

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"pixiv-tailor/backend/internal/ai"
	"pixiv-tailor/backend/internal/crawler"
	"pixiv-tailor/backend/internal/grpc"
	"pixiv-tailor/backend/internal/http"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/internal/service"
	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"
	pb "pixiv-tailor/proto"

	"github.com/urfave/cli/v2"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ============================================================================
// 端口检测工具函数
// ============================================================================

// findAvailablePort 查找可用端口
func findAvailablePort(startPort int, maxAttempts int) (int, error) {
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		addr := fmt.Sprintf(":%d", port)

		// 尝试监听端口
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}

		// 如果端口被占用，继续尝试下一个
		logger.Debugf("端口 %d 被占用，尝试下一个端口", port)
	}

	return 0, fmt.Errorf("在 %d 次尝试后未找到可用端口", maxAttempts)
}

// killProcessOnPort 杀死占用指定端口的进程
func killProcessOnPort(port int) error {
	// 在Windows上使用netstat和taskkill命令
	cmd := fmt.Sprintf("netstat -ano | findstr :%d", port)
	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		return fmt.Errorf("检查端口占用失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) && strings.Contains(line, "LISTENING") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				pid := parts[len(parts)-1]
				logger.Infof("发现端口 %d 被进程 %s 占用，正在终止...", port, pid)

				// 杀死进程
				killCmd := fmt.Sprintf("taskkill /PID %s /F", pid)
				killOutput, killErr := exec.Command("cmd", "/C", killCmd).Output()
				if killErr != nil {
					logger.Warnf("终止进程失败: %v, 输出: %s", killErr, string(killOutput))
					return fmt.Errorf("终止进程失败: %v", killErr)
				}

				logger.Infof("成功终止进程 %s，端口 %d 现在可用", pid, port)
				return nil
			}
		}
	}

	return fmt.Errorf("未找到占用端口 %d 的进程", port)
}

// ensurePortAvailable 确保端口可用，如果被占用则kill进程
func ensurePortAvailable(port int) error {
	addr := fmt.Sprintf(":%d", port)

	// 尝试监听端口
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		listener.Close()
		logger.Infof("端口 %d 可用", port)
		return nil
	}

	// 端口被占用，尝试kill进程
	logger.Infof("端口 %d 被占用，尝试终止占用进程...", port)
	if err := killProcessOnPort(port); err != nil {
		return fmt.Errorf("无法释放端口 %d: %v", port, err)
	}

	// 再次尝试监听端口
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("端口 %d 仍被占用: %v", port, err)
	}

	listener.Close()
	logger.Infof("端口 %d 现在可用", port)
	return nil
}

// ============================================================================
// 服务器命令处理
// ============================================================================

// ServerAction 服务器命令处理函数
func ServerAction(ctx *cli.Context) error {
	// 获取参数
	port := ctx.String("port")
	dbPath := ctx.String("db")
	configPath := ctx.String("config")

	// 使用 PathManager 提供默认值
	pm := paths.GetPathManager()
	if dbPath == "" {
		dbPath = pm.GetDatabasePath()
	}
	if configPath == "" {
		configPath = pm.GetMainConfigPath()
	}

	// 解析端口号
	grpcPort := 50051 // 固定gRPC端口
	httpPort := 50052 // 固定HTTP端口

	if strings.HasPrefix(port, ":") {
		if p, err := strconv.Atoi(port[1:]); err == nil {
			grpcPort = p
			httpPort = p + 1 // HTTP端口 = gRPC端口 + 1
		}
	}

	// 确保gRPC端口可用
	logger.Infof("检查gRPC端口 %d 可用性...", grpcPort)
	if err := ensurePortAvailable(grpcPort); err != nil {
		return fmt.Errorf("无法使用gRPC端口 %d: %v", grpcPort, err)
	}

	// 确保HTTP端口可用
	logger.Infof("检查HTTP端口 %d 可用性...", httpPort)
	if err := ensurePortAvailable(httpPort); err != nil {
		return fmt.Errorf("无法使用HTTP端口 %d: %v", httpPort, err)
	}

	// 初始化日志系统
	logger.Infof("启动 PixivTailor 服务器...")
	logger.Infof("gRPC端口: :%d", grpcPort)
	logger.Infof("HTTP端口: :%d", httpPort)
	logger.Infof("数据库: %s", dbPath)
	logger.Infof("配置: %s", configPath)

	// 初始化配置（暂时跳过，使用默认配置）
	// if err := config.InitGlobalConfig(configPath); err != nil {
	// 	return fmt.Errorf("初始化配置失败: %v", err)
	// }

	// 初始化存储
	store, err := repository.NewSQLiteStorage(dbPath)
	if err != nil {
		return fmt.Errorf("初始化存储失败: %v", err)
	}
	defer store.Close()

	// 初始化服务
	taskService := service.NewTaskService(store)
	configService := service.NewConfigService(store)
	dataService := service.NewDataService(store)
	systemService := service.NewSystemService()

	// 创建 gRPC 服务器
	server := grpcLib.NewServer()

	// 注册服务
	pb.RegisterPixivTailorServiceServer(server, &grpc.PixivTailorServer{
		TaskService:   taskService,
		ConfigService: configService,
		DataService:   dataService,
		SystemService: systemService,
	})

	// 启用反射（用于调试）
	reflection.Register(server)

	// 启动 gRPC 服务器
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("监听gRPC端口失败: %v", err)
	}

	logger.Infof("gRPC 服务器监听端口: :%d", grpcPort)
	logger.Infof("gRPC 服务器地址: localhost:%d", grpcPort)
	logger.Info("按 Ctrl+C 停止服务器")

	// 创建 HTTP 服务器
	httpServer := http.NewHTTPServer(taskService, configService, dataService, systemService)

	// 启动 gRPC 服务器
	go func() {
		if err := server.Serve(lis); err != nil {
			logger.Errorf("gRPC 服务器启动失败: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		httpPortStr := fmt.Sprintf("%d", httpPort)
		logger.Infof("HTTP 服务器启动在端口 %s", httpPortStr)
		if err := httpServer.Start(httpPortStr); err != nil {
			logger.Errorf("HTTP 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")
	server.GracefulStop()
	logger.Info("服务器已关闭")

	return nil
}

// ============================================================================
// 爬虫命令处理
// ============================================================================

// CrawlAction 爬虫命令处理函数
func CrawlAction(ctx *cli.Context) error {
	// 获取参数
	query := ctx.String("query")
	userID := ctx.Int("user-id")
	illustID := ctx.Int("illust-id")
	order := ctx.String("order")
	mode := ctx.String("mode")
	limit := ctx.Int("limit")
	output := ctx.String("output")

	// 使用 PathManager 提供默认值
	pm := paths.GetPathManager()
	if output == "" {
		output = pm.GetImagesDir()
	}

	// 创建爬虫实例
	crawlerInstance, err := crawler.NewCrawler()
	if err != nil {
		return fmt.Errorf("创建爬虫实例失败: %v", err)
	}

	// 配置代理设置
	proxyEnabled := ctx.Bool("proxy")
	proxyURL := ctx.String("proxy-url")

	if proxyEnabled {
		logger.Infof("启用代理: %s", proxyURL)
		crawlerInstance.SetProxy(true, "http://"+proxyURL)
	} else {
		logger.Info("使用直连模式")
		crawlerInstance.SetProxy(false, "")
	}

	var images []*models.PixivImage

	// 根据参数选择爬取方式
	if query != "" {
		// 按标签爬取
		logger.Infof("开始按标签爬取: %s", query)
		images, err = crawlerInstance.CrawlByTag(query, order, mode, limit)
		if err != nil {
			return fmt.Errorf("按标签爬取失败: %v", err)
		}
	} else if userID > 0 {
		// 按用户爬取
		logger.Infof("开始按用户爬取: %d", userID)
		images, err = crawlerInstance.CrawlByUser(userID, limit)
		if err != nil {
			return fmt.Errorf("按用户爬取失败: %v", err)
		}
	} else if illustID > 0 {
		// 按插画ID爬取
		logger.Infof("开始按插画ID爬取: %d", illustID)
		image, err := crawlerInstance.CrawlByIllust(illustID)
		if err != nil {
			return fmt.Errorf("按插画ID爬取失败: %v", err)
		}
		images = []*models.PixivImage{image}
	} else {
		return fmt.Errorf("请指定查询参数、用户ID或插画ID")
	}

	logger.Infof("成功爬取 %d 张图片", len(images))

	// 保存结果
	if output != "" {
		logger.Infof("开始保存图片到: %s", output)

		// 统计变量
		var successCount, failCount int
		var failedURLs []string

		// 根据爬取方式生成目录名
		var subDir string
		if query != "" {
			subDir = fmt.Sprintf("tag_%s", query)
		} else if userID > 0 {
			subDir = fmt.Sprintf("user_%d", userID)
		} else if illustID > 0 {
			subDir = fmt.Sprintf("illust_%d", illustID)
		}

		// 创建输出目录
		finalOutput := filepath.Join(output, subDir)
		if err := os.MkdirAll(finalOutput, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}

		logger.Infof("图片将保存到目录: %s", finalOutput)

		// 下载图片
		imageCount := make(map[int64]int) // 记录每个插画ID的图片数量

		for i, image := range images {
			if image.URL == "" {
				failCount++
				failedURLs = append(failedURLs, "URL为空")
				continue
			}

			// 为每个插画ID的图片编号
			imageCount[image.ID]++
			imageIndex := imageCount[image.ID]

			// 生成文件名 - 使用插画ID，格式为 artworks_{id}_{page}.jpg
			var filename string
			if image.ID > 0 {
				filename = fmt.Sprintf("artworks_%d_p%02d.jpg", image.ID, imageIndex)
			} else {
				filename = fmt.Sprintf("image_%d.jpg", i+1)
			}

			fullPath := filepath.Join(finalOutput, filename)

			// 下载图片
			if err := crawlerInstance.DownloadImage(image.URL, fullPath, "", nil); err != nil {
				logger.Errorf("下载图片失败 %s: %v", image.URL, err)
				failCount++
				failedURLs = append(failedURLs, image.URL)
				continue
			}

			successCount++
			logger.Infof("已保存图片: %s (插画ID: %d, 第%d页)", fullPath, image.ID, imageIndex)
		}

		// 输出统计结果
		logger.Infof("=== 下载统计 ===")
		logger.Infof("成功下载: %d 张", successCount)
		logger.Infof("下载失败: %d 张", failCount)
		logger.Infof("成功率: %.2f%%", float64(successCount)/float64(len(images))*100)

		if len(failedURLs) > 0 {
			logger.Warnf("失败的URL列表:")
			for i, url := range failedURLs {
				logger.Warnf("  %d. %s", i+1, url)
			}
		}
	}

	return nil
}

// ============================================================================
// 生成命令处理
// ============================================================================

// GenerateAction 生成命令处理函数
func GenerateAction(ctx *cli.Context) error {
	// 获取参数
	model := ctx.String("model")
	prompt := ctx.String("prompt")
	negativePrompt := ctx.String("negative-prompt")
	loraSlice := ctx.StringSlice("lora")
	loraWeights := ctx.Float64Slice("lora-weight")
	batchSize := ctx.Int("batch-size")
	steps := ctx.Int("steps")
	cfgScale := ctx.Float64("cfg-scale")
	width := ctx.Int("width")
	height := ctx.Int("height")
	seed := ctx.Int64("seed")
	output := ctx.String("output")

	// 使用 PathManager 提供默认值
	pm := paths.GetPathManager()
	if output == "" {
		output = pm.GetDataPath("generated")
	}

	// 验证参数
	if model == "" || prompt == "" {
		return fmt.Errorf("模型和提示词不能为空")
	}

	// 创建生成器实例
	generator, err := ai.NewGenerator()
	if err != nil {
		return fmt.Errorf("创建生成器实例失败: %v", err)
	}

	// 构建LoRA配置
	var loras []models.LoraConfig
	for i, loraStr := range loraSlice {
		parts := strings.Split(loraStr, ":")
		if len(parts) != 2 {
			return fmt.Errorf("LoRA格式错误: %s (应为 name:weight)", loraStr)
		}

		weight := 1.0
		if i < len(loraWeights) {
			weight = loraWeights[i]
		}

		loras = append(loras, models.LoraConfig{
			Name:   parts[0],
			Weight: weight,
		})
	}

	// 构建生成请求
	request := &models.GenerateRequest{
		Model:          model,
		Prompt:         prompt,
		NegativePrompt: negativePrompt,
		Loras:          loras,
		BatchSize:      batchSize,
		Steps:          steps,
		CFGScale:       cfgScale,
		Width:          width,
		Height:         height,
		Seed:           int(seed),
		SavePath:       output,
	}

	// 执行生成
	logger.Infof("开始生成图像，模型: %s", model)
	images, err := generator.GenerateImages(request)
	if err != nil {
		return fmt.Errorf("生成失败: %v", err)
	}

	logger.Infof("成功生成 %d 张图像", len(images))

	return nil
}

// ============================================================================
// 训练命令处理
// ============================================================================

// TrainAction 训练命令处理函数
func TrainAction(ctx *cli.Context) error {
	// 获取参数
	modelName := ctx.String("model-name")
	trainingData := ctx.String("training-data")
	pretrainedModel := ctx.String("pretrained-model")
	epochs := ctx.Int("epochs")
	batchSize := ctx.Int("batch-size")
	learningRate := ctx.Float64("learning-rate")
	output := ctx.String("output")

	// 使用 PathManager 提供默认值
	pm := paths.GetPathManager()
	if output == "" {
		output = pm.GetModelsDir()
	}

	// 验证参数
	if modelName == "" || trainingData == "" {
		return fmt.Errorf("模型名称和训练数据不能为空")
	}

	// 创建训练器实例
	trainer, err := ai.NewTrainer()
	if err != nil {
		return fmt.Errorf("创建训练器实例失败: %v", err)
	}

	// 构建训练请求
	request := &models.TrainRequest{
		Name:           modelName,
		InputDir:       trainingData,
		PretrainedPath: pretrainedModel,
		Epochs:         epochs,
		BatchSize:      batchSize,
		LearningRate:   learningRate,
		OutputDir:      output,
	}

	// 执行训练
	logger.Infof("开始训练模型: %s", modelName)
	trainedModel, err := trainer.TrainModel(request)
	if err != nil {
		return fmt.Errorf("训练失败: %v", err)
	}

	if trainedModel != nil {
		logger.Infof("训练完成，模型: %s", trainedModel.Name)
	} else {
		logger.Infof("训练完成，模型: %s", modelName)
	}

	return nil
}

// ============================================================================
// 标签命令处理
// ============================================================================

// TagAction 标签命令处理函数
func TagAction(ctx *cli.Context) error {
	// 获取参数
	inputDir := ctx.String("input")
	outputDir := ctx.String("output")
	analyzer := ctx.String("analyzer")
	saveType := ctx.String("save-type")

	// 验证参数
	if inputDir == "" {
		return fmt.Errorf("输入目录不能为空")
	}

	// 使用 PathManager 提供默认值
	pm := paths.GetPathManager()
	if outputDir == "" {
		outputDir = pm.GetTagsDir()
	}

	// 创建标签器实例
	tagger, err := ai.NewTagger()
	if err != nil {
		return fmt.Errorf("创建标签器实例失败: %v", err)
	}

	// 构建标签请求
	request := &models.TagRequest{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Analyzer:  analyzer,
		SaveType:  saveType,
	}

	// 执行标签生成
	logger.Infof("开始生成标签，输入目录: %s", inputDir)
	taggedImages, err := tagger.GenerateTags(request)
	if err != nil {
		return fmt.Errorf("标签生成失败: %v", err)
	}

	logger.Infof("成功处理 %d 张图像", len(taggedImages))

	return nil
}

// ============================================================================
// 分类命令处理
// ============================================================================

// ClassifyAction 分类命令处理函数
func ClassifyAction(ctx *cli.Context) error {
	// 获取参数
	inputFile := ctx.String("input")
	outputFile := ctx.String("output")
	categories := ctx.StringSlice("categories")

	// 验证参数
	if inputFile == "" {
		return fmt.Errorf("输入文件不能为空")
	}

	// 使用 PathManager 提供默认值
	pm := paths.GetPathManager()
	if outputFile == "" {
		outputFile = pm.GetDataPath("classified.json")
	}

	// 创建分类器实例
	classifier, err := ai.NewClassifier()
	if err != nil {
		return fmt.Errorf("创建分类器实例失败: %v", err)
	}

	// 构建分类请求
	request := &models.ClassifyRequest{
		Input:  categories,
		Output: outputFile,
	}

	// 执行分类
	logger.Infof("开始分类标签，输入文件: %s", inputFile)
	globalTags, err := classifier.ClassifyTags(request)
	if err != nil {
		return fmt.Errorf("分类失败: %v", err)
	}

	logger.Infof("分类完成，输出文件: %s", outputFile)

	// 保存结果
	err = classifier.SaveClassification(globalTags, outputFile)
	if err != nil {
		return fmt.Errorf("保存分类结果失败: %v", err)
	}

	return nil
}

// ============================================================================
// 工具函数
// ============================================================================

// ParseIntSlice 解析整数切片
func ParseIntSlice(slice []string) ([]int, error) {
	var result []int
	for _, s := range slice {
		val, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("无效的整数: %s", s)
		}
		result = append(result, val)
	}
	return result, nil
}

// ParseFloatSlice 解析浮点数切片
func ParseFloatSlice(slice []string) ([]float64, error) {
	var result []float64
	for _, s := range slice {
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的浮点数: %s", s)
		}
		result = append(result, val)
	}
	return result, nil
}
