package http

import (
	"encoding/json"
	"net/http"
	"time"

	"pixiv-tailor/backend/internal/models"
	"pixiv-tailor/backend/internal/service"
)

// CharacterHandler 角色特征处理器
type CharacterHandler struct {
	characterService *service.CharacterService
}

// NewCharacterHandler 创建角色特征处理器
func NewCharacterHandler() *CharacterHandler {
	return &CharacterHandler{
		characterService: service.NewCharacterService(),
	}
}

// handleExtractCharacterTags 提取角色特征标签
func (s *HTTPServer) handleExtractCharacterTags(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SourceDir    string   `json:"source_dir"`
		MinFrequency int      `json:"min_frequency"`
		MinWeight    float64  `json:"min_weight"`
		MaxTags      int      `json:"max_tags"`
		ExcludeTags  []string `json:"exclude_tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "解析请求失败", err.Error())
		return
	}

	// 设置默认值
	if request.MinFrequency == 0 {
		request.MinFrequency = 2 // 至少出现2次
	}
	if request.MinWeight == 0 {
		request.MinWeight = 0.3 // 最低置信度30%
	}
	if request.MaxTags == 0 {
		request.MaxTags = 20
	}

	charRequest := models.CharacterProfileRequest{
		MinFrequency: request.MinFrequency,
		MinWeight:    request.MinWeight,
		MaxTags:      request.MaxTags,
		ExcludeTags:  request.ExcludeTags,
	}

	tags, err := s.characterService.ExtractCharacterTags(request.SourceDir, charRequest)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "提取标签失败", err.Error())
		return
	}

	s.sendSuccessResponse(w, tags)
}

// handleCreateCharacterProfile 创建角色特征配置
func (s *HTTPServer) handleCreateCharacterProfile(w http.ResponseWriter, r *http.Request) {
	var request models.CharacterProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "解析请求失败", err.Error())
		return
	}

	// 验证必填字段
	if request.Name == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "角色名称不能为空", "")
		return
	}
	if request.SourceDir == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "源目录不能为空", "")
		return
	}

	// 设置默认值
	if request.MinFrequency == 0 {
		request.MinFrequency = 2
	}
	if request.MinWeight == 0 {
		request.MinWeight = 0.3
	}
	if request.MaxTags == 0 {
		request.MaxTags = 20
	}

	profile, err := s.characterService.CreateCharacterProfile(request)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "创建角色配置失败", err.Error())
		return
	}

	s.sendSuccessResponse(w, profile)
}

// handleListCharacterProfiles 列出所有角色配置
func (s *HTTPServer) handleListCharacterProfiles(w http.ResponseWriter, r *http.Request) {
	if s.characterService == nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "characterService未初始化", "")
		return
	}

	profiles, err := s.characterService.ListCharacterProfiles()
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "获取角色配置失败", err.Error())
		return
	}

	s.sendSuccessResponse(w, profiles)
}

// handleGetCharacterProfile 获取角色配置
func (s *HTTPServer) handleGetCharacterProfile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "角色名称不能为空", "")
		return
	}

	profile, err := s.characterService.LoadCharacterProfile(name)
	if err != nil {
		s.sendErrorResponse(w, http.StatusNotFound, "角色配置不存在", err.Error())
		return
	}

	s.sendSuccessResponse(w, profile)
}

// handleUpdateCharacterProfile 更新角色特征配置
func (s *HTTPServer) handleUpdateCharacterProfile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Tags        []string           `json:"tags"`
		TagWeights  map[string]float64 `json:"tag_weights"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "解析请求失败", err.Error())
		return
	}

	// 验证必填字段
	if request.Name == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "角色名称不能为空", "")
		return
	}

	// 先加载现有配置
	profile, err := s.characterService.LoadCharacterProfile(request.Name)
	if err != nil {
		s.sendErrorResponse(w, http.StatusNotFound, "角色配置不存在", err.Error())
		return
	}

	// 更新字段
	profile.Description = request.Description
	if len(request.Tags) > 0 {
		profile.Tags = request.Tags
	}
	if len(request.TagWeights) > 0 {
		profile.TagWeights = request.TagWeights
	}
	profile.UpdatedAt = time.Now()

	// 保存更新后的配置
	if err := s.characterService.SaveCharacterProfile(profile); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "更新角色配置失败", err.Error())
		return
	}

	s.sendSuccessResponse(w, profile)
}

// handleDeleteCharacterProfile 删除角色特征配置
func (s *HTTPServer) handleDeleteCharacterProfile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "解析请求失败", err.Error())
		return
	}

	// 验证必填字段
	if request.Name == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "角色名称不能为空", "")
		return
	}

	// 删除文件
	if err := s.characterService.DeleteCharacterProfile(request.Name); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "删除角色配置失败", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]string{"message": "删除成功"})
}
