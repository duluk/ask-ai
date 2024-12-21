# ask-ai

## Description

The utility everyone has written for themselves, a basic CLI tool for asking LLMs questions without bothering with a mouse.

Full disclosure: this is my first Go project. I mainly write in Ruby, C and Python.

## Building

Go:

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

### [NOTE]
> This is a work in progress and not all functionality has been added.
