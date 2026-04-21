package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Mood string

const (
	MoodHappy   Mood = "happy"
	MoodNeutral Mood = "neutral"
	MoodSad     Mood = "sad"
)

type State struct {
	Hunger    int `json:"hunger"`
	Happiness int `json:"happiness"`
	Energy    int `json:"energy"`
}

func (s *State) Mood() Mood {
	if s.Hunger < 30 || s.Happiness < 30 {
		return MoodSad
	}
	if s.Hunger > 60 && s.Happiness > 60 {
		return MoodHappy
	}
	return MoodNeutral
}

func (s *State) Feed() {
	s.Hunger = clamp(s.Hunger+20, 0, 100)
}

func (s *State) Play() {
	s.Happiness = clamp(s.Happiness+15, 0, 100)
	s.Energy = clamp(s.Energy-10, 0, 100)
}

func (s *State) Tick() {
	s.Hunger = clamp(s.Hunger-5, 0, 100)
	s.Happiness = clamp(s.Happiness-2, 0, 100)
	s.Energy = clamp(s.Energy-1, 0, 100)
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func statePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "fofus")
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return filepath.Join(p, "state.json"), nil
}

func Load() State {
	path, err := statePath()
	if err != nil {
		return defaultState()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultState()
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return defaultState()
	}
	return s
}

func (s *State) Save() error {
	path, err := statePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func defaultState() State {
	return State{Hunger: 80, Happiness: 80, Energy: 80}
}
