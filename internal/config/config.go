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
	))

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.expiry", "30m")
	viper.SetDefault("app.mode", "development")
	viper.SetDefault("minimax.model_name", "abab3.0-chat")
	viper.SetDefault("minimax.api_base_url", "https://api.minimax.chat/v1")
	viper.SetDefault("sensenova.model_name", "SenseChat")
	viper.SetDefault("sensenova.api_base_url", "https://api.sensenova.cn/v1")

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
