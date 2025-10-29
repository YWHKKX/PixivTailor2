package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pixiv-tailor/backend/internal/ai"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/internal/service"
	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"
)

// AIHandler AI相关请求处理器
type AIHandler struct {
	taskService             service.TaskService
	generationConfigService service.GenerationConfigService
}

// NewAIHandler 创建AI处理器
func NewAIHandler(aiService *ai.AIManager, taskService service.TaskService, generationConfigService service.GenerationConfigService) *AIHandler {
	handler := &AIHandler{
		taskService:             taskService,
		generationConfigService: generationConfigService,
	}

	// 注册 generate 任务的执行器
	taskService.RegisterExecutor("generate", handler)

	return handler
}

// ExecuteGenerateTask 实现 TaskExecutor 接口，执行AI生成任务
func (h *AIHandler) ExecuteGenerateTask(ctx context.Context, taskID string, config map[string]interface{}) {
	// 获取循环数量
	loopCount := 1
	if loopCountVal, ok := config["loop_count"].(float64); ok {
		loopCount = int(loopCountVal)
	}
	if loopCount <= 0 {
		loopCount = 1
	}

	logger.Infof("任务 %s 开始执行，循环数量: %d", taskID, loopCount)

	// 存储所有生成的图片URL
	var allImageUrls []string
	var totalImagesGenerated int

	// 循环发包
	for currentLoop := 1; currentLoop <= loopCount; currentLoop++ {
		logger.Infof("开始第 %d/%d 次发包", currentLoop, loopCount)

		// 检查上下文是否已被取消
		select {
		case <-ctx.Done():
			h.taskService.UpdateTaskStatus(taskID, "cancelled")
			h.taskService.UpdateTaskError(taskID, "任务已被用户取消")
			logger.Infof("任务 %s 已取消", taskID)
			return
		default:
		}

		// 计算当前循环的进度
		loopProgress := int(float64(currentLoop-1) / float64(loopCount) * 90) // 0-90%用于循环
		h.taskService.UpdateTaskProgress(taskID, 5+loopProgress)

		// 转发到 WebUI API
		h.taskService.UpdateTaskProgress(taskID, 5+loopProgress+5)
		response, err := h.forwardToWebUI(config)
		if err != nil {
			logger.Infof("第 %d 次发包失败: %v", currentLoop, err)
			h.taskService.UpdateTaskProgress(taskID, 0)
			h.taskService.UpdateTaskError(taskID, fmt.Sprintf("第 %d 次发包失败: %v", currentLoop, err))
			h.taskService.UpdateTaskStatus(taskID, "failed")
			return
		}

		// response已经是map[string]interface{}类型，直接使用
		webuiResponse := response

		// 检查是否有图片
		if images, ok := webuiResponse["images"].([]interface{}); ok && len(images) > 0 {
			logger.Infof("第 %d 次发包返回了 %d 张图片", currentLoop, len(images))

			// 更新图片统计
			h.taskService.UpdateTaskImagesFound(taskID, totalImagesGenerated+len(images))

			// 更新进度到当前循环的50%
			currentLoopProgress := int(float64(currentLoop-1)/float64(loopCount)*90) + int(float64(currentLoop)/float64(loopCount)*50)
			h.taskService.UpdateTaskProgress(taskID, 5+currentLoopProgress)

			// 下载并保存图片
			if err := h.downloadAndSaveImagesWithOffset(taskID, images, totalImagesGenerated); err != nil {
				logger.Infof("第 %d 次发包图片下载失败: %v", currentLoop, err)
				h.taskService.UpdateTaskProgress(taskID, 0)
				h.taskService.UpdateTaskError(taskID, fmt.Sprintf("图片下载失败: %v", err))
				h.taskService.UpdateTaskStatus(taskID, "failed")
				return
			}

			// 更新成功下载的图片统计
			h.taskService.UpdateTaskImagesDownloaded(taskID, totalImagesGenerated+len(images))

			// 构建图片URL列表
			for i := range images {
				imageUrl := fmt.Sprintf("http://localhost:50052/api/tasks/%s/images/%d", taskID, totalImagesGenerated+i+1)
				allImageUrls = append(allImageUrls, imageUrl)
			}
			totalImagesGenerated += len(images)

			// 更新进度到当前循环的80%
			currentLoopProgress = int(float64(currentLoop-1)/float64(loopCount)*90) + int(float64(currentLoop)/float64(loopCount)*80)
			h.taskService.UpdateTaskProgress(taskID, 5+currentLoopProgress)

			logger.Infof("第 %d 次发包完成，累计生成 %d 张图片", currentLoop, totalImagesGenerated)
		} else {
			logger.Infof("第 %d 次发包未返回图片数据", currentLoop)
			h.taskService.UpdateTaskProgress(taskID, 0)
			h.taskService.UpdateTaskError(taskID, "WebUI响应中缺少图片数据")
			h.taskService.UpdateTaskStatus(taskID, "failed")
			return
		}

		// 如果不是最后一次循环，等待一下再继续
		if currentLoop < loopCount {
			logger.Infof("等待 2 秒后开始下一次发包...")
			time.Sleep(2 * time.Second)
		}
	}

	// 所有循环完成，更新任务状态为完成
	taskResult := map[string]interface{}{
		"images": allImageUrls,
		"count":  totalImagesGenerated,
		"loops":  loopCount,
	}
	h.taskService.UpdateTaskResult(taskID, taskResult)
	h.taskService.UpdateTaskProgress(taskID, 100)
	h.taskService.UpdateTaskStatus(taskID, "completed")
	logger.Infof("任务 %s 完成，共执行 %d 次发包，生成 %d 张图片", taskID, loopCount, totalImagesGenerated)
}

