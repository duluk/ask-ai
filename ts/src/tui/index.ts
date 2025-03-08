#!/usr/bin/env node

import { loadConfig, getModelName, getDefaultModel } from '../config/index.js';
import { Database } from '../db/sqlite.js';
import { Logger } from '../utils/logger.js';
import path from 'node:path';
import os from 'node:os';

interface TUIOptions {
    model?: string;
    debug?: boolean;
}

/**
 * This is a temporary redirect to allow us to keep our JSX code in a .tsx file
 * while still providing the main entry point in a .ts file.
 */
export function startTUI(options: TUIOptions = {}): Promise<void> {
    return new Promise<void>((resolve, reject) => {
        try {
            // Initialize config, database and logger
            const config = loadConfig();
            const db = new Database(config.historyFile);

            const xdgConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
            const logPath = path.join(xdgConfigPath, 'ask-ai', 'ask-ai.ts.log');
            Logger.initialize(logPath);
            const logger = Logger.getInstance();

            // Set the model name from options or default
            const modelName: string = options.model
                ? getModelName(options.model, config)
                : getDefaultModel(config);

            // Use dynamic import to load the components
            import('./ink-app.js').then(({ startInkApp }) => {
                // Start the app
                startInkApp(config, db, modelName, logger)
                    .then(resolve)
                    .catch(reject);
            }).catch(reject);
        } catch (error) {
            reject(error);
        }
    });
}
