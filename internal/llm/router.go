package llm

import (
	"strings"

	"github.com/xndi/fofus/config"
	"github.com/xndi/fofus/internal/pet"
)

func Route(cfg config.Config, msg string, s pet.State) (string, bool, error) {
	prompt := pet.ChatPrompt(s)
	if strings.HasPrefix(msg, "/smart ") {
		userMsg := strings.TrimPrefix(msg, "/smart ")
		reply, err := ClaudeAsk(cfg.AnthropicAPIKey, prompt, userMsg)
		return reply, true, err
	}
	reply, err := OllamaAsk(cfg.OllamaURL, cfg.OllamaModel, prompt, msg)
	return reply, false, err
}
