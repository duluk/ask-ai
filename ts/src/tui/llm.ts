// This is a wrapper file for ESM imports
import { createLLMClient as originalCreateLLMClient } from '../llm/factory.js';
import { Message, LLMResponse, StreamingResponse } from '../llm/types.js';

export {
  Message,
  LLMResponse,
  StreamingResponse
};

export const createLLMClient = originalCreateLLMClient;