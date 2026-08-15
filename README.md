# ask-ai

## Description

The utility everyone has written for themselves, a basic CLI tool for asking LLMs questions without bothering with a mouse.

Full disclosure: this is my first Go project. I mainly write in Ruby, C and Python.

## Building

```bash
$ go mod tidy
$ go build cmd/ask-ai/main.go
```

Or, as I'm doing now (bc I'm old):
```bash
$ make
```

## Installation

```bash
$ make install
```

## Usage

#### Set the API Key
1. Set {OPENAI,ANTHROPIC,GOOGLE,XAI}_API_KEY in your environment; or
1. Put the key in a file located at `$HOME/.config/ask-ai/{openai,anthropic,google,xai}-api-key`

#### Ask a model a question
```bash
$ bin/ask-ai "What is the best chess opening for a beginner?"
```

* If no query is provided, ask-ai will prompt for one:
```
$ bin/ask-ai
chatgpt> What is the best chess opening for a checkers player?
```

* You can provide a model with `--model <model>`:
```bash
$ bin/ask-ai --model gemini "Why do you pull in so many modules for th Go API?"
```

* Search the web for current context before asking any configured model. Set a
  [Brave Search API](https://brave.com/search/api/) key in `BRAVE_SEARCH_API_KEY`
  (or `web.api_key` in the configuration), then use `--web` (the default is five
  results):
```bash
$ BRAVE_SEARCH_API_KEY=... bin/ask-ai --web "What changed in Go this week?"
```
  Use `--web-results N` to include 1–20 sources. Results are added to the
  request as untrusted reference text; ask-ai stores only your original prompt.
  Use `--web=auto` to let the model first decide whether current web context is
  necessary; it will search only when it is useful.
  To make this the default, set `web.mode: auto` in the configuration file.
  Like provider API keys, `web.api_key` may be a command such as
  `op read op://vault/brave/api-key`; its output is used as the key.

* Continue the conversation
```bash
$ bin/ask-ai --model grok "When is your knowledge cut-off?"
<...>
$ bin/ask-ai --model grok --continue "So you're always mostly up to date?"
```

* Use last `n` queries for context:
```bash
$ bin/ask-ai --context 3 "What are the last 3 things we talked about?"
```

* Search conversation history for a previous chat:
```bash
$ bin/ask-ai --search "chess openings"
```

* Show a specific conversation:
```bash
$ bin/ask-ai --show 3
```

* Continue a specific conversation:
```bash
$ bin/ask-ai --id 42 "What about the Reti?"
```

* List configured role names:
```bash
$ bin/ask-ai --list-roles
```

### [NOTE]
> This is a work in progress and not all functionality has been added.
