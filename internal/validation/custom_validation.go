package validation

import (
	"path/filepath"
	"regexp"
	"strings"
	"chat-app/internal/utils"

	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidation(v *validator.Validate) error {
	// black list email 
	var blockedDomains = map[string]bool {
		"blacklist.com": true,
		"edu.vn": true,
		"abc.com": true,
	}
	v.RegisterValidation("email_advanced", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return false
		}
		domain:= utils.NormalizeString(parts[1])
		return !blockedDomains[domain]
	})
	// password strong
	v.RegisterValidation("password_strong", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}
		hasLower := regexp.MustCompile(`[a-z]`)
		hasUpper := regexp.MustCompile(`[A-Z]`)
		hasDigit := regexp.MustCompile(`[0-9]`)
		hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
		return hasLower.MatchString(password) && hasUpper.MatchString(password) && hasDigit.MatchString(password) && hasSpecial.MatchString(password)
	})
	// Validate phone number with regex
	var slugRegex = regexp.MustCompile("^[a-z0-9]+(?:[-.][a-z0-9]+)*$")
	v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return slugRegex.MatchString(value)
	})
	var searchRegex = regexp.MustCompile(`^[a-zA-Z0-9\s]+$`)
	v.RegisterValidation("search", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return searchRegex.MatchString(value)
	})
	// file extension jpg , mp4, png, jpeg
	v.RegisterValidation("file_ext", func(fl validator.FieldLevel) bool {
		fileName := fl.Field().String() // lấy giá trị từ trường
		allowedStr := fl.Param()        // lấy tham số từ tag binding
		if allowedStr == "" {
			return false
		}
		allowedExt := strings.Fields(allowedStr)
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".") // lấy phần mở rộng của tệp
		for _, allowd := range allowedExt {
			if ext == strings.ToLower(allowd) {
				return true
			}
		}
		return false
	})
	return nil
}