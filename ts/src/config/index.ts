import fs from 'fs';
import path from 'path';
import yaml from 'yaml';
import os from 'os';

export interface Config {
  defaultModel: string;
  modelsMap: Record<string, string>;
  historyFile: string;
  maxHistoryItems: number;
}

const DEFAULT_CONFIG: Config = {
  defaultModel: 'gpt-3.5-turbo',
  modelsMap: {
    'chatgpt': 'gpt-3.5-turbo',
    'gpt4': 'gpt-4',
    'claude': 'claude-3-haiku-20240307',
    'gemini': 'gemini-pro',
    'llama2': 'llama2:latest',
    'mistral': 'mistral:latest',
    'deepseek': 'deepseek-coder:latest'
  },
  historyFile: path.join(os.homedir(), '.config', 'ask-ai', 'history.db'),
  maxHistoryItems: 100
};

export function loadConfig(): Config {
  const configDir = path.join(os.homedir(), '.config', 'ask-ai');
  const configPath = path.join(configDir, 'config.yml');
  
  try {
    if (!fs.existsSync(configDir)) {
      fs.mkdirSync(configDir, { recursive: true });
    }
    
    if (!fs.existsSync(configPath)) {
      return DEFAULT_CONFIG;
    }
    
    const fileContents = fs.readFileSync(configPath, 'utf8');
    const userConfig = yaml.parse(fileContents);
    
    return {
      ...DEFAULT_CONFIG,
      ...userConfig
    };
  } catch (error) {
    console.error('Error loading config:', error);
    return DEFAULT_CONFIG;
  }
}

export function getModelName(modelAlias: string, config: Config): string {
  // If it's a model name in the map, use the mapped value
  if (config.modelsMap[modelAlias]) {
    return config.modelsMap[modelAlias];
  }
  
  // Otherwise return the original string (might be a direct model name)
  return modelAlias;
}

export function getDefaultModel(config: Config): string {
  return config.defaultModel;
}