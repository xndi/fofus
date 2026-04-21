package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/xndi/fofus/internal/pet"
)

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func OllamaIdleQuip(baseURL, model string) (string, error) {
	return OllamaAsk(baseURL, model, pet.BubblePrompt(), "what are you thinking right now?")
}

func OllamaAsk(baseURL, model, systemPrompt, userMsg string) (string, error) {
	full := systemPrompt + "\nUser: " + userMsg + "\nFofus:"
	body, _ := json.Marshal(ollamaRequest{
		Model:  model,
		Prompt: full,
		Stream: false,
	})
	resp, err := http.Post(baseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama unreachable: %w", err)
	}
	defer resp.Body.Close()
	var r ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.Response, nil
}
