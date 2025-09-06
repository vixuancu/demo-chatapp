package validation

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitValidator() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("không thể lấy validator từ binding")
	}
	RegisterCustomValidation(v)
	return nil
}
func HandleValidationError(err error) gin.H {
	if validation, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range validation {
			switch e.Tag() {
			case "gt":
				errors[e.Field()] = fmt.Sprintf("%s phải lớn hơn %s", e.Field(), e.Param())
			case "lt":
				errors[e.Field()] = fmt.Sprintf("%s phải nhỏ hơn %s", e.Field(), e.Param())
			case "gte":
				errors[e.Field()] = fmt.Sprintf("%s phải lớn hơn hoặc bằng %s", e.Field(), e.Param())
			case "lte":
				errors[e.Field()] = fmt.Sprintf("%s phải nhỏ hơn hoặc bằng %s", e.Field(), e.Param())
			case "uuid":
				errors[e.Field()] = fmt.Sprintf("%s phải là một UUID hợp lệ", e.Field())
			case "slug":
				errors[e.Field()] = fmt.Sprintf("%s chỉ chữ cái thường,số,dấu - .", e.Field())
			case "min":
				errors[e.Field()] = fmt.Sprintf("%s phải lớn hơn %s", e.Field(), e.Param())
			case "max":
				errors[e.Field()] = fmt.Sprintf("%s phải nhỏ hơn %s", e.Field(), e.Param())
			case "oneof":
				allowdValues := strings.Join(strings.Split(e.Param(), " "), ",")
				errors[e.Field()] = fmt.Sprintf("%s phải là 1 trong các giá trị %s", e.Field(), allowdValues)
			case "required":
				errors[e.Field()] = fmt.Sprintf("Trường %s bắt buộc phải nhập", e.Field())
			case "search":
				errors[e.Field()] = fmt.Sprintf("Trường %s không nhập được các kí tự đặc biệt", e.Field())
			case "email":
				errors[e.Field()] = fmt.Sprintf("Trường %s phải đúng định dạng", e.Field())
			case "email_advanced":
				errors[e.Field()] = fmt.Sprintf("%s email này trong danh sách bị cấm", e.Value())
			case "datetime":
				errors[e.Field()] = fmt.Sprintf("Trường %s phải đúng định dạng YYYY-MM-DD", e.Field())
			case "password_strong":
				errors[e.Field()] = fmt.Sprintf(" %s phải có ít nhất 8 kí tự phải (chữ thường,chữ hoa,số và kí tự đặc biệt)", e.Field())
			case "file_ext":
				allowdValues := strings.Join(strings.Split(e.Param(), " "), ",")
				errors[e.Field()] = fmt.Sprintf("Trường %s phải có phần mở rộng thuộc %s", e.Field(), allowdValues)
			}
		}
		return gin.H{"error": errors}
	}
	return gin.H{
		"error": "Validation failed- yêu cầu không hợp lệ " ,
		"details": err.Error(),
	}
}