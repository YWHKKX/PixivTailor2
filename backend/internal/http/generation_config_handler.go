package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pixiv-tailor/backend/pkg/models"
	"pixiv-tailor/backend/pkg/paths"

	"github.com/gorilla/mux"
)

// 辅助函数
func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}

func getInt64FromMap(m map[string]interface{}, key string, defaultValue int64) int64 {
	if v, ok := m[key]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int64(f)
		}
	}
	return defaultValue
}

func getFloat64FromMap(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
		if i, ok := v.(int); ok {
			return float64(i)
		}
	}
	return defaultValue
}

func getBoolFromMap(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// 查询参数辅助函数
func getStringFromQuery(query map[string][]string, key string, defaultValue string) string {
	if values, ok := query[key]; ok && len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func getIntFromQuery(query map[string][]string, key string, defaultValue int) int {
	if values, ok := query[key]; ok && len(values) > 0 {
		if i, err := strconv.Atoi(values[0]); err == nil {
			return i
		}
	}
	return defaultValue
}

// saveConfigToFile 保存配置到JSON文件
func (s *HTTPServer) saveConfigToFile(config *models.GenerationConfig) error {
	// 使用 PathManager 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		return fmt.Errorf("path manager not initialized")
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 确保配置目录存在
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		return fmt.Errorf("failed to create configs directory: %v", err)
	}

	// 构建文件名（使用配置名称，替换特殊字符）
	fileName := fmt.Sprintf("%s.json", sanitizeFileName(config.Name))
	filePath := filepath.Join(configsDir, fileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("config file already exists: %s", fileName)
	}

	// 创建配置数据
	configData := map[string]interface{}{
		"name":            config.Name,
		"description":     config.Description,
		"category":        config.Category,
		"prompt":          config.Prompt,
		"negative_prompt": config.NegativePrompt,
		"steps":           config.Steps,
		"cfg_scale":       config.CFGScale,
		"width":           config.Width,
		"height":          config.Height,
		"seed":            config.Seed,
		"model":           config.Model,
		"sampler":         config.Sampler,
		"batch_size":      config.BatchSize,
		"enable_hr":       config.EnableHR,
		"loras":           config.Loras,
		"other_params":    config.OtherParams,
		"is_default":      config.IsDefault,
		"created_at":      config.CreatedAt.Format(time.RFC3339),
		"updated_at":      config.UpdatedAt.Format(time.RFC3339),
	}

	// 写入JSON文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(configData); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// sanitizeFileName 清理文件名，移除特殊字符
func sanitizeFileName(name string) string {
	// 替换不允许的字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

// convertResponseToConfig 将响应转换为配置模型
func convertResponseToConfig(response *models.GenerationConfigResponse) *models.GenerationConfig {
	return &models.GenerationConfig{
		ID:             response.ID,
		Name:           response.Name,
		Description:    response.Description,
		Category:       response.Category,
		Prompt:         response.Prompt,
		NegativePrompt: response.NegativePrompt,
		Steps:          response.Steps,
		CFGScale:       response.CFGScale,
		Width:          response.Width,
		Height:         response.Height,
		Seed:           response.Seed,
		Model:          response.Model,
		Sampler:        response.Sampler,
		BatchSize:      response.BatchSize,
		EnableHR:       response.EnableHR,
		Loras:          response.Loras,
		OtherParams:    response.OtherParams,
		IsDefault:      response.IsDefault,
		CreatedAt:      response.CreatedAt,
		UpdatedAt:      response.UpdatedAt,
	}
}

// 配置文件相关处理器

// handleCreateConfig 创建配置
func (s *HTTPServer) handleCreateConfig(w http.ResponseWriter, r *http.Request) {
	var req models.GenerationConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 检查配置名称是否已存在
	existingConfig, err := s.GenerationConfigService.GetConfigByName(req.Name)
	if err == nil && existingConfig != nil {
		s.sendErrorResponse(w, http.StatusConflict, "Config name already exists", "配置名称已存在，请使用不同的名称")
		return
	}

	// 创建配置到数据库
	config, err := s.GenerationConfigService.CreateConfig(&req)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create config", err.Error())
		return
	}

	// 保存配置到JSON文件
	if err := s.saveConfigToFile(convertResponseToConfig(config)); err != nil {
		// 如果保存文件失败，记录错误但不影响数据库操作
		log.Printf("Failed to save config to file: %v", err)
	}

	s.sendSuccessResponse(w, config)
}

// handleGetConfig 获取配置
func (s *HTTPServer) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// 使用 PathManager 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Path manager not initialized", "")
		return
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 读取配置目录中的所有JSON文件
	files, err := filepath.Glob(filepath.Join(configsDir, "*.json"))
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to read config directory", err.Error())
		return
	}

	// 查找指定ID的配置
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err != nil {
			continue
		}

		// 生成配置ID（基于文件名，去掉.json扩展名）
		fileName := filepath.Base(file)
		configId := strings.TrimSuffix(fileName, ".json")

		// 检查是否匹配（优先使用配置文件中的id字段，如果没有则使用文件名）
		var match bool
		if configID, ok := config["id"].(string); ok && configID != "" {
			match = configID == id
		} else {
			match = configId == id
		}

		if match {
			// 添加文件路径信息和ID
			config["file_path"] = file
			config["file_name"] = fileName
			config["id"] = configId
			s.sendSuccessResponse(w, config)
			return
		}
	}

	s.sendErrorResponse(w, http.StatusNotFound, "Config not found", "配置未找到")
}

