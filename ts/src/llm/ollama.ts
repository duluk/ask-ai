import { BaseLLMClient, LLMStreamEmitter } from './base.js';
import { LLMConfig, Message, LLMResponse, StreamingResponse } from './types.js';
import axios from 'axios';

interface OllamaCompletionResponse {
    model: string;
    created_at: string;
    message: {
        role: string;
        content: string;
    };
    done: boolean;
    total_duration?: number;
    load_duration?: number;
    prompt_eval_count?: number;
    eval_count?: number;
    eval_duration?: number;
}

export class OllamaClient extends BaseLLMClient {
    private baseUrl: string;

    constructor(config: LLMConfig) {
        super(config);

        // Default to localhost:11434 if not specified
        this.baseUrl = process.env.OLLAMA_BASE_URL || 'http://localhost:11434';
    }

    sendStream(messages: Message[], config?: Partial<LLMConfig>): StreamingResponse {
        throw new Error('Streaming not implemented for OllamaClient');
    }

    async send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse> {
        const mergedConfig = {
            ...this.config,
            ...config
        };

        try {
            // Format messages for Ollama
            const ollamaMessages = messages.map(msg => ({
                role: msg.role,
                content: msg.content
            }));

            // Make request to Ollama API
            const response = await axios.post<OllamaCompletionResponse>(`${this.baseUrl}/api/chat`, {
                model: mergedConfig.modelName,
                messages: ollamaMessages,
                options: {
                    num_predict: mergedConfig.maxTokens,
                    temperature: mergedConfig.temperature
                },
                stream: false
            });

            return {
                content: response.data.message.content,
                usage: {
                    // Ollama provides different metrics
                    promptTokens: response.data.prompt_eval_count,
                    completionTokens: response.data.eval_count,
                    totalTokens: (response.data.prompt_eval_count || 0) + (response.data.eval_count || 0)
                }
            };
        } catch (error) {
            console.error('Error calling Ollama API:', error);
            throw error;
        }
    }

    async isAvailable(): Promise<boolean> {
        try {
            // Check if Ollama is running by making a request to the API
            const response = await axios.get(`${this.baseUrl}/api/tags`);
            return response.status === 200;
        } catch (error) {
            return false;
        }
    }
}
