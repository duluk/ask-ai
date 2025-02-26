import { AnthropicClient } from '@/llm/anthropic';
import { LLMConfig, Message } from '@/llm/types';
import Anthropic from '@anthropic-ai/sdk';

jest.mock('@anthropic-ai/sdk');
jest.mock('fs');

describe('Anthropic Client', () => {
  const mockConfig: LLMConfig = {
    modelName: 'claude-3-haiku-20240307',
    maxTokens: 1000,
    temperature: 0.7,
    apiKey: 'test-api-key'
  };
  
  const mockMessages: Message[] = [
    { role: 'system', content: 'You are a helpful assistant' },
    { role: 'user', content: 'Hello!' }
  ];

  const mockResponse = {
    content: [{ text: 'Hello! How can I assist you today?' }],
    usage: {
      input_tokens: 15,
      output_tokens: 8
    }
  };

  beforeEach(() => {
    jest.clearAllMocks();
    process.env.ANTHROPIC_API_KEY = 'env-api-key';
    
    // Mock Anthropic response
    (Anthropic as unknown as jest.Mock).mockImplementation(() => ({
      messages: {
        create: jest.fn().mockResolvedValue(mockResponse)
      }
    }));
  });

  afterEach(() => {
    delete process.env.ANTHROPIC_API_KEY;
  });

  it('should initialize with provided API key', () => {
    const client = new AnthropicClient(mockConfig);
    expect(Anthropic).toHaveBeenCalledWith({ apiKey: 'test-api-key' });
  });

  it('should fall back to environment API key', () => {
    const config = { ...mockConfig };
    delete config.apiKey;
    
    const client = new AnthropicClient(config);
    expect(Anthropic).toHaveBeenCalledWith({ apiKey: 'env-api-key' });
  });

  it('should send messages and process response', async () => {
    const client = new AnthropicClient(mockConfig);
    const response = await client.send(mockMessages);
    
    expect(response).toEqual({
      content: 'Hello! How can I assist you today?',
      usage: {
        promptTokens: 15,
        completionTokens: 8,
        totalTokens: 23
      }
    });
  });

  it('should handle system messages correctly', async () => {
    const client = new AnthropicClient(mockConfig);
    await client.send(mockMessages);
    
    const anthropicClient = (Anthropic as unknown as jest.Mock).mock.results[0].value;
    const createMethod = anthropicClient.messages.create;
    
    expect(createMethod).toHaveBeenCalledWith(expect.objectContaining({
      system: 'You are a helpful assistant',
      messages: [{ role: 'user' as const, content: 'Hello!' }]
    }));
  });

  it('should throw error if client is not initialized', async () => {
    delete process.env.ANTHROPIC_API_KEY;
    const config = { ...mockConfig };
    delete config.apiKey;
    
    // Mock fs.existsSync to return false to simulate no API key file
    const fs = require('fs');
    (fs.existsSync as jest.Mock).mockReturnValue(false);
    
    const client = new AnthropicClient(config);
    
    // Mock the client to be null
    Object.defineProperty(client, 'client', { value: null });
    
    await expect(client.send(mockMessages)).rejects.toThrow('Anthropic client not initialized');
  });

  it('should merge config with instance config', async () => {
    const client = new AnthropicClient(mockConfig);
    const sendConfig = { maxTokens: 500, temperature: 0.5 };
    
    await client.send(mockMessages, sendConfig);
    
    const anthropicClient = (Anthropic as unknown as jest.Mock).mock.results[0].value;
    const createMethod = anthropicClient.messages.create;
    
    expect(createMethod).toHaveBeenCalledWith(expect.objectContaining({
      model: 'claude-3-haiku-20240307',
      max_tokens: 500,
      temperature: 0.5,
      messages: expect.any(Array)
    }));
  });
});