// handleGetConfigByName 根据名称获取配置
func (s *HTTPServer) handleGetConfigByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	config, err := s.GenerationConfigService.GetConfigByName(name)
	if err != nil {
		s.sendErrorResponse(w, http.StatusNotFound, "Config not found", err.Error())
		return
	}

	s.sendSuccessResponse(w, config)
}

// handleUpdateConfig 更新配置
func (s *HTTPServer) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.GenerationConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	config, err := s.GenerationConfigService.UpdateConfig(id, &req)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to update config", err.Error())
		return
	}

	s.sendSuccessResponse(w, config)
}

// handleDeleteConfig 删除配置
func (s *HTTPServer) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.GenerationConfigService.DeleteConfig(id); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete config", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Config deleted successfully"})
}

// handleListConfigs 列出配置 - 从文件系统读取
func (s *HTTPServer) handleListConfigs(w http.ResponseWriter, r *http.Request) {
	// 使用 PathManager 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Path manager not initialized", "")
		return
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 读取配置目录中的所有JSON文件
	files, err := filepath.Glob(filepath.Join(configsDir, "*.json"))
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to read config directory", err.Error())
		return
	}

	var allConfigs []map[string]interface{}
	var categories []string
	categorySet := make(map[string]bool)

	for _, file := range files {
		// 读取文件内容
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		// 解析JSON
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err != nil {
			continue
		}

		// 添加文件路径信息
		config["file_path"] = file
		config["file_name"] = filepath.Base(file)

		// 生成配置ID（基于文件名，去掉.json扩展名）
		fileName := filepath.Base(file)
		configId := strings.TrimSuffix(fileName, ".json")
		config["id"] = configId

		// 收集分类
		if category, ok := config["category"].(string); ok && category != "" {
			if !categorySet[category] {
				categories = append(categories, category)
				categorySet[category] = true
			}
		}

		allConfigs = append(allConfigs, config)
	}

	// 应用过滤和搜索
	var filteredConfigs []map[string]interface{}
	query := r.URL.Query()
	searchQuery := getStringFromQuery(query, "name", "")
	categoryFilter := getStringFromQuery(query, "category", "")

	for _, config := range allConfigs {
		// 分类过滤
		if categoryFilter != "" {
			if configCategory, ok := config["category"].(string); !ok || configCategory != categoryFilter {
				continue
			}
		}

		// 搜索过滤
		if searchQuery != "" {
			name, _ := config["name"].(string)
			description, _ := config["description"].(string)
			if !strings.Contains(strings.ToLower(name), strings.ToLower(searchQuery)) &&
				!strings.Contains(strings.ToLower(description), strings.ToLower(searchQuery)) {
				continue
			}
		}

		filteredConfigs = append(filteredConfigs, config)
	}

	// 分页处理
	page := getIntFromQuery(query, "page", 1)
	pageSize := getIntFromQuery(query, "pageSize", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	total := len(filteredConfigs)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var pagedConfigs []map[string]interface{}
	if start < total {
		pagedConfigs = filteredConfigs[start:end]
	}

	response := map[string]interface{}{
		"configs":    pagedConfigs,
		"total":      int64(total),
		"page":       page,
		"page_size":  pageSize,
		"categories": categories,
	}

	s.sendSuccessResponse(w, response)
}

