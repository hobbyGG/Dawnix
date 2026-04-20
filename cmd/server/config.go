package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/hobbyGG/Dawnix/internal/workflow/conf"
	"github.com/spf13/viper"
)

func LoadBootstrapConfig() (*conf.Bootstrap, error) {
	cfg := &conf.Bootstrap{}

	yamlViper := viper.New()
	yamlViper.SetConfigName("dev")
	yamlViper.SetConfigType("yaml")
	yamlViper.AddConfigPath("configs")
	yamlViper.AddConfigPath("../configs")
	yamlViper.AddConfigPath("../../configs")
	if err := yamlViper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read dev config failed: %w", err)
		}
	}
	if err := yamlViper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal dev config failed: %w", err)
	}

	cfg.ApplyDefaults()

	if err := overrideByEnv(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func overrideByEnv(cfg *conf.Bootstrap) error {
	envViper := viper.New()
	envViper.SetConfigName("local")
	envViper.SetConfigType("env")
	envViper.AddConfigPath(".")
	envViper.AddConfigPath("..")
	envViper.AddConfigPath("../..")
	envViper.AutomaticEnv()
	if err := envViper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("read local env failed: %w", err)
		}
	}

	if val := envViper.GetString("SMTP_TOKEN"); val != "" {
		cfg.Worker.SMTPToken = val
	}
	if val := envViper.GetString("SMTP_EMAIL"); val != "" {
		cfg.Worker.SMTPEmail = val
	}
	if val := envViper.GetString("REDIS_ADDR"); val != "" {
		cfg.Worker.RedisAddr = val
		cfg.Data.Redis.Addr = val
	}
	if val := envViper.GetString("DB_DSN"); val != "" {
		cfg.Data.Database.DSN = val
	}
	if val := envViper.GetBool("EMAIL_SERVICE_ENABLED"); val {
		cfg.Biz.Features.EmailService.Enabled = true
	}
	if val := os.Getenv("SMTP_TOKEN"); val != "" {
		cfg.Worker.SMTPToken = val
	}
	if cfg.Biz.Features.EmailService.Enabled && cfg.Worker.SMTPToken == "" {
		return fmt.Errorf("SMTP_TOKEN not found in config")
	}
	return nil
}
