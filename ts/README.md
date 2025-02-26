# ask-ai (TypeScript version)

A TypeScript implementation of the ask-ai CLI tool for asking LLMs questions without a GUI.

## Installation

```bash
# From the ts directory
npm install
```

## Usage

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