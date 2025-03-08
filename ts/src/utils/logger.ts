import { Config } from '../config/index.js';

import fs from 'node:fs';
import path from 'node:path';

export class Logger {
    private static instance: Logger;
    private logFile: fs.WriteStream;

    private constructor(logPath: string) {
        const dir = path.dirname(logPath);
        fs.mkdirSync(dir, { recursive: true });
        this.logFile = fs.createWriteStream(logPath, { flags: 'a' });
    }

    static initialize(logPath: string): Logger {
        if (!Logger.instance) {
            Logger.instance = new Logger(logPath);
        }
        return Logger.instance;
    }

    static getInstance(): Logger {
        if (!Logger.instance) {
            // TODO: would be nice to create the instance but need to get the
            // log path from config
            throw new Error('Logger not initialized. Call initialize() first.');
        }
        return Logger.instance;
    }

    log(level: string, message: string, metadata?: Record<string, any>) {
        // Return if log level is lower than config setting
        // const config = loadConfig();
        // if (level === 'debug' && !config.debug) {
        //     return;
        // }

        const timestamp = new Date().toISOString();
        const logEntry = {
            timestamp,
            level,
            message,
            ...metadata
        };
        this.logFile.write(`${JSON.stringify(logEntry)}\n`);
    }

    close() {
        this.logFile.end();
    }
}
