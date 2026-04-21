# fofus

an LLM-powered terminal creature with long-term plans you won’t like

```
╭──────────────────────────────────────╮
│  fofus                               │
│                                      │
│         (◕‿◕)                        │
│         /|  |\                       │
│          |  |                        │
│                                      │
│  ♥ hunger  ████████░░  80%           │
│  ★ happy   ██████░░░░  60%           │
│  ⚡energy  █████████░  90%           │
│                                      │
╰──────────────────────────────────────╯
```


## Running fofus

**1. Install Ollama**
```bash
brew install ollama
ollama serve &
ollama pull llama3.2
```

**2. (Optional) Set Anthropic key to make fofus a nerdy boy with `/smart` responses**
```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

**3. Run fofus the greatest**
```bash
go run .
```

Or build a binary:
```bash
go build -o fofus .
./fofus
```


