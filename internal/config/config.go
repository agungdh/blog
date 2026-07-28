package config

import "os"

type Config struct {
	DBPath string
	Port   string
	AI     AIConfig
}

type AIConfig struct {
	Enabled  bool
	BaseURL  string
	APIKey   string
	Model    string
	Interval string
}

func Load() Config {
	return Config{
		DBPath: envOrDefault("DB_PATH", "blog.db"),
		Port:   envOrDefault("PORT", "8080"),
		AI: AIConfig{
			Enabled:  os.Getenv("AI_ENABLED") == "true",
			BaseURL:  os.Getenv("AI_BASE_URL"),
			APIKey:   os.Getenv("AI_API_KEY"),
			Model:    os.Getenv("AI_MODEL"),
			Interval: envOrDefault("AI_INTERVAL", "1h"),
		},
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
