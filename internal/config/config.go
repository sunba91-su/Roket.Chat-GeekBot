package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServerURL string
	BotUser   string
	BotPass   string
	BotToken  string
	BotUserID string
	MainAdmin string
}

func Load() (*Config, error) {
	botPass := os.Getenv("ROCKETCHAT_BOT_PASSWORD")
	if botPass == "" {
		if path := os.Getenv("ROCKETCHAT_BOT_PASSWORD_FILE"); path != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("reading ROCKETCHAT_BOT_PASSWORD_FILE: %w", err)
			}
			botPass = strings.TrimSpace(string(b))
		}
	}

	cfg := &Config{
		ServerURL: os.Getenv("ROCKETCHAT_SERVER_URL"),
		BotUser:   os.Getenv("ROCKETCHAT_BOT_USERNAME"),
		BotPass:   botPass,
		BotToken:  os.Getenv("ROCKETCHAT_BOT_TOKEN"),
		BotUserID: os.Getenv("ROCKETCHAT_BOT_USER_ID"),
		MainAdmin: os.Getenv("ROCKETCHAT_MAIN_ADMIN"),
	}

	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("ROCKETCHAT_SERVER_URL is required")
	}
	if cfg.BotUser == "" {
		return nil, fmt.Errorf("ROCKETCHAT_BOT_USERNAME is required")
	}

	hasToken := cfg.BotToken != "" && cfg.BotUserID != ""
	hasPass := cfg.BotPass != ""

	if !hasToken && !hasPass {
		return nil, fmt.Errorf("either ROCKETCHAT_BOT_PASSWORD or (ROCKETCHAT_BOT_TOKEN + ROCKETCHAT_BOT_USER_ID) must be set")
	}
	if cfg.MainAdmin == "" {
		return nil, fmt.Errorf("ROCKETCHAT_MAIN_ADMIN is required")
	}

	return cfg, nil
}
