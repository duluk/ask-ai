import { EventEmitter } from 'events';
import { spawn } from 'child_process';
import path from 'path';

// Define types that mirror the ones in ../llm/types.ts
export interface Message {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

export interface StreamingResponse extends EventEmitter {
  on(event: 'data', listener: (chunk: string) => void): this;
  on(event: 'done', listener: (response: any) => void): this;
  on(event: 'error', listener: (error: Error) => void): this;
}

export interface LLMClient {
  send(messages: Message[]): Promise<any>;
  sendStream(messages: Message[]): StreamingResponse;
  getModelName(): string;
}

// Create a simple client that uses the CLI as a subprocess
export class CliClient implements LLMClient {
  private modelName: string;
  private cliPath: string;

  constructor(modelName: string) {
    this.modelName = modelName;
    // Path to the CLI executable
    this.cliPath = path.resolve(__dirname, '..', 'cli', 'index.js');
  }

  async send(messages: Message[]): Promise<any> {
    const lastMessage = messages[messages.length - 1];
    return new Promise((resolve, reject) => {
      const proc = spawn('node', [
        this.cliPath,
        lastMessage.content,
        '--model', this.modelName,
        '--quiet'
      ]);

      let output = '';
      proc.stdout.on('data', (data) => {
        output += data.toString();
      });

      proc.stderr.on('data', (data) => {
        console.error(`Error: ${data}`);
      });

      proc.on('close', (code) => {
        if (code === 0) {
          resolve({ content: output, finishReason: 'stop' });
        } else {
          reject(new Error(`Process exited with code ${code}`));
        }
      });
    });
  }

  sendStream(messages: Message[]): StreamingResponse {
    const emitter = new EventEmitter() as StreamingResponse;
    const lastMessage = messages[messages.length - 1];

    const proc = spawn('node', [
      this.cliPath,
      lastMessage.content,
      '--model', this.modelName,
      '--stream'
    ]);

    let fullContent = '';

    proc.stdout.on('data', (data) => {
      const chunk = data.toString();
      fullContent += chunk;
      emitter.emit('data', chunk);
    });

    proc.stderr.on('data', (data) => {
      console.error(`Error: ${data}`);
      emitter.emit('error', new Error(data.toString()));
    });

    proc.on('close', (code) => {
      if (code === 0) {
        emitter.emit('done', { 
          content: fullContent, 
          finishReason: 'stop',
          usage: { 
            promptTokens: 0, 
            completionTokens: 0 
          }
        });
      } else {
        emitter.emit('error', new Error(`Process exited with code ${code}`));
      }
    });

    return emitter;
  }

  getModelName(): string {
    return this.modelName;
  }
}

export function createLLMClient(modelName: string): LLMClient {
  return new CliClient(modelName);
}