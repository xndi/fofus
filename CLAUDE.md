# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go run .          # run the TUI
go build -o fofus # build binary
go build ./...    # check all packages compile
```

No test suite yet. Verification is manual (see below).

## Architecture

fofus is a Bubble Tea TUI pet. The Bubble Tea model lives in `internal/tui/` and drives everything:

- `tui/model.go` — `Model` struct (pet state, chat log, input, thinking flag), `Init` returns the first tick cmd
- `tui/update.go` — key handling split into `handleNav` (normal mode) and `handleChatInput` (chat mode); LLM calls are dispatched as `tea.Cmd` goroutines returning `llmResponseMsg`
- `tui/view.go` — lipgloss layout; chat lines are word-wrapped at `contentWidth=36` before rendering to avoid truncation inside the border

Pet state (hunger/happiness/energy, 0–100) lives in `internal/pet/state.go` and is saved to `~/.config/fofus/state.json` on quit. Mood (happy/neutral/sad) is derived from stats and controls which ASCII frame `art.go` returns.

LLM routing in `internal/llm/router.go`: messages prefixed `/smart ` go to Claude (`claude.go` via Anthropic SDK, model `claude-haiku-4-5`); all other messages go to Ollama (`ollama.go` via plain `net/http` POST to `/api/generate`, stream=false). Both receive the same system prompt built from current pet stats and mood. Config (API key, Ollama URL/model) is read from env in `config/config.go`.

Stats decay every 15s via a `tea.Tick` cmd that re-schedules itself on each fire.

## Key Bindings

| Key     | Action                  |
|---------|-------------------------|
| `f`     | Feed (+20 hunger)       |
| `p`     | Play (+15 happiness)    |
| `i`     | Open chat input         |
| `Esc`   | Close chat input        |
| `Enter` | Send message            |
| `q`/`^C`| Quit + save state       |

`q` is blocked while `m.thinking` is true (waiting for LLM) to prevent accidental quit.

## LLM Backends

- **Ollama** (default): must be running locally (`ollama serve`), model `llama3.2` pulled. URL defaults to `http://localhost:11434`, overridable via `OLLAMA_URL` / `OLLAMA_MODEL`.
- **Claude** (`/smart` prefix): requires `ANTHROPIC_API_KEY` env var. Uses `claude-haiku-4-5`.

## Verification

1. `go run .` → bordered TUI with creature and stat bars appears
2. `f` → hunger bar fills, creature may change mood
3. `i` → type `hello` → Enter → Ollama responds (shown wrapped in pink)
4. `i` → type `/smart explain quantum entanglement` → Enter → Claude responds
5. Wait 30s → stats tick down, mood may shift, ASCII art updates
6. `q` → re-run → stats persist from last session

