package config

import (
	"time"
)

// AppConfig holds service configuration
type AppConfig struct {
	ServiceName string
	Environment string
	Debug       bool
	LogLevel    string

	Database DBConfig
	Server   ServerConfig
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host          string
	Port          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	ShutdownTimeout time.Duration
}

// DefaultAppConfig returns default configuration
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ServiceName: "",
		Environment: "development",
		Debug:       false,
		LogLevel:    "info",
		Database: DBConfig{
			Host: "localhost",
			Port: "3306",
			User: "jellyfin",
		},
		Server: ServerConfig{
			Host:       "0.0.0.0",
			Port:       "8000",
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout: 60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
	}
}
