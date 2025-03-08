import { createLLMClient } from '@/llm/factory.js';
import { OpenAIClient } from '@/llm/openai.js';
import { AnthropicClient } from '@/llm/anthropic.js';
import { ModelProvider } from '@/llm/types.js';

jest.mock('@/llm/openai.js');
jest.mock('@/llm/anthropic.js');

describe('LLM Factory', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should create OpenAI client for gpt models', () => {
    const client = createLLMClient('gpt-4');
    expect(client).toBeInstanceOf(OpenAIClient);
  });

  it('should create Anthropic client for Claude models', () => {
    const client = createLLMClient('claude-3-haiku');
    expect(client).toBeInstanceOf(AnthropicClient);
  });

  it('should default to OpenAI client for unknown models', () => {
    const client = createLLMClient('unknown-model');
    expect(client).toBeInstanceOf(OpenAIClient);
  });

  it('should pass configuration to the client', () => {
    const config = {
      modelName: 'gpt-4',
      maxTokens: 500,
      temperature: 0.5,
      apiKey: 'test-key'
    };
    
    const client = createLLMClient(config.modelName, config);
    expect(OpenAIClient).toHaveBeenCalledWith(expect.objectContaining(config));
  });
});