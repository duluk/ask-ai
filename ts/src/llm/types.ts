export interface LLMConfig {
  modelName: string;
  maxTokens?: number;
  temperature?: number;
  apiKey?: string;
}

export interface Message {
  role: 'system' | 'user' | 'assistant';
  content: string;
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

export interface LLMClient {
  send(messages: Message[], config?: Partial<LLMConfig>): Promise<LLMResponse>;
  getModelName(): string;
  isAvailable(): Promise<boolean>;
}

export enum ModelProvider {
  OpenAI = 'openai',
  Anthropic = 'anthropic',
  Google = 'google',
  Ollama = 'ollama',
  DeepSeek = 'deepseek'
}
