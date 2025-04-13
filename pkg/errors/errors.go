package errors

import (
	"errors"
	"fmt"
)

// 继承标准库错误
var (
	New    = errors.New
	Unwrap = errors.Unwrap
	Is     = errors.Is
	As     = errors.As
)

// AppError 表示应用程序错误
type AppError struct {
	Code    string // 错误代码
	Message string // 错误信息
	Err     error  // 原始错误
}

// 预定义错误代码
const (
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeInvalidInput  = "INVALID_INPUT"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeUnavailable   = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout       = "TIMEOUT"
	ErrCodeRateLimited   = "RATE_LIMITED"
	ErrCodeDatabaseError = "DATABASE_ERROR"
)

// Error 返回错误字符串
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 返回底层错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建一个新的应用程序错误
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NotFound 创建一个"未找到"错误
func NotFound(message string, err error) *AppError {
	return NewAppError(ErrCodeNotFound, message, err)
}

// InvalidInput 创建一个"无效输入"错误
func InvalidInput(message string, err error) *AppError {
	return NewAppError(ErrCodeInvalidInput, message, err)
}

// Unauthorized 创建一个"未授权"错误
func Unauthorized(message string, err error) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, err)
}

// Forbidden 创建一个"禁止访问"错误
func Forbidden(message string, err error) *AppError {
	return NewAppError(ErrCodeForbidden, message, err)
}

// Internal 创建一个"内部错误"
func Internal(message string, err error) *AppError {
	return NewAppError(ErrCodeInternal, message, err)
}

// Conflict 创建一个"冲突"错误
func Conflict(message string, err error) *AppError {
	return NewAppError(ErrCodeConflict, message, err)
}

// Unavailable 创建一个"服务不可用"错误
func Unavailable(message string, err error) *AppError {
	return NewAppError(ErrCodeUnavailable, message, err)
}

// Timeout 创建一个"超时"错误
func Timeout(message string, err error) *AppError {
	return NewAppError(ErrCodeTimeout, message, err)
}

// RateLimited 创建一个"速率限制"错误
func RateLimited(message string, err error) *AppError {
	return NewAppError(ErrCodeRateLimited, message, err)
}

// DatabaseError 创建一个"数据库错误"
func DatabaseError(message string, err error) *AppError {
	return NewAppError(ErrCodeDatabaseError, message, err)
}

// GetErrorCode 从错误中提取错误代码
func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeInternal
}

// GetErrorMessage 从错误中提取错误信息
func GetErrorMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}

// Wrap 包装错误并添加上下文信息
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
} 