import { EventEmitter } from 'events';

export interface Message {
    role: 'system' | 'user' | 'assistant';
    content: string;
}

export interface LLMConfig {
    modelName: string;
    maxTokens?: number;
    temperature?: number;
    apiKey?: string;
    stream?: boolean;
}

export interface StreamingResponse extends EventEmitter {
    on(event: 'data', listener: (chunk: string) => void): this;
    on(event: 'done', listener: (response: LLMResponse) => void): this;
    on(event: 'error', listener: (error: Error) => void): this;
}

export interface LLMClient {
    send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse>;
    sendStream(messages: Message[], config?: Partial<LLMConfig>): StreamingResponse;
    getModelName(): string;
    isAvailable(): Promise<boolean>;
}

export interface LLMResponse {
    content: string;
    finishReason?: string;
    usage?: {
        promptTokens?: number;
        completionTokens?: number;
        totalTokens?: number;
    };
}

export enum ModelProvider {
    OpenAI = 'openai',
    Anthropic = 'anthropic',
    Google = 'google',
    Ollama = 'ollama',
    DeepSeek = 'deepseek'
}
