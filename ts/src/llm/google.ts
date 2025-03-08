import { BaseLLMClient, LLMStreamEmitter } from './base.js';
import { LLMConfig, Message, LLMResponse, StreamingResponse } from './types.js';
import { GoogleGenerativeAI, GenerativeModel } from '@google/generative-ai';

export class GoogleClient extends BaseLLMClient {
    private client: GoogleGenerativeAI | null = null;
    private model: GenerativeModel | null = null;

    constructor(config: LLMConfig) {
        super(config);

        const apiKey = this.getApiKey('google');
        if (apiKey) {
            this.client = new GoogleGenerativeAI(apiKey);
            this.model = this.client.getGenerativeModel({ model: config.modelName });
        }
    }

    sendStream(messages: Message[], config?: Partial<LLMConfig>): StreamingResponse {
        const emitter = new LLMStreamEmitter();

        if (!this.client || !this.model) {
            process.nextTick(() => {
                emitter.emit('error', new Error('Google client not initialized. API key may be missing.'));
            });
            return emitter;
        }

        const model = this.model;

        const mergedConfig = {
            ...this.config,
            ...config
        };

        // Start the streaming process
        (async () => {
            try {
                // Format the prompt for streaming
                const prompt = messages[messages.length - 1].content;

                // Get system prompt if present
                const systemMessage = messages.find(msg => msg.role === 'system');

                // Create generation config
                const generationConfig = {
                    maxOutputTokens: mergedConfig.maxTokens,
                    temperature: mergedConfig.temperature
                };

                // Stream the response
                const result = await model.generateContentStream({
                    contents: [{ role: 'user', parts: [{ text: prompt }] }],
                    generationConfig,
                    ...(systemMessage && {
                        systemPrompt: systemMessage.content
                    })
                });

                let fullContent = '';
                for await (const response of result.stream) {
                    const text = response.text();
                    if (text) {
                        fullContent += text;
                        emitter.emit('data', text);
                    }
                }

                // Emit the final response
                emitter.emit('done', {
                    content: fullContent,
                    usage: {
                        promptTokens: undefined,
                        completionTokens: undefined,
                        totalTokens: undefined
                    }
                });

            } catch (error) {
                console.error('Error in Google stream:', error);
                emitter.emit('error', error);
            }
        })();

        return emitter;
    }

    async send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse> {
        if (!this.client || !this.model) {
            throw new Error('Google client not initialized. API key may be missing.');
        }

        const mergedConfig = {
            ...this.config,
            ...config
        };

        try {
            // Extract system message if present
            const systemMessage = messages.find(msg => msg.role === 'system');

            // Format chat history for Google's API
            const chatHistory = messages
                .filter(msg => msg.role !== 'system')
                .map(msg => ({
                    role: msg.role === 'assistant' ? 'model' : 'user',
                    parts: [{ text: msg.content }]
                }));

            // Create chat session with generation config
            const chatConfig: any = {
                generationConfig: {
                    maxOutputTokens: mergedConfig.maxTokens,
                    temperature: mergedConfig.temperature
                }
            };

            // Add system instruction if available
            // Note: Different versions of the Google API may handle system prompts differently
            if (systemMessage?.content) {
                chatConfig.history = [
                    {
                        role: 'user',
                        parts: [{ text: `System instruction: ${systemMessage.content}` }]
                    }
                ];
            }

            const chat = this.model.startChat(chatConfig);

            // If there's existing chat history, replay it
            let lastResponse = null;
            if (chatHistory.length > 0) {
                for (let i = 0; i < chatHistory.length; i++) {
                    const message = chatHistory[i];
                    if (message.role === 'user') {
                        // For user messages, send to chat
                        lastResponse = await chat.sendMessage(message.parts[0].text);
                    }
                }
            }

            // Get the last user message and send it (if it wasn't already sent)
            const lastUserMessage = messages[messages.length - 1];
            if (lastUserMessage.role === 'user' && !lastResponse) {
                lastResponse = await chat.sendMessage(lastUserMessage.content);
            }

            if (!lastResponse) {
                throw new Error('No response from Gemini API');
            }

            return {
                content: lastResponse.response.text(),
                usage: {
                    // Google API may not provide token usage info in the same way
                    promptTokens: undefined,
                    completionTokens: undefined,
                    totalTokens: undefined
                }
            };
        } catch (error) {
            console.error('Error calling Google Gemini API:', error);
            throw error;
        }
    }

    async isAvailable(): Promise<boolean> {
        return !!this.client && !!this.model;
    }
}
