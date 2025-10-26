package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	httphandler "pixiv-tailor/backend/internal/http"
	"pixiv-tailor/backend/internal/repository"
	"pixiv-tailor/backend/pkg/models"
)

// MockTaskService 模拟任务服务
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(taskType, config string) (*repository.Task, error) {
	args := m.Called(taskType, config)
	return args.Get(0).(*repository.Task), args.Error(1)
}

func (m *MockTaskService) GetTask(id string) (*repository.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*repository.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTaskStatus(id, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskProgress(id string, progress int) error {
	args := m.Called(id, progress)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskError(id, errorMsg string) error {
	args := m.Called(id, errorMsg)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskCompletedAt(id string, completedAt time.Time) error {
	args := m.Called(id, completedAt)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskImagesFound(id string, count int) error {
	args := m.Called(id, count)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskImagesDownloaded(id string, count int) error {
	args := m.Called(id, count)
	return args.Error(0)
}

func (m *MockTaskService) UpdateTaskResult(id string, result map[string]interface{}) error {
	args := m.Called(id, result)
	return args.Error(0)
}

func (m *MockTaskService) ListTasks(page, pageSize int32, status, taskType string) ([]*repository.Task, int, error) {
	args := m.Called(page, pageSize, status, taskType)
	return args.Get(0).([]*repository.Task), args.Int(1), args.Error(2)
}

func (m *MockTaskService) StartTask(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskService) StopTask(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskService) CancelTask(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskService) DeleteTask(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskService) CleanupTasks(cleanupType string) (int, error) {
	args := m.Called(cleanupType)
	return args.Int(0), args.Error(1)
}

func (m *MockTaskService) SetLogCallback(callback func(taskID, level, message string)) {
	m.Called(callback)
}

func (m *MockTaskService) SetStatusCallback(callback func(taskID, status string, progress int)) {
	m.Called(callback)
}

// MockGenerationConfigService 模拟生成配置服务
type MockGenerationConfigService struct {
	mock.Mock
}

func (m *MockGenerationConfigService) CreateConfig(req *models.GenerationConfigRequest) (*models.GenerationConfigResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) GetConfig(id string) (*models.GenerationConfigResponse, error) {
	args := m.Called(id)
	return args.Get(0).(*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) GetConfigByName(name string) (*models.GenerationConfigResponse, error) {
	args := m.Called(name)
	return args.Get(0).(*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) UpdateConfig(id string, req *models.GenerationConfigRequest) (*models.GenerationConfigResponse, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) DeleteConfig(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGenerationConfigService) ListConfigs(req *models.ConfigSearchRequest) (*models.ConfigListResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.ConfigListResponse), args.Error(1)
}

func (m *MockGenerationConfigService) GetCategories() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGenerationConfigService) GetDefaultConfig() (*models.GenerationConfigResponse, error) {
	args := m.Called()
	return args.Get(0).(*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) SetDefaultConfig(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGenerationConfigService) UseConfig(id string) (*models.GenerationConfigResponse, error) {
	args := m.Called(id)
	return args.Get(0).(*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) ImportConfig(req *models.ConfigImportRequest) (*models.GenerationConfigResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) ExportConfigs(req *models.ConfigExportRequest) ([]*models.GenerationConfigResponse, error) {
	args := m.Called(req)
	return args.Get(0).([]*models.GenerationConfigResponse), args.Error(1)
}

func (m *MockGenerationConfigService) UpdateConfigFile(id string, config map[string]interface{}) error {
	args := m.Called(id, config)
	return args.Error(0)
}

func (m *MockGenerationConfigService) CreateConfigFile(config map[string]interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockGenerationConfigService) DeleteConfigFile(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAIHandler_HandleGenerateWithConfig(t *testing.T) {
	// 创建模拟服务
	mockTaskService := &MockTaskService{}
	mockConfigService := &MockGenerationConfigService{}

	// 创建AI处理器
	handler := httphandler.NewAIHandler(nil, mockTaskService, mockConfigService)

	// 测试数据
	testConfig := &models.GenerationConfigResponse{
		ID:          "test-config",
		Name:        "测试配置",
		Category:    "动漫",
		Description: "测试用配置",
		Model:       "test_model.safetensors",
		Loras: []models.LoraConfig{
			{
				Path:   "test_lora.safetensors",
				Weight: 0.8,
			},
		},
		Steps:    20,
		CFGScale: 7,
		Width:    512,
		Height:   512,
		Sampler:  "DPM++ 2M Karras",
		EnableHR: false,
	}

	testTask := &repository.Task{
		ID:     "test-task-id",
		Type:   "generate",
		Status: "pending",
		Config: `{"config_id":"test-config","prompt":"test prompt"}`,
	}

	// 设置模拟期望
	mockConfigService.On("GetConfig", "test-config").Return(testConfig, nil)
	mockTaskService.On("CreateTask", "generate", mock.AnythingOfType("string")).Return(testTask, nil)

	// 创建测试请求
	requestBody := models.GenerateRequest{
		Model:          "test_model.safetensors",
		Prompt:         "test prompt",
		NegativePrompt: "bad quality",
		Steps:          20,
		CFGScale:       7,
		Width:          512,
		Height:         512,
		BatchCount:     1,
		EnableHR:       false,
		Options:        map[string]interface{}{},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/generate-with-config", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 执行请求
	handler.HandleGenerateWithConfig(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "AI生成任务已创建", response["message"])

	// 验证模拟调用
	mockConfigService.AssertExpectations(t)
	mockTaskService.AssertExpectations(t)
}

func TestAIHandler_HandleGenerateWithConfig_InvalidConfig(t *testing.T) {
	// 创建模拟服务
	mockTaskService := &MockTaskService{}
	mockConfigService := &MockGenerationConfigService{}

	// 创建AI处理器
	handler := httphandler.NewAIHandler(nil, mockTaskService, mockConfigService)

	// 设置模拟期望 - 配置不存在
	mockConfigService.On("GetConfig", "invalid-config").Return((*models.GenerationConfigResponse)(nil), assert.AnError)

	// 创建测试请求
	requestBody := models.GenerateRequest{
		Model:  "test_model.safetensors",
		Prompt: "test prompt",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/generate-with-config", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 执行请求
	handler.HandleGenerateWithConfig(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])

	// 验证模拟调用
	mockConfigService.AssertExpectations(t)
}

// 简化的测试函数
func TestAIHandler_BasicFunctionality(t *testing.T) {
	// 创建模拟服务
	mockTaskService := &MockTaskService{}
	mockConfigService := &MockGenerationConfigService{}

	// 创建AI处理器
	handler := httphandler.NewAIHandler(nil, mockTaskService, mockConfigService)

	// 验证处理器已创建
	assert.NotNil(t, handler)
}
