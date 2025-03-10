import fs from 'fs';
import path from 'path';
import os from 'os';
import sqlite3 from 'sqlite3';
import { Message } from './llm-wrapper';

// A simplified version of the Database class in db/sqlite.ts
export class Database {
  private db: sqlite3.Database;
  private isOpen: boolean = false;

  constructor(dbPath: string) {
    // Ensure the directory exists
    const dbDir = path.dirname(dbPath);
    if (!fs.existsSync(dbDir)) {
      fs.mkdirSync(dbDir, { recursive: true });
    }

    this.db = new sqlite3.Database(dbPath);
    this.isOpen = true;
    this.initialize();
  }

  private initialize(): void {
    this.db.serialize(() => {
      // Create conversations table
      this.db.run(`
        CREATE TABLE IF NOT EXISTS conversations (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          timestamp TEXT DEFAULT CURRENT_TIMESTAMP,
          model TEXT
        )
      `);

      // Create conversation_items table
      this.db.run(`
        CREATE TABLE IF NOT EXISTS conversation_items (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          conversation_id INTEGER,
          role TEXT,
          content TEXT,
          prompt_tokens INTEGER DEFAULT 0,
          completion_tokens INTEGER DEFAULT 0,
          timestamp TEXT DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (conversation_id) REFERENCES conversations (id)
        )
      `);
    });
  }

  async createConversation(model: string): Promise<number> {
    return new Promise((resolve, reject) => {
      this.db.run(
        'INSERT INTO conversations (model) VALUES (?)',
        [model],
        function (err) {
          if (err) {
            reject(err);
          } else {
            resolve(this.lastID);
          }
        }
      );
    });
  }

  async addConversationItem(
    conversationId: number,
    role: string,
    content: string,
    promptTokens: number = 0,
    completionTokens: number = 0
  ): Promise<void> {
    return new Promise((resolve, reject) => {
      this.db.run(
        'INSERT INTO conversation_items (conversation_id, role, content, prompt_tokens, completion_tokens) VALUES (?, ?, ?, ?, ?)',
        [conversationId, role, content, promptTokens, completionTokens],
        (err) => {
          if (err) {
            reject(err);
          } else {
            resolve();
          }
        }
      );
    });
  }
  
  async getLastConversationId(): Promise<number | null> {
    return new Promise((resolve, reject) => {
      this.db.get(
        'SELECT id FROM conversations ORDER BY id DESC LIMIT 1',
        (err, row: { id: number } | undefined) => {
          if (err) {
            reject(err);
          } else {
            resolve(row ? row.id : null);
          }
        }
      );
    });
  }

  async getMessagesForLLM(conversationId: number, contextCount: number = 0): Promise<Message[]> {
    return new Promise((resolve, reject) => {
      // If contextCount is 0, get all messages
      const limit = contextCount > 0 ? contextCount : 1000000;
      
      this.db.all(
        `SELECT role, content FROM conversation_items 
         WHERE conversation_id = ? 
         ORDER BY id DESC 
         LIMIT ?`,
        [conversationId, limit],
        (err, rows: any[]) => {
          if (err) {
            reject(err);
          } else {
            // Convert to the required format and reverse to get chronological order
            const messages: Message[] = rows.reverse().map(row => ({
              role: row.role as 'system' | 'user' | 'assistant',
              content: row.content
            }));
            resolve(messages);
          }
        }
      );
    });
  }

  async close(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.isOpen) {
        resolve();
        return;
      }

      this.db.close((err) => {
        if (err) {
          reject(err);
        } else {
          this.isOpen = false;
          resolve();
        }
      });
    });
  }
}