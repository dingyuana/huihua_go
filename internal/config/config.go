package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Minimax   MinimaxConfig
	Sensenova SensenovaConfig
	Agnes     AgnesConfig
}

type AppConfig struct {
	// Mode gates destructive operations (e.g. clear-data APIs).
	// "production" blocks them; "development" (default) allows.
	Mode string
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return r.Host + ":" + r.Port
}

type JWTConfig struct {
	Secret string
	Expiry string
}

type MinimaxConfig struct {
	APIKey    string
	ModelName string
	APIBaseURL string
}

type SensenovaConfig struct {
	APIKey    string
	ModelName string
	APIBaseURL string
}

type AgnesConfig struct {
	APIKey    string
	ModelName string
	APIBaseURL string
}

func Load() *Config {
	viper.SetEnvPrefix("HF")
	viper.AutomaticEnv()
	// Map HF_DATABASE_* → database.*
	// HF_DATABASE_HOST → database.host, HF_DATABASE_PORT → database.port, etc.
	viper.SetEnvKeyReplacer(strings.NewReplacer(
		"DATABASE_", "database.",
		"REDIS_", "redis.",
		"JWT_", "jwt.",
		"SERVER_", "server.",
		"APP_", "app.",
		"MINIMAX_", "minimax.",
		"SENSENOVA_", "sensenova.",
		"AGNES_", "agnes.",
	))

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.expiry", "30m")
	viper.SetDefault("app.mode", "development")
	viper.SetDefault("minimax.model_name", "abab3.0-chat")
	viper.SetDefault("minimax.api_base_url", "https://api.minimax.chat/v1")
	viper.SetDefault("sensenova.model_name", "deepseek-v4-flash")
	viper.SetDefault("sensenova.api_base_url", "https://token.sensenova.cn/v1")
	viper.SetDefault("agnes.model_name", "agnes-2.0")
	viper.SetDefault("agnes.api_base_url", "https://www.agnes-ai.com/doc/agnes-2.0")

	// Try executable dir first, then current working directory
	exeDir := filepath.Dir(os.Args[0])
	paths := []string{exeDir, "."}
	if absExe, err := filepath.Abs(exeDir); err == nil {
		paths = []string{absExe, exeDir, "."}
	}
	for _, p := range paths {
		viper.AddConfigPath(p)
	}
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("[DEBUG] viper.ReadInConfig error (tried %v): %v", paths, err)
	} else {
		log.Printf("[DEBUG] viper config file: %s", viper.ConfigFileUsed())
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	log.Printf("[DEBUG] cfg.Database.Host=%q user=%q", cfg.Database.Host, cfg.Database.User)
	return &cfg
}
