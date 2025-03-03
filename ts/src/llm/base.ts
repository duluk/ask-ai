import { EventEmitter } from 'events';
import { LLMClient, LLMConfig, Message, LLMResponse, StreamingResponse } from './types';

export class LLMStreamEmitter extends EventEmitter implements StreamingResponse { }

export abstract class BaseLLMClient implements LLMClient {
    protected config: LLMConfig;

    constructor(config: LLMConfig) {
        this.config = {
            modelName: config.modelName,
            maxTokens: config.maxTokens || 2048,
            temperature: config.temperature || 0.7,
            apiKey: config.apiKey,
            stream: config.stream || false
        };
    }

    abstract send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse>;

    abstract sendStream(messages: Message[], config?: Partial<LLMConfig>): StreamingResponse;

    getModelName(): string {
        return this.config.modelName;
    }

    async isAvailable(): Promise<boolean> {
        return !!this.config.apiKey;
    }

    protected getApiKey(provider: string): string | undefined {
        if (this.config.apiKey) {
            return this.config.apiKey;
        }

        const envKey = process.env[`${provider.toUpperCase()}_API_KEY`];
        if (envKey) {
            return envKey;
        }

        const homeDir = process.env.HOME || process.env.USERPROFILE;
        const path = require('path');
        const configDir = process.env.XDG_CONFIG_HOME || path.join(homeDir, '.config');
        if (configDir) {
            const fs = require('fs');
            const path = require('path');

            const keyPath = path.join(configDir, 'ask-ai', `${provider.toLowerCase()}-api-key`);
            try {
                if (fs.existsSync(keyPath)) {
                    return fs.readFileSync(keyPath, 'utf8').trim();
                }
            } catch (error) {
                console.error(`Error reading API key from ${keyPath}:`, error);
            }
        }

        return undefined;
    }
}
