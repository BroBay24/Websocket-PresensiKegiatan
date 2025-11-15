package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port string
	DB   DBConfig
}

type DBConfig struct {
	Host    string
	Port    string
	User    string
	Password string
	Name    string
	LogMode bool
}

func Load() Config {
	cfg := Config{
		Port: getEnv("PORT", "3000"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "presensi_qr"),
			LogMode:  getEnvAsBool("DB_LOG_MODE", false),
		},
	}

	return cfg
}

func getEnv(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func getEnvAsBool(key string, def bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("invalid boolean for %s, fallback to %t", key, def)
		return def
	}
	return parsed
}
