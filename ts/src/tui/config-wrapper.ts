import fs from 'fs';
import path from 'path';
import os from 'os';
import YAML from 'yaml';

// Duplicating some functionality from config/index.ts to avoid import issues
export interface Config {
  defaultModel: string;
  models: {
    [key: string]: string;
  };
  historyFile: string;
}

const defaultConfig: Config = {
  defaultModel: 'gpt-4o',
  models: {
    'gpt-4': 'gpt-4',
    'gpt-4o': 'gpt-4o',
    'gpt-3.5': 'gpt-3.5-turbo',
    'claude-3-opus': 'claude-3-opus-20240229',
    'claude-3-sonnet': 'claude-3-sonnet-20240229',
    'claude-3-haiku': 'claude-3-haiku-20240307',
    'gemini-pro': 'gemini-pro',
    'gemini-ultra': 'gemini-1.5-pro-latest',
    'deepseek': 'deepseek-chat',
    'deepseek-coder': 'deepseek-coder',
    'mistral': 'mistral:latest',
    'llama3': 'llama3:latest'
  },
  historyFile: path.join(os.homedir(), '.ask-ai', 'history.db')
};

export function loadConfig(): Config {
  try {
    const xdgConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
    const configDir = path.join(xdgConfigPath, 'ask-ai');
    const configFile = path.join(configDir, 'config.yaml');

    if (fs.existsSync(configFile)) {
      const fileContents = fs.readFileSync(configFile, 'utf8');
      return { ...defaultConfig, ...YAML.parse(fileContents) };
    }
  } catch (error) {
    console.error('Error loading config:', error);
  }

  return defaultConfig;
}

export function getModelName(shortName: string, config: Config): string {
  return config.models[shortName] || shortName;
}

export function getDefaultModel(config: Config): string {
  return config.defaultModel;
}