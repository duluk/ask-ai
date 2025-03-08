import { EventEmitter } from 'node:events';
import { LLMConfig, Message, LLMResponse, LLMClient, StreamingResponse } from './types.js';
import path from 'node:path';
import fs from 'node:fs';
import os from 'node:os';

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

        const homeDir = os.homedir();
        const configDir = process.env.XDG_CONFIG_HOME || path.join(homeDir, '.config');
        if (configDir) {
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
