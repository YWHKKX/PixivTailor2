package tagger

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pixiv-tailor/backend/internal/config"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"
)

// WD14Tagger WD14 Tagger 服务（支持 WD14-Tagger 和 DeepBooru）
type WD14Tagger struct {
	baseURL   string
	timeout   int
	client    *http.Client
	threshold float64
	model     string // 模型名称
	analyzer  string // 分析器类型：wd14tagger 或 deepbooru
}

// NewWD14Tagger 创建新的 WD14 Tagger 实例
func NewWD14Tagger() (*WD14Tagger, error) {
	return NewWD14TaggerWithModel("")
}

// NewWD14TaggerWithModel 创建新的 WD14 Tagger 实例（支持指定模型）
func NewWD14TaggerWithModel(model string) (*WD14Tagger, error) {
	aiConfig := config.GetAIConfig()

	// 如果没有指定模型，使用默认模型
	if model == "" {
		model = "wd14-convnext-v2"
	}

	return &WD14Tagger{
		baseURL:  aiConfig.SDWebUI.URL,
		timeout:  aiConfig.SDWebUI.Timeout,
		analyzer: "wd14tagger", // 默认使用 WD14-Tagger
		client: &http.Client{
			Timeout: time.Duration(aiConfig.SDWebUI.Timeout) * time.Second,
		},
		threshold: 0.35,
		model:     model,
	}, nil
}

// TagRequest WD14 Tagger 请求结构
type TagRequest struct {
	Image      string  `json:"image"` // Base64 编码的图片
	Model      string  `json:"model"`
	Threshold  float64 `json:"threshold"`
	BatchCount int     `json:"batch_count"`
}

// TagResponse WD14 Tagger 响应结构
type TagResponse struct {
	Tags string `json:"caption"` // 标签字符串，逗号分隔
}

// InterrogateRequest 完整的 WD14 Tagger API 请求
type InterrogateRequest struct {
	Image      string  `json:"image"`       // Base64 编码的图片
	Model      string  `json:"model"`       // 模型名称，如 "wd14-vit-v2"
	Threshold  float64 `json:"threshold"`   // 置信度阈值
	BatchCount int     `json:"batch_count"` // 批处理数量
}

// InterrogateResponse WD14 Tagger API 响应的各种可能结构
// 根据不同的 WD14 Tagger 扩展版本，响应格式可能不同
type InterrogateResponse struct {
	Caption interface{} `json:"caption"` // 可能是字符串、对象或数组
}

// GenerateTagsCallback 进度更新回调函数类型
type GenerateTagsCallback func(current int, total int)

// GenerateTagsLogCallback 日志回调函数类型
type GenerateTagsLogCallback func(level string, message string)

// GenerateTags 为目录中的图片生成标签
func (t *WD14Tagger) GenerateTags(request *models.TagRequest) error {
	return t.GenerateTagsWithCallback(request, nil, nil)
}

