package config

import "os"

type Config struct {
	AnthropicAPIKey string
	OllamaURL       string
	OllamaModel     string
}

func Load() Config {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3.2"
	}
	return Config{
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		OllamaURL:       ollamaURL,
		OllamaModel:     ollamaModel,
	}
}
