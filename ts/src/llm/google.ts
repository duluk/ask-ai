import { BaseLLMClient } from './base';
import { LLMConfig, Message, LLMResponse } from './types';
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