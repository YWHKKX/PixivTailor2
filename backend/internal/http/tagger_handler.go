package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pixiv-tailor/backend/internal/service"
	"pixiv-tailor/backend/pkg/models"
)

// TaggerHandler 标签处理器
type TaggerHandler struct {
	taskService service.TaskService
}

// NewTaggerHandler 创建标签处理器
func NewTaggerHandler(taskService service.TaskService) *TaggerHandler {
	return &TaggerHandler{
		taskService: taskService,
	}
}

// CreateTagTask 创建标签任务
func (h *TaggerHandler) CreateTagTask(w http.ResponseWriter, r *http.Request) {
	var request models.TagRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	// 验证请求参数
	if request.InputDir == nil {
		http.Error(w, "输入目录不能为空", http.StatusBadRequest)
		return
	}

	// 检查 InputDir 是否为空字符串或空数组
	inputDirs := request.GetInputDirs()
	if len(inputDirs) == 0 || (len(inputDirs) == 1 && inputDirs[0] == "") {
		http.Error(w, "输入目录不能为空", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if request.Analyzer == "" {
		request.Analyzer = "wd14tagger"
	}
	if request.TagOrder == "" {
		request.TagOrder = "score"
	}
	if request.SaveType == "" {
		request.SaveType = "txt"
	}
	if request.Limit == 0 {
		request.Limit = 100
	}

	// 创建任务配置
	config := map[string]interface{}{
		"type":        "tag",
		"input_dir":   request.InputDir,
		"output_dir":  request.OutputDir,
		"analyzer":    request.Analyzer,
		"skip_tags":   request.SkipTags,
		"extend_tags": request.ExtendTags,
		"tag_order":   request.TagOrder,
		"save_type":   request.SaveType,
		"limit":       request.Limit,
	}

	// 创建任务
	configJSON, _ := json.Marshal(config)
	task, err := h.taskService.CreateTask("tag", string(configJSON))
	if err != nil {
		http.Error(w, fmt.Sprintf("创建任务失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回任务信息
	response := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    200,
			"message": "任务创建成功",
		},
		"data": task,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTaggedImages 获取已标签的图像
func (h *TaggerHandler) GetTaggedImages(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	_ = taskID // 暂时未使用

	// 获取tags目录路径
	tagsDir := filepath.Join("backend", "data", "tags")
	if _, err := os.Stat(tagsDir); os.IsNotExist(err) {
		response := map[string]interface{}{
			"status": map[string]interface{}{
				"code":    200,
				"message": "获取成功",
			},
			"data": []models.TaggedImage{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 读取所有JSON文件
	files, err := ioutil.ReadDir(tagsDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取目录失败: %v", err), http.StatusInternalServerError)
		return
	}

	taggedImages := []models.TaggedImage{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// 读取JSON文件
		filePath := filepath.Join(tagsDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue
		}

		// 解析JSON
		var tagData map[string]interface{}
		if err := json.Unmarshal(data, &tagData); err != nil {
			continue
		}

		// 提取图片路径
		imagePath, ok := tagData["image_path"].(string)
		if !ok {
			continue
		}

		// 转换为相对路径
		relativePath := strings.Replace(imagePath, filepath.Join("backend", "data", "images")+string(filepath.Separator), "images/", 1)

		// 提取标签
		tags := []models.Tag{}
		if tagsSorted, ok := tagData["tags_sorted"].([]interface{}); ok {
			for i, tagObj := range tagsSorted {
				if tagMap, ok := tagObj.(map[string]interface{}); ok {
					name, _ := tagMap["Name"].(string)
					confidence, _ := tagMap["Confidence"].(float64)
					tags = append(tags, models.Tag{
						Name:      name,
						Score:     confidence,
						Category:  "general",
						IsGeneral: true,
					})
					if i >= 15 { // 限制标签数量
						break
					}
				}
			}
		}

		// 创建TaggedImage对象
		taggedImage := models.TaggedImage{
			BaseModel: models.BaseModel{
				ID:        int64(len(taggedImages) + 1),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			ImagePath: relativePath,
			Tags:      tags,
			Analyzer:  "wd14tagger",
			Metadata: map[string]string{
				"filename": file.Name(),
			},
		}

		taggedImages = append(taggedImages, taggedImage)
	}

	response := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    200,
			"message": "获取成功",
		},
		"data": taggedImages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAvailableAnalyzers 获取可用的分析器
func (h *TaggerHandler) GetAvailableAnalyzers(w http.ResponseWriter, r *http.Request) {
	analyzers := []string{
		"wd14tagger",
		"deepdanbooru",
	}

	response := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    200,
			"message": "获取成功",
		},
		"data": analyzers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AnalyzeImage 分析单张图像
func (h *TaggerHandler) AnalyzeImage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ImagePath string `json:"image_path"`
		Analyzer  string `json:"analyzer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	// 验证参数
	if request.ImagePath == "" {
		http.Error(w, "图像路径不能为空", http.StatusBadRequest)
		return
	}
	if request.Analyzer == "" {
		request.Analyzer = "wd14tagger"
	}

	// 这里应该调用实际的图像分析服务
	// 暂时返回模拟数据
	taggedImage := models.TaggedImage{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ImagePath: request.ImagePath,
		Tags: []models.Tag{
			{Name: "1girl", Score: 0.95, Category: "character", IsGeneral: true},
			{Name: "anime", Score: 0.90, Category: "style", IsGeneral: true},
			{Name: "cute", Score: 0.85, Category: "quality", IsGeneral: false},
		},
		Analyzer: request.Analyzer,
		Metadata: map[string]string{
			"width":  "512",
			"height": "512",
		},
	}

	response := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    200,
			"message": "分析成功",
		},
		"data": taggedImage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTagTaskStatus 获取标签任务状态
func (h *TaggerHandler) GetTagTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 获取任务状态
	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取任务失败: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    200,
			"message": "获取成功",
		},
		"data": task,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopTagTask 停止标签任务
func (h *TaggerHandler) StopTagTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "任务ID不能为空", http.StatusBadRequest)
		return
	}

	// 停止任务
	err := h.taskService.CancelTask(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("停止任务失败: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    200,
			"message": "任务已停止",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
