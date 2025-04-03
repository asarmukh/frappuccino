package config

import "os"

// Функция для получения переменных окружения с дефолтным значением
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