// GenerateTagsWithCallback 为目录中的图片生成标签（带进度回调和日志回调）
func (t *WD14Tagger) GenerateTagsWithCallback(request *models.TagRequest, callback GenerateTagsCallback, logCallback GenerateTagsLogCallback) error {
	// 如果请求中指定了分析器，使用请求中的分析器
	if request.Analyzer != "" {
		t.analyzer = request.Analyzer
		logger.Infof("使用分析器: %s", t.analyzer)
		if logCallback != nil {
			logCallback("info", fmt.Sprintf("使用分析器: %s", t.analyzer))
		}
	}

	// 如果请求中指定了模型，使用请求中的模型
	if request.Model != "" {
		t.model = request.Model
		logger.Infof("使用指定模型: %s", t.model)
	}

	// 获取所有输入目录
	inputDirs := request.GetInputDirs()
	if len(inputDirs) == 0 {
		errMsg := "输入目录不能为空"
		if logCallback != nil {
			logCallback("error", errMsg)
		}
		return fmt.Errorf(errMsg)
	}

	logger.Infof("开始生成标签: 输入目录数量=%d, output_dir=%s", len(inputDirs), request.OutputDir)
	if logCallback != nil {
		logCallback("info", fmt.Sprintf("开始生成标签: 输入目录数量=%d, 输出目录=%s", len(inputDirs), request.OutputDir))
	}

	// 将相对路径转换为绝对路径
	pathManager := paths.GetPathManager()

	// 收集所有目录的图片文件
	var allImageFiles []string
	var dirLimits map[string]int // 记录每个目录已使用的图片数量

	for dirIndex, inputDir := range inputDirs {
		logger.Infof("处理输入目录 %d/%d: %s", dirIndex+1, len(inputDirs), inputDir)
		if logCallback != nil {
			logCallback("info", fmt.Sprintf("处理输入目录 %d/%d: %s", dirIndex+1, len(inputDirs), inputDir))
		}

		// 解析路径
		// 输入目录应该是相对于 images/ 目录的，需要添加前缀
		var resolvedDir string
		if pathManager != nil {
			// 如果已经有前缀，直接解析；否则添加 images/ 前缀
			if strings.HasPrefix(inputDir, "images/") || strings.HasPrefix(inputDir, "tags/") || strings.HasPrefix(inputDir, "cache/") {
				resolvedDir = pathManager.ResolvePath(inputDir)
			} else {
				// 默认认为是 images/ 下的目录
				resolvedDir = pathManager.ResolvePath("images/" + inputDir)
			}
		} else {
			resolvedDir = inputDir
		}

		logger.Infof("输入目录路径: %s (原始: %s)", resolvedDir, inputDir)
		if logCallback != nil {
			logCallback("info", fmt.Sprintf("输入目录路径: %s", resolvedDir))
		}

		// 检查输入目录是否存在
		if _, err := os.Stat(resolvedDir); os.IsNotExist(err) {
			warnMsg := fmt.Sprintf("输入目录不存在，跳过: %s", inputDir)
			logger.Warnf(warnMsg)
			if logCallback != nil {
				logCallback("warning", warnMsg)
			}
			continue
		}

		logger.Infof("输入目录验证通过: %s", resolvedDir)

		// 获取当前目录的图片文件（不在这里应用 limit，limit 稍后统一应用）
		dirImageFiles, err := getImageFiles(resolvedDir, 0) // 0 表示不限制
		if err != nil {
			logger.Warnf("获取图片文件失败，跳过目录 %s: %v", inputDir, err)
			continue
		}

		logger.Infof("目录 %s 中找到 %d 张图片", inputDir, len(dirImageFiles))
		if logCallback != nil {
			logCallback("info", fmt.Sprintf("目录 %s 中找到 %d 张图片", inputDir, len(dirImageFiles)))
		}
		allImageFiles = append(allImageFiles, dirImageFiles...)

		if dirLimits == nil {
			dirLimits = make(map[string]int)
		}
		dirLimits[inputDir] = len(dirImageFiles)
	}

	logger.Infof("所有输入目录共找到 %d 张图片", len(allImageFiles))
	if logCallback != nil {
		logCallback("info", fmt.Sprintf("所有输入目录共找到 %d 张图片", len(allImageFiles)))
	}

	// 如果有图片数量限制，应用到所有文件
	imageFiles := allImageFiles
	if request.Limit > 0 && len(imageFiles) > request.Limit {
		imageFiles = imageFiles[:request.Limit]
		logger.Infof("应用图片数量限制，只处理前 %d 张", request.Limit)
		if logCallback != nil {
			logCallback("info", fmt.Sprintf("应用图片数量限制，只处理前 %d 张", request.Limit))
		}
	}

	// 转换输出目录路径
	// 确保使用正确的路径前缀
	requestedOutputDir := request.OutputDir
	if requestedOutputDir == "" {
		// 如果输出目录为空，使用 tags/ 根目录
		requestedOutputDir = "tags/"
	} else if !strings.HasPrefix(requestedOutputDir, "tags/") && !strings.HasPrefix(requestedOutputDir, "images/") && !strings.HasPrefix(requestedOutputDir, "cache/") {
		// 如果不是已有前缀，默认使用 tags/ 前缀
		requestedOutputDir = "tags/" + requestedOutputDir
	}

	// 使用 PathManager 解析路径
	outputDir := pathManager.ResolvePath(requestedOutputDir)
	logger.Infof("输出目录解析: 原始=%s, 处理后=%s, 解析后=%s", request.OutputDir, requestedOutputDir, outputDir)
	if logCallback != nil {
		logCallback("info", fmt.Sprintf("输出目录: %s", outputDir))
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		errMsg := fmt.Sprintf("创建输出目录失败: %v", err)
		logger.Errorf(errMsg)
		if logCallback != nil {
			logCallback("error", errMsg)
		}
		return fmt.Errorf(errMsg)
	}
	logger.Infof("输出目录已验证/创建: %s", outputDir)
	if logCallback != nil {
		logCallback("info", fmt.Sprintf("输出目录已验证/创建: %s", outputDir))
	}

	// 处理每张图片
	taggedCount := 0
	failedCount := 0
	lastError := ""

	logger.Infof("开始处理 %d 张图片", len(imageFiles))
	if logCallback != nil {
		logCallback("info", fmt.Sprintf("开始处理 %d 张图片", len(imageFiles)))
	}

	for i, imagePath := range imageFiles {
		logger.Infof("处理图片 %d/%d: %s", i+1, len(imageFiles), filepath.Base(imagePath))
		if logCallback != nil && (i == 0 || (i+1)%10 == 0 || i == len(imageFiles)-1) {
			logCallback("info", fmt.Sprintf("处理图片 %d/%d: %s", i+1, len(imageFiles), filepath.Base(imagePath)))
		}

		// 调用进度回调
		if callback != nil {
			callback(i+1, len(imageFiles))
		}

		// 读取图片并转换为 Base64
		imageData, err := readImageAsBase64(imagePath)
		if err != nil {
			errMsg := fmt.Sprintf("读取图片失败 %s: %v", filepath.Base(imagePath), err)
			logger.Errorf(errMsg)
			if logCallback != nil {
				logCallback("error", errMsg)
			}
			failedCount++
			lastError = fmt.Sprintf("读取图片失败: %v", err)
			continue
		}

		// 调用 WD14 Tagger API
		result, err := t.interrogateImage(imageData, request)
		if err != nil {
			errMsg := fmt.Sprintf("生成标签失败 %s: %v", filepath.Base(imagePath), err)
			logger.Errorf(errMsg)
			if logCallback != nil {
				logCallback("error", errMsg)
			}
			failedCount++
			lastError = fmt.Sprintf("生成标签失败: %v", err)
			continue
		}

		// 保存标签文件
		if err := t.saveTags(imagePath, result, outputDir, request.SaveType); err != nil {
			errMsg := fmt.Sprintf("保存标签文件失败 %s: %v", filepath.Base(imagePath), err)
			logger.Errorf(errMsg)
			if logCallback != nil {
				logCallback("error", errMsg)
			}
			failedCount++
			lastError = fmt.Sprintf("保存标签文件失败: %v", err)
			continue
		}

		taggedCount++
		logger.Infof("成功处理图片 %d/%d", taggedCount, len(imageFiles))
		if logCallback != nil && (taggedCount%10 == 0 || taggedCount == len(imageFiles)) {
			logCallback("info", fmt.Sprintf("成功处理 %d/%d 张图片", taggedCount, len(imageFiles)))
		}
	}

	logger.Infof("标签生成完成: 成功处理 %d/%d 张图片", taggedCount, len(imageFiles))
	if logCallback != nil {
		logCallback("info", fmt.Sprintf("标签生成完成: 成功处理 %d/%d 张图片", taggedCount, len(imageFiles)))
	}

	// 如果所有图片都处理失败，返回错误
	if taggedCount == 0 && failedCount > 0 {
		errMsg := fmt.Sprintf("所有图片处理失败（失败 %d 张），最后错误: %s", failedCount, lastError)
		logger.Errorf(errMsg)
		if logCallback != nil {
			logCallback("error", errMsg)
		}
		return fmt.Errorf(errMsg)
	}

	// 如果部分失败，记录警告但继续
	if failedCount > 0 {
		warnMsg := fmt.Sprintf("部分图片处理失败（成功 %d 张，失败 %d 张）", taggedCount, failedCount)
		logger.Warnf(warnMsg)
		if logCallback != nil {
			logCallback("warning", warnMsg)
		}
	}

	return nil
}

