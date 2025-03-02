import fs from 'fs';
import path from 'path';

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
            throw new Error('Logger not initialized. Call initialize() first.');
        }
        return Logger.instance;
    }

    log(level: string, message: string, metadata?: Record<string, any>) {
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
