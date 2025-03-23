package common

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Version is the current version of the AI toolkit
const Version = "1.0.0"

// DefaultAIModel is the default Gemini model to use
const DefaultAIModel = "gemini-2.0-flash"

// LoadEnv loads environment variables from .env file and cleans Windows line endings
func LoadEnv() {
	// Load .env file if it exists (ignore errors)
	_ = godotenv.Load()
	
	// Clean Windows line endings from key environment variables
	CleanWindowsLineEndings()
}

// CleanWindowsLineEndings removes carriage returns from common environment variables
func CleanWindowsLineEndings() {
	// Common environment variables used across tools
	keysToClean := []string{
		"GEMINI_API_KEY", 
		"DEFAULT_LANG", 
		"DEFAULT_MODEL", 
		"DEFAULT_TEMPERATURE", 
		"DEFAULT_TIMEOUT", 
		"DEFAULT_VERBOSE",
	}
	
	for _, key := range keysToClean {
		value := os.Getenv(key)
		if value != "" {
			// Remove carriage returns
			cleanValue := strings.ReplaceAll(value, "\r", "")
			if cleanValue != value {
				os.Setenv(key, cleanValue)
			}
		}
	}
}

// GetEnvOrDefault gets an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvOrDefaultInt gets an environment variable as an int or returns a default value
func GetEnvOrDefaultInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetEnvOrDefaultFloat gets an environment variable as a float or returns a default value
func GetEnvOrDefaultFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

// GetEnvOrDefaultBool gets an environment variable as a bool or returns a default value
func GetEnvOrDefaultBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}