// Simple logger implementation to avoid dependencies on winston
export class Logger {
  private static instance: Logger;
  private logPath: string;

  private constructor(logPath: string) {
    this.logPath = logPath;
  }

  static initialize(logPath: string): void {
    if (!Logger.instance) {
      Logger.instance = new Logger(logPath);
    }
  }

  static getInstance(): Logger {
    if (!Logger.instance) {
      throw new Error('Logger not initialized');
    }
    return Logger.instance;
  }

  log(level: string, message: string, meta?: any): void {
    // In a real implementation, we would write to the log file
    // For now, we'll just log to console if it's an error
    if (level === 'error') {
      console.error(`[${level}] ${message}`, meta);
    }
  }
}