// AIGenerationRequest AI生成请求
type AIGenerationRequest struct {
	Prompt         string              `json:"prompt"`
	NegativePrompt string              `json:"negative_prompt"`
	Steps          int                 `json:"steps"`
	CFGScale       float64             `json:"cfg_scale"`
	Width          int                 `json:"width"`
	Height         int                 `json:"height"`
	Seed           int64               `json:"seed"`
	Model          string              `json:"model"`
	Sampler        string              `json:"sampler"`
	BatchSize      int                 `json:"batch_size"`
	EnableHR       bool                `json:"enable_hr"`
	Loras          []models.LoraConfig `json:"loras,omitempty"`
}

// AIGenerationResponse AI生成响应
type AIGenerationResponse struct {
	TaskID      string     `json:"task_id"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Message     string     `json:"message"`
	Result      []string   `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// GenerationWithConfigRequest 使用配置生成请求
type GenerationWithConfigRequest struct {
	ConfigID string                 `json:"config_id"`
	Override map[string]interface{} `json:"override,omitempty"`
}

// HandleGenerate 已删除 - 使用 HandleGenerateWithConfig 替代

// HandleGenerateWithConfig 使用配置生成图像
func (h *AIHandler) HandleGenerateWithConfig(w http.ResponseWriter, r *http.Request) {
	var req GenerationWithConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 检查 WebUI 是否运行
	webuiStatus := h.checkWebUIStatus()
	if !webuiStatus {
		h.sendErrorResponse(w, http.StatusServiceUnavailable, "WebUI is not running", "请先启动 WebUI")
		return
	}

	// 获取配置文件
	config, err := h.loadConfigFromFile(req.ConfigID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Config not found", err.Error())
		return
	}

	logger.Infof("HandleGenerateWithConfig - 配置模型: '%s'", config.Model)

	// 合并覆盖参数
	generationParams := h.mergeConfigWithOverride(config, req.Override)

	// 创建任务记录
	taskConfig := map[string]interface{}{
		"type":            "generate",
		"config_id":       req.ConfigID,
		"prompt":          generationParams["prompt"],
		"negative_prompt": generationParams["negative_prompt"],
		"steps":           generationParams["steps"],
		"cfg_scale":       generationParams["cfg_scale"],
		"width":           generationParams["width"],
		"height":          generationParams["height"],
		"batch_size":      generationParams["batch_size"],
		"batch_count":     generationParams["batch_count"],
		"loop_count":      generationParams["loop_count"],
		"sampler":         generationParams["sampler"],
		"seed":            generationParams["seed"],
		"model":           generationParams["model"],
		"enable_hr":       generationParams["enable_hr"],
	}

	configJSON, _ := json.Marshal(taskConfig)
	task, err := h.taskService.CreateTask("generate", string(configJSON))
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create task", err.Error())
		return
	}

	// 不再需要后台处理，task_service 会自动调用 ExecuteGenerateTask
	h.sendSuccessResponse(w, map[string]interface{}{
		"task_id": task.ID,
		"message": "AI生成任务已创建",
		"status":  task.Status,
	})
}

