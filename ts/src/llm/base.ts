import { LLMClient, LLMConfig, Message, LLMResponse } from './types';

export abstract class BaseLLMClient implements LLMClient {
  protected config: LLMConfig;

  constructor(config: LLMConfig) {
    this.config = {
      modelName: config.modelName,
      maxTokens: config.maxTokens || 2048,
      temperature: config.temperature || 0.7,
      apiKey: config.apiKey
    };
  }

  abstract send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse>;

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
    if (homeDir) {
      const fs = require('fs');
      const path = require('path');

      const keyPath = path.join(homeDir, '.config', 'ask-ai', `${provider.toLowerCase()}-api-key`);
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
