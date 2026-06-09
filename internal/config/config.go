package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServerURL  string
	BotUser    string
	BotPass    string
	MainAdmin  string
}

func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	if (s[0] == '\'' && s[len(s)-1] == '\'') ||
		(s[0] == '"' && s[len(s)-1] == '"') {
		return s[1 : len(s)-1]
	}
	return s
}

func Load() (*Config, error) {
	botPass := os.Getenv("ROCKETCHAT_BOT_PASSWORD")
	botPass = trimQuotes(botPass)
	if botPass == "" {
		if path := os.Getenv("ROCKETCHAT_BOT_PASSWORD_FILE"); path != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("reading ROCKETCHAT_BOT_PASSWORD_FILE: %w", err)
			}
			botPass = strings.TrimSpace(string(b))
			botPass = trimQuotes(botPass)
		}
	}

	cfg := &Config{
		ServerURL: os.Getenv("ROCKETCHAT_SERVER_URL"),
		BotUser:   os.Getenv("ROCKETCHAT_BOT_USERNAME"),
		BotPass:   botPass,
		MainAdmin: os.Getenv("ROCKETCHAT_MAIN_ADMIN"),
	}

	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("ROCKETCHAT_SERVER_URL is required")
	}
	if cfg.BotUser == "" {
		return nil, fmt.Errorf("ROCKETCHAT_BOT_USERNAME is required")
	}
	if cfg.BotPass == "" {
		return nil, fmt.Errorf("ROCKETCHAT_BOT_PASSWORD is required")
	}
	if cfg.MainAdmin == "" {
		return nil, fmt.Errorf("ROCKETCHAT_MAIN_ADMIN is required")
	}

	return cfg, nil
}
