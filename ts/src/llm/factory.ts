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
  // Determine which provider to use based on the model name
  const provider = determineProvider(modelName);
  
  // Create configuration for the client
  const clientConfig: LLMConfig = {
    modelName,
    maxTokens: config.maxTokens,
    temperature: config.temperature,
    apiKey: config.apiKey
  };
  
  // Return the appropriate client
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
  
  // Default to OpenAI
  return ModelProvider.OpenAI;
}