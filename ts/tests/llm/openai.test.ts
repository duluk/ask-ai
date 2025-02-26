import { OpenAIClient } from '@/llm/openai';
import { LLMConfig, Message } from '@/llm/types';
import OpenAI from 'openai';

jest.mock('openai');
jest.mock('fs');

describe('OpenAI Client', () => {
  const mockConfig: LLMConfig = {
    modelName: 'gpt-4',
    maxTokens: 1000,
    temperature: 0.7,
    apiKey: 'test-api-key'
  };
  
  const mockMessages: Message[] = [
    { role: 'system', content: 'You are a helpful assistant' },
    { role: 'user', content: 'Hello!' }
  ];

  const mockResponse = {
    choices: [
      {
        message: { content: 'Hello there! How can I help you today?' },
        finish_reason: 'stop'
      }
    ],
    usage: {
      prompt_tokens: 20,
      completion_tokens: 10,
      total_tokens: 30
    }
  };

  beforeEach(() => {
    jest.clearAllMocks();
    process.env.OPENAI_API_KEY = 'env-api-key';
    
    // Mock OpenAI response
    (OpenAI as unknown as jest.Mock).mockImplementation(() => ({
      chat: {
        completions: {
          create: jest.fn().mockResolvedValue(mockResponse)
        }
      }
    }));
  });

  afterEach(() => {
    delete process.env.OPENAI_API_KEY;
  });

  it('should initialize with provided API key', () => {
    const client = new OpenAIClient(mockConfig);
    expect(OpenAI).toHaveBeenCalledWith({ apiKey: 'test-api-key' });
  });

  it('should fall back to environment API key', () => {
    const config = { ...mockConfig };
    delete config.apiKey;
    
    const client = new OpenAIClient(config);
    expect(OpenAI).toHaveBeenCalledWith({ apiKey: 'env-api-key' });
  });

  it('should send messages and process response', async () => {
    const client = new OpenAIClient(mockConfig);
    const response = await client.send(mockMessages);
    
    expect(response).toEqual({
      content: 'Hello there! How can I help you today?',
      finishReason: 'stop',
      usage: {
        promptTokens: 20,
        completionTokens: 10,
        totalTokens: 30
      }
    });
  });

  it('should throw error if client is not initialized', async () => {
    delete process.env.OPENAI_API_KEY;
    const config = { ...mockConfig };
    delete config.apiKey;
    
    // Mock fs.existsSync to return false to simulate no API key file
    const fs = require('fs');
    (fs.existsSync as jest.Mock).mockReturnValue(false);
    
    const client = new OpenAIClient(config);
    
    // Mock the client to be null
    Object.defineProperty(client, 'client', { value: null });
    
    await expect(client.send(mockMessages)).rejects.toThrow('OpenAI client not initialized');
  });

  it('should merge config with instance config', async () => {
    const client = new OpenAIClient(mockConfig);
    const sendConfig = { maxTokens: 500, temperature: 0.5 };
    
    await client.send(mockMessages, sendConfig);
    
    const openaiClient = (OpenAI as unknown as jest.Mock).mock.results[0].value;
    const createMethod = openaiClient.chat.completions.create;
    
    expect(createMethod).toHaveBeenCalledWith(expect.objectContaining({
      model: 'gpt-4',
      max_tokens: 500,
      temperature: 0.5
    }));
  });
});