// checkWebUIStatus 检查 WebUI 是否运行
func (h *AIHandler) checkWebUIStatus() bool {
	// 通过HTTP请求检查WebUI API是否可用
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 尝试访问WebUI的API端点
	resp, err := client.Get("http://127.0.0.1:7860/sdapi/v1/options")
	if err != nil {
		return false
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	// 检查响应状态码
	return resp.StatusCode == http.StatusOK
}

// loadConfigFromFile 从文件加载配置
func (h *AIHandler) loadConfigFromFile(configID string) (*models.GenerationConfigResponse, error) {
	// 直接使用配置ID作为文件名（配置ID已经是去掉.json扩展名的文件名）
	configPath := filepath.Join("data", "configs", configID+".json")

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON配置
	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 转换为标准格式
	modelName := h.getStringFromMap(config, "model")
	// logger.Infof("配置解析 - 模型名称: '%s'", modelName)

	// 如果模型名称为空，尝试从原始配置中直接获取
	if modelName == "" {
		if modelVal, exists := config["model"]; exists {
			logger.Infof("原始模型值: %v, 类型: %T", modelVal, modelVal)
			if str, ok := modelVal.(string); ok {
				modelName = str
				logger.Infof("重新解析模型名称: '%s'", modelName)
			}
		}
	}

	response := &models.GenerationConfigResponse{
		ID:                h.getStringFromMap(config, "id"),
		Name:              h.getStringFromMap(config, "name"),
		Prompt:            h.getStringFromMap(config, "prompt"),
		NegativePrompt:    h.getStringFromMap(config, "negative_prompt"),
		Steps:             h.getIntFromMap(config, "steps", 20),
		CFGScale:          h.getFloat64FromMap(config, "cfg_scale", 7.0),
		Width:             h.getIntFromMap(config, "width", 512),
		Height:            h.getIntFromMap(config, "height", 512),
		Sampler:           h.getStringFromMap(config, "sampler"),
		BatchSize:         h.getIntFromMap(config, "batch_size", 1),
		Model:             modelName,
		Seed:              h.getInt64FromMap(config, "seed", -1),
		EnableHR:          h.getBoolFromMap(config, "enable_hr", false),
		RestoreFaces:      h.getBoolFromMap(config, "restore_faces", false),
		ClipSkip:          h.getIntFromMap(config, "clip_skip", 2),
		DenoisingStrength: h.getFloat64FromMap(config, "denoising_strength", 0.5),
		VAE:               h.getStringFromMap(config, "vae"),
		Tiling:            h.getBoolFromMap(config, "tiling", false),
		UpscaleBy:         h.getFloat64FromMap(config, "upscale_by", 1.0),
		Upscaler:          h.getStringFromMap(config, "upscaler"),
		HiresSteps:        h.getIntFromMap(config, "hires_steps", 0),
		HiresUpscaler:     h.getStringFromMap(config, "hires_upscaler"),
		HiresUpscale:      h.getFloat64FromMap(config, "hires_upscale", 1.0),
	}

	// 解析LoRA配置
	if lorasData, ok := config["loras"].([]interface{}); ok {
		loras := make([]models.LoraConfig, 0, len(lorasData))
		for _, loraData := range lorasData {
			if loraMap, ok := loraData.(map[string]interface{}); ok {
				lora := models.LoraConfig{
					Name:        h.getStringFromMap(loraMap, "name"),
					FullName:    h.getStringFromMap(loraMap, "full_name"),
					Weight:      h.getFloat64FromMap(loraMap, "weight", 1.0),
					Path:        h.getStringFromMap(loraMap, "path"),
					Description: h.getStringFromMap(loraMap, "description"),
				}
				loras = append(loras, lora)
			}
		}
		response.Loras = loras
	}

	return response, nil
}

// forwardToWebUI 转发请求到 WebUI API
func (h *AIHandler) forwardToWebUI(params map[string]interface{}) (map[string]interface{}, error) {
	// 获取并转换参数，确保不是 nil
	getInt := func(key string, defaultValue int) int {
		if val, ok := params[key]; ok && val != nil {
			switch v := val.(type) {
			case int:
				return v
			case float64:
				return int(v)
			case int64:
				return int(v)
			}
		}
		return defaultValue
	}

	getFloat64 := func(key string, defaultValue float64) float64 {
		if val, ok := params[key]; ok && val != nil {
			switch v := val.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			}
		}
		return defaultValue
	}

	getString := func(key string, defaultValue string) string {
		if val, ok := params[key]; ok && val != nil {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return defaultValue
	}

	getBool := func(key string, defaultValue bool) bool {
		if val, ok := params[key]; ok && val != nil {
			if b, ok := val.(bool); ok {
				return b
			}
		}
		return defaultValue
	}

	// 构造 WebUI API 请求 - 参考PixivTailor项目的实现
	webuiRequest := map[string]interface{}{
		"prompt":                 getString("prompt", ""),
		"negative_prompt":        getString("negative_prompt", ""),
		"steps":                  getInt("steps", 20),
		"cfg_scale":              getFloat64("cfg_scale", 7.0),
		"width":                  getInt("width", 512),
		"height":                 getInt("height", 512),
		"sampler_name":           getString("sampler", "DPM++ 2M Karras"),
		"batch_size":             getInt("batch_size", 1),
		"n_iter":                 getInt("batch_count", 1),
		"seed":                   getInt("seed", -1),
		"return_grid":            true,
		"restore_faces":          getBool("restore_faces", false),
		"face_restoration":       "CodeFormer",
		"face_restoration_model": "null",
		"send_images":            true, // 强制返回图片数据
		"save_images":            true, // 保存到磁盘
		"do_not_save_samples":    false,
		"do_not_save_grid":       getBool("do_not_save_grid", false),
		"enable_hr":              getBool("enable_hr", false),
		"enable-checkbox":        false,
		"switch_at":              0.8,
		"denoising_strength":     getFloat64("denoising_strength", 0.7),
		"firstphase_width":       0,
		"firstphase_height":      0,
		"hr_scale":               getFloat64("upscale_by", 2.0),
		"hr_second_pass_steps":   0,
		"hr_resize_x":            0,
		"hr_resize_y":            0,
		"hires_steps":            getInt("hires_steps", 0),
		"hr_checkpoint_name":     "",
		"hr_sampler_name":        "",
		"hr_prompt":              getString("prompt", ""),          // 使用相同的prompt
		"hr_negative_prompt":     getString("negative_prompt", ""), // 使用相同的negative_prompt
		"hr-checkbox":            false,
		"hr_upscaler":            getString("hires_upscaler", "Latent"),
		"tiling":                 getBool("tiling", false),
		"clip_skip":              getInt("clip_skip", 2),
		"eta":                    getFloat64("eta", 0.0),
		"ensd":                   getInt("ensd", 0),
		"override_settings": func() map[string]interface{} {
			model := getString("model", "")
			settings := map[string]interface{}{}
			if model != "" {
				settings["sd_model_checkpoint"] = model
			}
			// 只有当VAE不为空时才添加VAE设置
			vae := getString("vae", "")
			if vae != "" {
				settings["sd_vae"] = vae
			}
			return settings
		}(),
		"override_settings_restore_afterwards": true,
	}

	// 添加LoRA配置 - 参考PixivTailor项目的实现
	if loras, ok := params["loras"].([]models.LoraConfig); ok && len(loras) > 0 {
		// 在prompt中添加LoRA标签，格式：<lora:name:weight>
		prompt := webuiRequest["prompt"].(string)
		for _, lora := range loras {
			if lora.Weight > 0 {
				prompt += fmt.Sprintf(", <lora:%s:%.2f>", lora.Name, lora.Weight)
			}
		}
		webuiRequest["prompt"] = prompt
	}

	// 发送到 WebUI API - 使用无超时的客户端
	requestBody, err := json.Marshal(webuiRequest)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:7860/sdapi/v1/txt2img", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 创建没有超时限制的HTTP客户端
	client := &http.Client{
		Timeout: 0, // 无超时限制
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WebUI API 请求失败: %v", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		errorMap := make(map[string]interface{})
		json.Unmarshal(body, &errorMap)

		// 检查是否是 CUDA out of memory 错误
		errorMsg := ""
		if errVal, ok := errorMap["error"]; ok {
			errorMsg = fmt.Sprintf("%v", errVal)
		}
		if errDetail, ok := errorMap["detail"]; ok {
			errorMsg += fmt.Sprintf(" - %v", errDetail)
		}

		// 如果是 CUDA OOM 错误，提供更详细的建议
		if strings.Contains(errorMsg, "out of memory") || strings.Contains(errorMsg, "CUDA") {
			return nil, fmt.Errorf("显存不足 (GPU Out of Memory): %s\n\n建议:\n1. 降低图片尺寸 (width/height)\n2. 将 batch_size 设为 1\n3. 关闭其他占用显存的程序\n4. 重启 WebUI 释放显存\n5. 使用更小的模型", errorMsg)
		}

		return nil, fmt.Errorf("WebUI API 返回错误: %d, %s", resp.StatusCode, errorMsg)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析 WebUI 响应失败: %v", err)
	}

	return result, nil
}

// processGenerationTask 已删除 - 使用 processGenerationTaskWithConfig 替代

// setDefaultValues 已删除 - 不再需要

// createTaskConfig 已删除 - 不再需要

// mergeConfigWithOverride 合并配置和覆盖参数
func (h *AIHandler) mergeConfigWithOverride(config *models.GenerationConfigResponse, override map[string]interface{}) map[string]interface{} {
	generationParams := map[string]interface{}{
		"prompt":             config.Prompt,
		"negative_prompt":    config.NegativePrompt,
		"steps":              config.Steps,
		"cfg_scale":          config.CFGScale,
		"width":              config.Width,
		"height":             config.Height,
		"seed":               config.Seed,
		"model":              config.Model,
		"sampler":            config.Sampler,
		"batch_size":         config.BatchSize,
		"batch_count":        1, // 默认批次数量
		"loop_count":         1, // 默认循环次数
		"enable_hr":          config.EnableHR,
		"restore_faces":      config.RestoreFaces,
		"clip_skip":          config.ClipSkip,
		"denoising_strength": config.DenoisingStrength,
		"vae":                config.VAE,
		"tiling":             config.Tiling,
		"upscale_by":         config.UpscaleBy,
		"upscaler":           config.Upscaler,
		"hires_steps":        config.HiresSteps,
		"hires_upscaler":     config.HiresUpscaler,
		"hires_upscale":      config.HiresUpscale,
		"eta":                config.Eta,
		"ensd":               config.ENSD,
		"save_images":        config.SaveImages,
		"save_grid":          config.SaveGrid,
		"send_images":        config.SendImages,
		"do_not_save_grid":   config.DoNotSaveGrid,
		"loras":              config.Loras,
	}

	// 应用覆盖参数
	for k, v := range override {
		// 对于空字符串的覆盖参数，跳过覆盖（保持配置中的值）
		if str, ok := v.(string); ok && str == "" {
			continue
		}

		generationParams[k] = v
	}

	return generationParams
}

// convertToGenerationRequest 已删除 - 不再需要

// saveGeneratedImages 已删除 - 使用 downloadAndSaveImages 替代

// 辅助函数
func (h *AIHandler) getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (h *AIHandler) getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return defaultValue
}

