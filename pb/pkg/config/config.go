package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	// Telegram
	TgAPIID       int
	TgAPIHash     string
	TgPhone       string
	TgSessionPath string
	TargetChatIDs []int64 // Whitelisted channel/chat IDs

	// MeiliSearch
	MeiliHost   string
	MeiliMasterKey string

	// OpenAI
	OpenAIAPIKey  string
	OpenAIBaseURL string
}

// Load reads configuration from environment variables.
func Load() *Config {
	apiID, _ := strconv.Atoi(os.Getenv("TG_API_ID"))

	return &Config{
		// Telegram
		TgAPIID:       apiID,
		TgAPIHash:     os.Getenv("TG_API_HASH"),
		TgPhone:       os.Getenv("TG_PHONE"),
		TgSessionPath: getEnvOrDefault("TG_SESSION_PATH", "session.json"),
		TargetChatIDs: parseIntList(os.Getenv("TARGET_CHAT_IDS")),

		// MeiliSearch
		MeiliHost:   getEnvOrDefault("MEILI_HOST", "http://localhost:7700"),
		MeiliMasterKey: os.Getenv("MEILI_MASTER_KEY"),

		// OpenAI
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL: os.Getenv("OPENAI_BASE_URL"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// parseIntList parses a comma-separated list of int64 values.
func parseIntList(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.ParseInt(p, 10, 64); err == nil {
			result = append(result, v)
		}
	}
	return result
}

// IsChatAllowed checks if the given chat ID is in the whitelist.
// If no whitelist is configured, all chats are allowed.
func (c *Config) IsChatAllowed(chatID int64) bool {
	if len(c.TargetChatIDs) == 0 {
		return true
	}
	for _, id := range c.TargetChatIDs {
		if id == chatID {
			return true
		}
	}
	return false
}
