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

## Usage

* Set OPENAI_API_KEY in your environment or put it in a file located at `$HOME/.config/ask-ai/openai-api-key`

```bash
$ ./ask-ai "What is the best chess opening for a beginner?"
```

* If no query is provided, ask-ai will prompt for one:
```
$ ./ask-ai
> What is the best chess opening for a checkers player?
```

### [NOTE]
> This is a work in progress and not all functionality has been added.
> For instance, currently it supports only ChatGPT. I plan to add flags for
> other LLMs at some point.
>
> I also plan to provide a flag for context - that is, use the last `n` queries to send to the LLM for context.
