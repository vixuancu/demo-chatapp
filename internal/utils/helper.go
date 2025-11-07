package utils

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

func GetIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return valueInt
}

// ParseQueryInt parses an integer query parameter with a default value
func ParseQueryInt(c *gin.Context, key string, defaultValue int) int {
	value := c.DefaultQuery(key, strconv.Itoa(defaultValue))
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
