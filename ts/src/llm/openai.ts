import { BaseLLMClient } from './base';
import { LLMConfig, Message, LLMResponse } from './types';
import OpenAI from 'openai';

export class OpenAIClient extends BaseLLMClient {
  private client: OpenAI | null = null;
  
  constructor(config: LLMConfig) {
    super(config);
    
    const apiKey = this.getApiKey('openai');
    if (apiKey) {
      this.client = new OpenAI({
        apiKey: apiKey
      });
    }
  }
  
  async send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse> {
    if (!this.client) {
      throw new Error('OpenAI client not initialized. API key may be missing.');
    }
    
    const mergedConfig = {
      ...this.config,
      ...config
    };
    
    try {
      const response = await this.client.chat.completions.create({
        model: mergedConfig.modelName,
        messages: messages,
        max_tokens: mergedConfig.maxTokens,
        temperature: mergedConfig.temperature,
      });
      
      const usage = response.usage;
      
      return {
        content: response.choices[0]?.message.content || '',
        finishReason: response.choices[0]?.finish_reason,
        usage: {
          promptTokens: usage?.prompt_tokens,
          completionTokens: usage?.completion_tokens,
          totalTokens: usage?.total_tokens
        }
      };
    } catch (error) {
      console.error('Error calling OpenAI API:', error);
      throw error;
    }
  }
  
  async isAvailable(): Promise<boolean> {
    return !!this.client;
  }
}