func (h *AIHandler) getInt64FromMap(m map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return defaultValue
}

func (h *AIHandler) getFloat64FromMap(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultValue
}

func (h *AIHandler) getBoolFromMap(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// downloadAndSaveImages 下载并保存WebUI生成的图片
func (h *AIHandler) downloadAndSaveImages(taskID string, images []interface{}) error {
	return h.downloadAndSaveImagesWithOffset(taskID, images, 0)
}

// downloadAndSaveImagesWithOffset 下载并保存WebUI生成的图片（带偏移量）
func (h *AIHandler) downloadAndSaveImagesWithOffset(taskID string, images []interface{}, offset int) error {
	// 获取路径管理器
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return fmt.Errorf("路径管理器未初始化")
	}

	// 获取任务信息以生成正确的任务目录名
	task, err := h.taskService.GetTask(taskID)
	var taskDir string

	if err == nil && task != nil {
		// 使用新格式：[时间]_任务类型_哈希值
		taskDir = pathManager.GetTaskImagesDir(taskID, task.Type, task.CreatedAt)
	} else {
		// 回退到旧格式：task_{taskID}
		taskDir = filepath.Join(pathManager.GetImagesDir(), fmt.Sprintf("task_%s", taskID))
	}

	// 确保目录存在
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("创建任务图片目录失败: %v", err)
	}

	// 下载每张图片
	for i, imageData := range images {
		imageStr, ok := imageData.(string)
		if !ok {
			logger.Infof("图片 %d 数据格式错误", i+1)
			continue
		}

		// 解码base64图片数据
		imageBytes, err := h.decodeBase64Image(imageStr)
		if err != nil {
			logger.Infof("解码图片 %d 失败: %v", i+1, err)
			continue
		}

		// 检测图片格式并确定文件扩展名
		fileExt := h.detectImageFormat(imageBytes)
		filename := fmt.Sprintf("generated_%s_%d.%s", taskID, offset+i+1, fileExt)
		filepath := filepath.Join(taskDir, filename)

		if err := os.WriteFile(filepath, imageBytes, 0644); err != nil {
			logger.Infof("保存图片 %d 失败: %v", i+1, err)
			continue
		}

		logger.Infof("图片已保存: %s", filepath)
	}

	return nil
}

