import { LLMClient, LLMConfig, ModelProvider } from './types';
import { OpenAIClient } from './openai';
import { AnthropicClient } from './anthropic';
import { GoogleClient } from './google';
import { OllamaClient } from './ollama';
import { DeepSeekClient } from './deepseek';

/**
 * Creates and returns the appropriate LLM client based on the model name
 */
export function createLLMClient(modelName: string, config: Partial<LLMConfig> = {}): LLMClient {
    const provider = determineProvider(modelName);

    const clientConfig: LLMConfig = {
        modelName,
        maxTokens: config.maxTokens,
        temperature: config.temperature,
        apiKey: config.apiKey
    };

    switch (provider) {
        case ModelProvider.Anthropic:
            return new AnthropicClient(clientConfig);

        case ModelProvider.Google:
            return new GoogleClient(clientConfig);

        case ModelProvider.Ollama:
            return new OllamaClient(clientConfig);

        case ModelProvider.DeepSeek:
            return new DeepSeekClient(clientConfig);

        case ModelProvider.OpenAI:
        default:
            return new OpenAIClient(clientConfig);
    }
}

/**
 * Determines the LLM provider based on the model name
 */
// TODO: Have arrays of all supported models by each provider and compare to that
function determineProvider(modelName: string): ModelProvider {
    const lowerCaseModelName = modelName.toLowerCase();

    if (lowerCaseModelName.includes('claude')) {
        return ModelProvider.Anthropic;
    }

    if (lowerCaseModelName.includes('gemini')) {
        return ModelProvider.Google;
    }

    if (lowerCaseModelName.includes('llama') ||
        lowerCaseModelName.includes('mistral') ||
        lowerCaseModelName.includes('phi')) {
        return ModelProvider.Ollama;
    }

    if (lowerCaseModelName.includes('deepseek')) {
        return ModelProvider.DeepSeek;
    }

    // Default
    return ModelProvider.OpenAI;
}
