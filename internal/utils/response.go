package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	ErrorCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInternalServer  ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrorCodeConflict        ErrorCode = "CONFLICT"
	ErrorCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
)

type AppError struct {
	Message string
	Code    ErrorCode
	Err     error
}
type APIResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
	Pagination any    `json:"pagination,omitempty"`
}

func (ae *AppError) Error() string { // Implement the error interface cho phép AppError có thể được sử dụng như một error
	return ""
}

// sử dụng khi : validation lỗi, không tìm thấy dữ liệu, lỗi máy chủ nội bộ, không có quyền truy cập, xung đột dữ liệu
func NewError(message string, code ErrorCode) error {
	return &AppError{
		Message: message,
		Code:    code,
	}
}

// sử dụng khi : lỗi từ database, lỗi từ service bên ngoài
func WrapError(err error, message string, code ErrorCode) error {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// ResponseError phân tích lỗi và trả về mã trạng thái HTTP tương ứng (trong handler)
func ResponseError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		status := httpStatusFromCode(appErr.Code)
		response := gin.H{
			"error": CapitalizeFirst(appErr.Message),
			"code":  appErr.Code,
		}
		if appErr.Err != nil {
			response["details"] = appErr.Err.Error() // thêm thông tin chi tiết nếu có
		}
		c.JSON(status, response)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
		"code":  ErrorCodeInternalServer,
	})

}
func ResponSuccess(c *gin.Context, status int, message string, data ...any) {
	response := APIResponse{
		Status:  "success",
		Message: CapitalizeFirst(message),
	}
	if len(data) > 0 && data[0] != nil {

		if m, ok := data[0].(map[string]any); ok {
			if p, exists := m["pagination"]; exists {
				response.Pagination = p
			}
			if d, exists := m["data"]; exists {
				response.Data = d
			} else {
				response.Data = m
			}

		} else {
			response.Data = data[0]
		}

	}
	c.JSON(status, response)
}
func ResponseStatusCode(c *gin.Context, status int) {
	c.Status(status)
}
func ResponseValidator(c *gin.Context, data any) {
	c.JSON(http.StatusBadRequest, data)
}
func httpStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrorCodeBadRequest:
		return 400
	case ErrorCodeNotFound:
		return 404
	case ErrorCodeValidation:
		return 422
	case ErrorCodeTooManyRequests:
		return 429
	case ErrorCodeInternalServer:
		return 500
	case ErrorCodeUnauthorized:
		return 401
	case ErrorCodeForbidden:
		return 403
	case ErrorCodeConflict:
		return 409
	default:
		return 500 // Mặc định trả về lỗi máy chủ nội bộ nếu không xác định được mã lỗi
	}
}
