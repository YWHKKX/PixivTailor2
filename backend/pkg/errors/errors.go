package errors

import (
	"fmt"
)

// ErrorCode 错误代码类型
type ErrorCode int

const (
	// 通用错误
	ErrCodeUnknown ErrorCode = iota
	ErrCodeInvalidParam
	ErrCodeNotFound
	ErrCodeUnauthorized
	ErrCodeForbidden
	ErrCodeTimeout
	ErrCodeInternal

	// 配置相关错误
	ErrCodeConfigLoad
	ErrCodeConfigParse
	ErrCodeConfigValidate
	ErrCodeConfigSave

	// 任务相关错误
	ErrCodeTaskCreate
	ErrCodeTaskNotFound
	ErrCodeTaskCancel
	ErrCodeTaskExecute

	// 爬虫相关错误
	ErrCodeCrawlFailed
	ErrCodeCrawlTimeout
	ErrCodeCrawlAuth

	// AI相关错误
	ErrCodeAIGenerate
	ErrCodeAITrain
	ErrCodeAITag
	ErrCodeAIClassify

	// 存储相关错误
	ErrCodeStorageConnect
	ErrCodeStorageQuery
	ErrCodeStorageSave
	ErrCodeStorageDelete
)

// PixivTailorError 自定义错误类型
type PixivTailorError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *PixivTailorError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewError 创建新错误
func NewError(code ErrorCode, message string) *PixivTailorError {
	return &PixivTailorError{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithDetails 创建带详情的错误
func NewErrorWithDetails(code ErrorCode, message, details string) *PixivTailorError {
	return &PixivTailorError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Is 检查错误是否匹配指定的错误代码
func Is(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}

	if pixivErr, ok := err.(*PixivTailorError); ok {
		return pixivErr.Code == code
	}

	return false
}

// Wrap 包装错误
func Wrap(err error, code ErrorCode, message string) *PixivTailorError {
	details := ""
	if err != nil {
		details = err.Error()
	}
	return NewErrorWithDetails(code, message, details)
}

// 预定义错误
var (
	// 通用错误
	ErrUnknown      = NewError(ErrCodeUnknown, "未知错误")
	ErrInvalidParam = NewError(ErrCodeInvalidParam, "参数无效")
	ErrNotFound     = NewError(ErrCodeNotFound, "资源不存在")
	ErrUnauthorized = NewError(ErrCodeUnauthorized, "未授权")
	ErrForbidden    = NewError(ErrCodeForbidden, "禁止访问")
	ErrTimeout      = NewError(ErrCodeTimeout, "操作超时")
	ErrInternal     = NewError(ErrCodeInternal, "内部错误")

	// 配置相关错误
	ErrConfigLoad     = NewError(ErrCodeConfigLoad, "配置加载失败")
	ErrConfigParse    = NewError(ErrCodeConfigParse, "配置解析失败")
	ErrConfigValidate = NewError(ErrCodeConfigValidate, "配置验证失败")
	ErrConfigSave     = NewError(ErrCodeConfigSave, "配置保存失败")

	// 任务相关错误
	ErrTaskCreate   = NewError(ErrCodeTaskCreate, "任务创建失败")
	ErrTaskNotFound = NewError(ErrCodeTaskNotFound, "任务不存在")
	ErrTaskCancel   = NewError(ErrCodeTaskCancel, "任务取消失败")
	ErrTaskExecute  = NewError(ErrCodeTaskExecute, "任务执行失败")

	// 爬虫相关错误
	ErrCrawlFailed  = NewError(ErrCodeCrawlFailed, "爬取失败")
	ErrCrawlTimeout = NewError(ErrCodeCrawlTimeout, "爬取超时")
	ErrCrawlAuth    = NewError(ErrCodeCrawlAuth, "爬取认证失败")

	// AI相关错误
	ErrAIGenerate = NewError(ErrCodeAIGenerate, "AI生成失败")
	ErrAITrain    = NewError(ErrCodeAITrain, "AI训练失败")
	ErrAITag      = NewError(ErrCodeAITag, "AI标签失败")
	ErrAIClassify = NewError(ErrCodeAIClassify, "AI分类失败")

	// 存储相关错误
	ErrStorageConnect = NewError(ErrCodeStorageConnect, "存储连接失败")
	ErrStorageQuery   = NewError(ErrCodeStorageQuery, "存储查询失败")
	ErrStorageSave    = NewError(ErrCodeStorageSave, "存储保存失败")
	ErrStorageDelete  = NewError(ErrCodeStorageDelete, "存储删除失败")
)
