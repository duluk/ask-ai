import { BaseLLMClient } from './base';
import { LLMConfig, Message, LLMResponse } from './types';
import Anthropic from '@anthropic-ai/sdk';

export class AnthropicClient extends BaseLLMClient {
  private client: Anthropic | null = null;

  constructor(config: LLMConfig) {
    super(config);

    const apiKey = this.getApiKey('anthropic');
    if (apiKey) {
      this.client = new Anthropic({
        apiKey: apiKey
      });
    }
  }

  async send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse> {
    if (!this.client) {
      throw new Error('Anthropic client not initialized. API key may be missing.');
    }

    const mergedConfig = {
      ...this.config,
      ...config
    };

    try {
      // Convert messages to Anthropic format
      const formattedMessages = messages
        .filter(msg => msg.role !== 'system')
        .map(msg => ({
          role: msg.role === 'assistant' ? 'assistant' as const : 'user' as const,
          content: msg.content
        }));

      // System message needs special handling
      const systemMessage = messages.find(msg => msg.role === 'system');
      const systemPrompt = systemMessage?.content;

      // Make sure max_tokens is a number and not undefined
      const maxTokens = mergedConfig.maxTokens || 1024;

      const response = await this.client.messages.create({
        model: mergedConfig.modelName,
        messages: formattedMessages,
        system: systemPrompt,
        max_tokens: maxTokens,
        temperature: mergedConfig.temperature,
      });

      return {
        content: response.content[0]?.text || '',
        usage: {
          // Anthropic may not provide detailed token usage in the same format
          promptTokens: response.usage?.input_tokens,
          completionTokens: response.usage?.output_tokens,
          totalTokens: (response.usage?.input_tokens || 0) + (response.usage?.output_tokens || 0)
        }
      };
    } catch (error) {
      console.error('Error calling Anthropic API:', error);
      throw error;
    }
  }

  async isAvailable(): Promise<boolean> {
    return !!this.client;
  }
}
