import { BaseLLMClient } from './base';
import { LLMConfig, Message, LLMResponse, StreamingResponse } from './types';
import OpenAI from 'openai';

// DeepSeek uses the OpenAI API format
export class DeepSeekClient extends BaseLLMClient {
    private client: OpenAI | null = null;

    constructor(config: LLMConfig) {
        super(config);

        const apiKey = this.getApiKey('deepseek');
        if (apiKey) {
            this.client = new OpenAI({
                apiKey: apiKey,
                baseURL: 'https://api.deepseek.com/v1'  // DeepSeek API endpoint
            });
        }
    }

    sendStream(messages: Message[], config?: Partial<LLMConfig>): StreamingResponse {
        throw new Error('Streaming not implemented for DeepSeekClient');
    }

    async send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse> {
        if (!this.client) {
            throw new Error('DeepSeek client not initialized. API key may be missing.');
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
            console.error('Error calling DeepSeek API:', error);
            throw error;
        }
    }

    async isAvailable(): Promise<boolean> {
        return !!this.client;
    }
}
