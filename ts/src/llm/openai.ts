import { BaseLLMClient } from './base';
import { LLMConfig, Message, LLMResponse, StreamingResponse } from './types';
// Is this the correct way to do this?
import { LLMStreamEmitter } from './base';

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

    sendStream(messages: Message[], config?: Partial<LLMConfig>): StreamingResponse {
        const emitter = new LLMStreamEmitter();

        if (!this.client) {
            emitter.emit('error', new Error('OpenAI client not initialized. API key may be missing.'));
            return emitter;
        }

        // Create a non-null reference to make LSP happy
        const client = this.client;

        const mergedConfig = {
            ...this.config,
            ...config
        };

        (async () => {
            try {
                const stream = await client.chat.completions.create({
                    model: mergedConfig.modelName,
                    messages: messages,
                    max_tokens: mergedConfig.maxTokens,
                    temperature: mergedConfig.temperature,
                    stream: true,
                });

                let fullContent = '';
                for await (const chunk of stream) {
                    const content = chunk.choices[0]?.delta?.content || '';
                    fullContent += content;
                    emitter.emit('data', content);
                }

                emitter.emit('done', {
                    content: fullContent,
                    finishReason: 'stop',
                    usage: undefined // Token usage not available in streaming mode
                });
            } catch (error) {
                emitter.emit('error', error);
            }
        })();

        return emitter;
    }

    async isAvailable(): Promise<boolean> {
        return !!this.client;
    }
}
