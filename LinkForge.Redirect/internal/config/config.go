package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL        string
	RedisURL           string
	Port               string
	LogLevel           string
	CORSAllowedOrigins string
	RateLimitRPS       float64
	RateLimitBurst     int
	AdminAPIKey        string
	KafkaBrokers       []string
	KafkaTopic         string
	KafkaGroupID       string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("KAFKA_GROUP_ID", "redirect-service")
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("RATE_LIMIT_RPS", 100)
	viper.SetDefault("RATE_LIMIT_BURST", 200)

	cfg := &Config{
		DatabaseURL:        viper.GetString("DATABASE_URL"),
		RedisURL:           viper.GetString("REDIS_URL"),
		Port:               viper.GetString("PORT"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		CORSAllowedOrigins: viper.GetString("CORS_ALLOWED_ORIGINS"),
		RateLimitRPS:       viper.GetFloat64("RATE_LIMIT_RPS"),
		RateLimitBurst:     viper.GetInt("RATE_LIMIT_BURST"),
		AdminAPIKey:        viper.GetString("ADMIN_API_KEY"),
		KafkaBrokers:       strings.Split(viper.GetString("KAFKA_BROKERS"), ","),
		KafkaTopic:         viper.GetString("KAFKA_TOPIC"),
		KafkaGroupID:       viper.GetString("KAFKA_GROUP_ID"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}
	if cfg.KafkaTopic == "" {
		return nil, fmt.Errorf("KAFKA_TOPIC is required")
	}
	if len(cfg.KafkaBrokers) == 0 || cfg.KafkaBrokers[0] == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}

	return cfg, nil
}