// decodeBase64Image 解码base64图片数据
func (h *AIHandler) decodeBase64Image(imageStr string) ([]byte, error) {
	// WebUI返回的图片数据可能包含前缀，需要移除
	// 格式通常是: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
	var base64Data string
	if strings.HasPrefix(imageStr, "data:image/") {
		// 移除data URL前缀
		parts := strings.Split(imageStr, ",")
		if len(parts) == 2 {
			base64Data = parts[1]
		} else {
			base64Data = imageStr
		}
	} else {
		base64Data = imageStr
	}

	// 解码base64数据
	imageBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("base64解码失败: %v", err)
	}

	return imageBytes, nil
}

// detectImageFormat 检测图片格式
func (h *AIHandler) detectImageFormat(imageBytes []byte) string {
	// 检查文件头来确定图片格式
	if len(imageBytes) < 4 {
		return "png" // 默认格式
	}

	// PNG: 89 50 4E 47
	if imageBytes[0] == 0x89 && imageBytes[1] == 0x50 && imageBytes[2] == 0x4E && imageBytes[3] == 0x47 {
		return "png"
	}

	// JPEG: FF D8 FF
	if imageBytes[0] == 0xFF && imageBytes[1] == 0xD8 && imageBytes[2] == 0xFF {
		return "jpg"
	}

	// WebP: 52 49 46 46 (RIFF)
	if len(imageBytes) >= 12 && imageBytes[0] == 0x52 && imageBytes[1] == 0x49 && imageBytes[2] == 0x46 && imageBytes[3] == 0x46 {
		if imageBytes[8] == 0x57 && imageBytes[9] == 0x45 && imageBytes[10] == 0x42 && imageBytes[11] == 0x50 {
			return "webp"
		}
	}

	// 默认返回PNG
	return "png"
}

// stopWebUIGeneration 停止WebUI的当前生成任务
func (h *AIHandler) stopWebUIGeneration() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 调用WebUI的interrupt API
	resp, err := client.Post("http://127.0.0.1:7860/sdapi/v1/interrupt", "application/json", nil)
	if err != nil {
		return fmt.Errorf("调用WebUI停止API失败: %v", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WebUI停止API返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// sendSuccessResponse 发送成功响应
func (h *AIHandler) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": map[string]interface{}{
			"code":    0,
			"message": "success",
		},
		"data": data,
	})
}

// sendErrorResponse 发送错误响应
func (h *AIHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": map[string]interface{}{
			"code":    statusCode,
			"message": message,
			"details": details,
		},
	})
}
