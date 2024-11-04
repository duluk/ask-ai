# ask-ai

## Description

The utility everyone has written for themselves, a basic CLI tool for asking LLMs questions without bothering with a mouse.

Full disclosure: this is my first Go project. I mainly write in Ruby, C and Python.

## Installation

Just build the program and run it:

```bash
$ go mod tidy
$ go build ask-ai.go
```

Or, as I'm doing now (bc I'm old):
```bash
$ make build
```

## Usage

#### Set the API Key
1. Set {OPENAI,ANTHROPIC}_API_KEY in your environment; or
1. Put the key in a file located at `$HOME/.config/ask-ai/{openai,anthropic,google}-api-key`

#### Ask a model a question
```bash
$ bin/ask-ai "What is the best chess opening for a beginner?"
```

* If no query is provided, ask-ai will prompt for one:
```
$ bin/ask-ai
> What is the best chess opening for a checkers player?
```

* You can provide a model with `--model <model>`:
```bash
$ bin/ask-ai --model gemini "Why do you pull in so many modules for th Go API?"
```

### [NOTE]
> This is a work in progress and not all functionality has been added.
>
> I also plan to provide a flag for context - that is, use the last `n` queries to send to the LLM for context.