// handleGetCategories 获取分类 - 从文件系统读取
func (s *HTTPServer) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	// 使用 PathManager 获取配置目录
	pathManager := paths.GetPathManager()
	if pathManager == nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Path manager not initialized", "")
		return
	}

	configsDir := filepath.Join(pathManager.GetDataDir(), "configs")

	// 读取配置目录中的所有JSON文件
	files, err := filepath.Glob(filepath.Join(configsDir, "*.json"))
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to read config directory", err.Error())
		return
	}

	var categories []string
	categorySet := make(map[string]bool)

	for _, file := range files {
		// 读取文件内容
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		// 解析JSON
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err != nil {
			continue
		}

		// 收集分类
		if category, ok := config["category"].(string); ok && category != "" {
			if !categorySet[category] {
				categories = append(categories, category)
				categorySet[category] = true
			}
		}
	}

	s.sendSuccessResponse(w, map[string]interface{}{
		"categories": categories,
	})
}

// handleGetDefaultConfig 获取默认配置
func (s *HTTPServer) handleGetDefaultConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.GenerationConfigService.GetDefaultConfig()
	if err != nil {
		s.sendErrorResponse(w, http.StatusNotFound, "Default config not found", err.Error())
		return
	}

	s.sendSuccessResponse(w, config)
}

// handleSetDefaultConfig 设置默认配置
func (s *HTTPServer) handleSetDefaultConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := s.GenerationConfigService.SetDefaultConfig(req.ID); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to set default config", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Default config set successfully"})
}

// handleUseConfig 使用配置
func (s *HTTPServer) handleUseConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	config, err := s.GenerationConfigService.UseConfig(id)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to use config", err.Error())
		return
	}

	s.sendSuccessResponse(w, config)
}

// handleImportConfig 导入配置
func (s *HTTPServer) handleImportConfig(w http.ResponseWriter, r *http.Request) {
	var req models.ConfigImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 检查配置名称是否已存在
	existingConfig, err := s.GenerationConfigService.GetConfigByName(req.ConfigName)
	if err == nil && existingConfig != nil {
		s.sendErrorResponse(w, http.StatusConflict, "Config name already exists", "配置名称已存在，请使用不同的名称")
		return
	}

	// 导入配置到数据库
	config, err := s.GenerationConfigService.ImportConfig(&req)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to import config", err.Error())
		return
	}

	// 保存配置到JSON文件
	if err := s.saveConfigToFile(convertResponseToConfig(config)); err != nil {
		// 如果保存文件失败，记录错误但不影响数据库操作
		log.Printf("Failed to save imported config to file: %v", err)
	}

	s.sendSuccessResponse(w, config)
}

// handleExportConfigs 导出配置
func (s *HTTPServer) handleExportConfigs(w http.ResponseWriter, r *http.Request) {
	var req models.ConfigExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	configs, err := s.GenerationConfigService.ExportConfigs(&req)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to export configs", err.Error())
		return
	}

	s.sendSuccessResponse(w, configs)
}

// handleImportFromFile 从文件导入配置
func (s *HTTPServer) handleImportFromFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 配置管理已改为文件系统模式，请直接编辑JSON文件
	s.sendErrorResponse(w, http.StatusNotImplemented, "配置管理已改为文件系统模式", "请直接编辑 backend/data/configs/ 目录下的JSON文件")
}

// handleExportToFile 导出配置到文件
func (s *HTTPServer) handleExportToFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfigIDs []string `json:"config_ids"`
		FilePath  string   `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// 配置管理已改为文件系统模式，请直接复制JSON文件
	s.sendErrorResponse(w, http.StatusNotImplemented, "配置管理已改为文件系统模式", "请直接复制 backend/data/configs/ 目录下的JSON文件")
}

// handleGenerateWithConfig 使用配置生成图片
func (s *HTTPServer) handleGenerateWithConfig(w http.ResponseWriter, r *http.Request) {
	aiHandler := NewAIHandler(nil, s.TaskService, s.GenerationConfigService)
	aiHandler.HandleGenerateWithConfig(w, r)
}

// handleUpdateConfigFile 更新配置文件
func (s *HTTPServer) handleUpdateConfigFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := s.GenerationConfigService.UpdateConfigFile(id, config); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to update config file", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Config file updated successfully"})
}

// handleCreateConfigFile 创建配置文件
func (s *HTTPServer) handleCreateConfigFile(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := s.GenerationConfigService.CreateConfigFile(config); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create config file", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Config file created successfully"})
}

// handleDeleteConfigFile 删除配置文件
func (s *HTTPServer) handleDeleteConfigFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.GenerationConfigService.DeleteConfigFile(id); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete config file", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "Config file deleted successfully"})
}