// TagConfidence 标签和置信度
type TagConfidence struct {
	Name       string
	Confidence float64
}

// InterrogateResult 包含原始标签数据和解析后的标签字符串
type InterrogateResult struct {
	TagsString string                 `json:"tags_string"` // 逗号分隔的标签字符串
	TagsData   map[string]interface{} `json:"tags_data"`   // 原始标签数据（tag -> confidence）
	RatingData map[string]interface{} `json:"rating_data"` // rating 数据
}

// interrogateImage 调用标签生成 API（WD14 Tagger 或 DeepBooru），返回完整结果
func (t *WD14Tagger) interrogateImage(imageData string, request *models.TagRequest) (*InterrogateResult, error) {
	var url string
	var jsonData []byte
	var err error

	// 根据分析器类型选择不同的 API
	if t.analyzer == "deepbooru" {
		// DeepBooru 使用标准 WebUI API
		url = fmt.Sprintf("%s/sdapi/v1/interrogate", strings.TrimRight(t.baseURL, "/"))

		// DeepBooru 的请求体格式
		reqBody := map[string]interface{}{
			"image": imageData,
			"model": "deepbooru", // 指定使用 deepbooru 模型
		}
		jsonData, err = json.Marshal(reqBody)
		logger.Infof("发送 DeepBooru API 请求")
	} else {
		// WD14-Tagger 使用扩展 API
		url = fmt.Sprintf("%s/tagger/v1/interrogate", strings.TrimRight(t.baseURL, "/"))

		// WD14-Tagger 的请求体格式
		reqBody := InterrogateRequest{
			Image:      imageData,
			Model:      t.model, // 使用配置的模型（默认 wd14-convnext-v2，支持 CPU）
			Threshold:  t.threshold,
			BatchCount: 1,
		}
		jsonData, err = json.Marshal(reqBody)
		logger.Infof("发送 WD14 Tagger 请求: model=%s, threshold=%.2f", t.model, t.threshold)
	}

	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 记录实际响应内容用于调试
	analyzerName := "WD14 Tagger"
	if t.analyzer == "deepbooru" {
		analyzerName = "DeepBooru"
	}
	logger.Infof("%s API 响应: status=%d, body长度=%d", analyzerName, resp.StatusCode, len(respBody))
	if len(respBody) < 500 {
		logger.Infof("完整响应: %s", string(respBody))
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		// 检查是否是 CUDA/ONNX Runtime 错误
		if strings.Contains(string(respBody), "ONNX Runtime") || strings.Contains(string(respBody), "CUDA") || strings.Contains(string(respBody), "cuDNN") {
			return nil, fmt.Errorf("GPU/CUDA 配置错误: %s\n\n"+
				"解决方案：\n"+
				"1. 确认已安装 CUDA 12.x 和匹配的 cuDNN 版本\n"+
				"2. 或在 WebUI 中配置使用 CPU 模式（修改 WebUI 启动参数或扩展配置）\n"+
				"3. 检查 WebUI 日志了解详细错误信息", string(respBody))
		}
		return nil, fmt.Errorf("API 调用失败: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	// 首先尝试作为 map 解析
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &rawResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// logger.Infof("原始响应内容: %+v", rawResponse)

	// DeepBooru 可能返回不同的字段名
	var captionInterface interface{}
	var exists bool

	if t.analyzer == "deepbooru" {
		// DeepBooru 可能返回 "caption" 或 "result"
		captionInterface, exists = rawResponse["caption"]
		if !exists {
			captionInterface, exists = rawResponse["result"]
		}
	} else {
		// WD14-Tagger 返回 "caption" 字段
		captionInterface, exists = rawResponse["caption"]
	}

	if !exists {
		return nil, fmt.Errorf("响应中没有 caption 字段（分析器: %s）: %+v", t.analyzer, rawResponse)
	}

	logger.Infof("提取到 caption: 类型=%T, 值（前200字符）=%v", captionInterface, func() string {
		str := fmt.Sprintf("%v", captionInterface)
		if len(str) > 200 {
			return str[:200] + "..."
		}
		return str
	}())

	// 处理 caption 字段（可能是字符串、对象或数组）
	result := &InterrogateResult{
		TagsData:   make(map[string]interface{}),
		RatingData: make(map[string]interface{}),
	}

	switch v := captionInterface.(type) {
	case string:
		result.TagsString = v
		logger.Infof("caption 是字符串: %s", result.TagsString)
	case map[string]interface{}:
		// 如果是对象，可能是两种格式：
		// 1. { "tags": [...], "rating": [...] } - 结构化格式
		// 2. { "tag1": 0.9, "tag2": 0.8 } - 标签名->置信度映射
		logger.Infof("caption 是对象，包含 %d 个键", len(v))
		logger.Infof("前10个键: %v", func() []string {
			keys := make([]string, 0, 10)
			count := 0
			for k := range v {
				if count < 10 {
					keys = append(keys, k)
					count++
				}
			}
			return keys
		}())

		// 将 map 按 rating 和 tags 分类
		tagNames := make([]string, 0, len(v))
		result.TagsData = make(map[string]interface{})
		result.RatingData = make(map[string]interface{})

		for key, value := range v {
			// 检查是否是 rating 相关字段
			if key == "general" || key == "sensitive" || key == "questionable" || key == "explicit" {
				result.RatingData[key] = value
			} else if key != "rating" {
				// 这是一个标签
				result.TagsData[key] = value

				// 检查置信度（如果是数字），满足阈值才加入标签字符串
				if conf, ok := value.(float64); ok {
					if conf >= t.threshold {
						tagNames = append(tagNames, key)
					}
				} else {
					// 如果不是数字，也加入
					tagNames = append(tagNames, key)
				}
			}
		}

		result.TagsString = strings.Join(tagNames, ", ")
		logger.Infof("提取到 %d 个标签（阈值过滤后）", len(tagNames))
		logger.Infof("Rating 数据: %+v", result.RatingData)
	case []interface{}:
		// 如果是数组
		tagStrings := make([]string, 0, len(v))
		for _, tag := range v {
			if tagStr, ok := tag.(string); ok && tagStr != "" {
				tagStrings = append(tagStrings, tagStr)
			}
		}
		result.TagsString = strings.Join(tagStrings, ", ")
		logger.Infof("从数组提取到标签: %s", result.TagsString)
	default:
		return nil, fmt.Errorf("未知的 caption 类型: %T, 值: %v", v, v)
	}

	if result.TagsString == "" {
		return nil, fmt.Errorf("无法提取标签，caption: %+v", captionInterface)
	}

	logger.Infof("最终标签: %s", result.TagsString)

	// 返回完整结果
	return result, nil
}

// saveTags 保存标签到文件
func (t *WD14Tagger) saveTags(imagePath string, result *InterrogateResult, outputDir, saveType string) error {
	logger.Infof("开始保存标签: 图片=%s, 输出目录=%s, 保存类型=%s", imagePath, outputDir, saveType)

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}
	logger.Infof("输出目录已创建/验证: %s", outputDir)

	// 生成输出文件名（与原图同名，扩展名不同）
	imageName := filepath.Base(imagePath)
	imageNameWithoutExt := strings.TrimSuffix(imageName, filepath.Ext(imageName))

	var outputPath string
	switch saveType {
	case "json":
		outputPath = filepath.Join(outputDir, imageNameWithoutExt+".json")
		logger.Infof("保存为 JSON 格式: %s", outputPath)

		// 按置信度排序 tags
		tagList := make([]TagConfidence, 0, len(result.TagsData))
		for tagName, confValue := range result.TagsData {
			if conf, ok := confValue.(float64); ok {
				tagList = append(tagList, TagConfidence{Name: tagName, Confidence: conf})
			}
		}
		// 按置信度降序排序
		for i := 0; i < len(tagList); i++ {
			for j := i + 1; j < len(tagList); j++ {
				if tagList[i].Confidence < tagList[j].Confidence {
					tagList[i], tagList[j] = tagList[j], tagList[i]
				}
			}
		}

		// 构建类似 WebUI 的输出结构，包含格式化数据
		formattedOutput := t.formatWebUIOutput(result, tagList)

		// 构建类似 WebUI 的输出结构
		tagData := map[string]interface{}{
			"image_path":  imagePath,
			"tags_string": result.TagsString,
			"tags":        result.TagsData,
			"ratings":     result.RatingData,
			"formatted":   formattedOutput, // WebUI 风格的格式化输出
			"tags_sorted": tagList,         // 排序后的标签列表
			"created_at":  time.Now().Format(time.RFC3339),
		}
		jsonData, err := json.MarshalIndent(tagData, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化 JSON 失败: %v", err)
		}
		if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
		logger.Infof("✓ 成功保存 JSON 文件: %s (大小: %d 字节)", outputPath, len(jsonData))
	case "txt":
		fallthrough
	default:
		outputPath = filepath.Join(outputDir, imageNameWithoutExt+".txt")
		logger.Infof("保存为 TXT 格式: %s", outputPath)
		// 保存为 TXT 格式
		if err := os.WriteFile(outputPath, []byte(result.TagsString), 0644); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
		logger.Infof("✓ 成功保存 TXT 文件: %s (大小: %d 字节)", outputPath, len(result.TagsString))
	}

	return nil
}

// formatWebUIOutput 格式化为 WebUI 风格的输出
func (t *WD14Tagger) formatWebUIOutput(result *InterrogateResult, tagList []TagConfidence) string {
	var builder strings.Builder

	// Tags
	builder.WriteString("Tags\n")
	builder.WriteString(result.TagsString)
	builder.WriteString("\n\n")

	// Rating confidents
	builder.WriteString("Rating confidents\n")
	ratingKeys := []string{"explicit", "general", "sensitive", "questionable"}
	for _, key := range ratingKeys {
		if val, ok := result.RatingData[key]; ok {
			conf := 0.0
			if v, ok := val.(float64); ok {
				conf = v
			}
			builder.WriteString(fmt.Sprintf("%s\n%s\n%.0f%%\n\n", key, key, conf*100))
		}
	}

	// Tag confidents
	builder.WriteString("Tag confidents\n")
	for _, tag := range tagList {
		builder.WriteString(fmt.Sprintf("%s\n%s\n%.0f%%\n\n", tag.Name, tag.Name, tag.Confidence*100))
	}

	return builder.String()
}

// getImageFiles 获取目录中的所有图片文件
func getImageFiles(dir string, limit int) ([]string, error) {
	var imageFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" {
			imageFiles = append(imageFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 应用限制
	if limit > 0 && limit < len(imageFiles) {
		imageFiles = imageFiles[:limit]
	}

	return imageFiles, nil
}

// readImageAsBase64 读取图片并转换为 Base64
func readImageAsBase64(imagePath string) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}
