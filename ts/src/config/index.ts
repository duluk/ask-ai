import fs from 'fs';
import path from 'path';
import yaml from 'yaml';
import os from 'os';

import { Logger } from '../utils/logger';

export interface Config {
    defaultModel: string;
    modelsMap: Record<string, string>;
    historyFile: string;
    maxHistoryItems: number;
    maxTokens: number;
    temperature: number;
    systemPrompt: string;
    logFile: string;
    stream: boolean;
    debug: boolean;
}

const XDGConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');

const DEFAULT_CONFIG: Config = {
    defaultModel: 'chatgpt-4o-latest',
    modelsMap: {
        'chatgpt': 'chatgpt-4o-latest',
        'gpt4': 'gpt-4',
        'claude': 'claude-3-7-sonnet-20250219',
        'gemini': 'gemini-2.0-flash-001',
        'llama2': 'llama2:latest',
        'mistral': 'mistral:latest',
        'deepseek': 'deepseek-coder:latest'
    },
    historyFile: path.join(XDGConfigPath, 'ask-ai', 'history.db'),
    maxHistoryItems: 100,
    maxTokens: 1024,
    temperature: 0.7,
    systemPrompt: "You are a helpful assistant who doesn't talk back",
    logFile: path.join(XDGConfigPath, 'ask-ai', 'ask-ai.log'),
    stream: false,
    debug: false
};

export function loadConfig(): Config {
    const configDir = path.join(XDGConfigPath, 'ask-ai');
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
    const logger = Logger.getInstance();
    logger.log('debug', 'getModelName:', { modelAlias, config })
    // If it's a model name in the map, use the mapped value
    if (config.modelsMap[modelAlias]) {
        return config.modelsMap[modelAlias];
    }

    // // Otherwise return the original string (might be a direct model name)
    // return modelAlias;

    // Actually, return "" if no match
    return "";
}

export function getDefaultModel(config: Config): string {
    return config.defaultModel;
}
