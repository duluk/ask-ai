# ask-ai (TypeScript version)

A TypeScript implementation of the ask-ai CLI tool for asking LLMs questions without a GUI.

## Installation

```bash
# From the ts directory
npm install
```

## Usage

### Command Line Interface

```bash
# Build the TypeScript code
npm run build

# Run the CLI
npm start -- "What is the best chess opening for a beginner?"

# Or run directly
node dist/cli/index.js "What is the best chess opening for a beginner?"

# Run in development mode
npm run dev -- "What is the best chess opening for a beginner?"
```

### Text User Interface (TUI)

For a more interactive experience, use the TUI:

```bash
# Build and run the TUI
npm run build
npm run start:tui

# Or run in development mode
npm run dev:tui
```

The TUI includes:
- An output box for viewing the conversation
- A status line showing the current model and application state
- An input box with Emacs-style keybindings
- Support for slash commands

#### TUI Keyboard Shortcuts

- `Ctrl+C` - Exit the application
- `Ctrl+A` - Move to beginning of line
- `Ctrl+E` - Move to end of line
- `Ctrl+U` - Clear the line
- `Ctrl+W` - Delete word before cursor

#### TUI Slash Commands

- `/help` - Show help
- `/exit` or `/quit` - Exit the application
- `/model [name]` - Show or change the current model
- `/clear` - Clear the current conversation

## Development

```bash
# Run tests
npm test

# Lint code
npm run lint

# Format code
npm run format
```

## Supported LLM Providers

- OpenAI (ChatGPT, GPT-4)
- Anthropic (Claude)
- Google (Gemini)
- Ollama (Llama, Mistral, Phi, etc.)
- DeepSeek

## Environment Variables

The following environment variables can be set to configure API keys:

- `OPENAI_API_KEY` - For OpenAI models
- `ANTHROPIC_API_KEY` - For Anthropic Claude models
- `GOOGLE_API_KEY` - For Google Gemini models
- `DEEPSEEK_API_KEY` - For DeepSeek models
- `OLLAMA_BASE_URL` - For Ollama models (defaults to http://localhost:11434)

Alternatively, you can store API keys in files at:
`~/.config/ask-ai/{openai,anthropic,google,deepseek}-api-key`