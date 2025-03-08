import sqlite3 from 'sqlite3';
import { Message } from '../llm/types';
import path from 'path';
import fs from 'fs';
import os from 'os';

export interface Conversation {
    id: number;
    timestamp: string;
    model: string;
}

export interface ConversationItem {
    id: number;
    conversationId: number;
    role: string;
    content: string;
    inputTokens: number;
    outputTokens: number;
    timestamp: string;
}

export class Database {
    private db: sqlite3.Database;
    private initialized: boolean = false;

    constructor(dbPath?: string) {
        const xdgConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
        const defaultPath = path.join(xdgConfigPath, 'ask-ai', 'history.db');
        const finalPath = dbPath || defaultPath;

        // Ensure directory exists
        const dir = path.dirname(finalPath);
        if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true });
        }

        this.db = new sqlite3.Database(finalPath);
    }

    async initialize(): Promise<void> {
        if (this.initialized) return;

        return new Promise((resolve, reject) => {
            this.db.serialize(() => {
                // Create tables if they don't exist
                this.db.run(`
          CREATE TABLE IF NOT EXISTS conversations (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp TEXT DEFAULT CURRENT_TIMESTAMP,
            model TEXT
          )
        `);

                this.db.run(`
          CREATE TABLE IF NOT EXISTS conversation_items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            conversation_id INTEGER,
            role TEXT,
            content TEXT,
            input_tokens INTEGER DEFAULT 0,
            output_tokens INTEGER DEFAULT 0,
            timestamp TEXT DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(conversation_id) REFERENCES conversations(id)
          )
        `, (err) => {
                    if (err) reject(err);
                    else {
                        this.initialized = true;
                        resolve();
                    }
                });
            });
        });
    }

    async createConversation(model: string): Promise<number> {
        await this.initialize();

        return new Promise((resolve, reject) => {
            this.db.run(
                'INSERT INTO conversations (model) VALUES (?)',
                [model],
                function (err) {
                    if (err) reject(err);
                    else resolve(this.lastID);
                }
            );
        });
    }

    async addConversationItem(
        conversationId: number,
        role: string,
        content: string,
        inputTokens: number = 0,
        outputTokens: number = 0
    ): Promise<void> {
        await this.initialize();

        return new Promise((resolve, reject) => {
            this.db.run(
                'INSERT INTO conversation_items (conversation_id, role, content, input_tokens, output_tokens) VALUES (?, ?, ?, ?, ?)',
                [conversationId, role, content, inputTokens, outputTokens],
                (err) => {
                    if (err) reject(err);
                    else resolve();
                }
            );
        });
    }

    async getConversation(id: number): Promise<ConversationItem[]> {
        await this.initialize();

        return new Promise((resolve, reject) => {
            this.db.all(
                `SELECT ci.id, ci.conversation_id as conversationId, ci.role, ci.content, 
                ci.input_tokens as inputTokens, ci.output_tokens as outputTokens, ci.timestamp
         FROM conversation_items ci
         WHERE ci.conversation_id = ?
         ORDER BY ci.id ASC`,
                [id],
                (err, rows) => {
                    if (err) reject(err);
                    else resolve(rows as ConversationItem[]);
                }
            );
        });
    }

    async getRecentConversations(limit: number = 10): Promise<Conversation[]> {
        await this.initialize();

        return new Promise((resolve, reject) => {
            this.db.all(
                'SELECT id, timestamp, model FROM conversations ORDER BY id DESC LIMIT ?',
                [limit],
                (err, rows) => {
                    if (err) reject(err);
                    else resolve(rows as Conversation[]);
                }
            );
        });
    }

    async searchConversations(query: string): Promise<Conversation[]> {
        await this.initialize();

        return new Promise((resolve, reject) => {
            this.db.all(
                `SELECT DISTINCT c.id, c.timestamp, c.model
         FROM conversations c
         JOIN conversation_items ci ON c.id = ci.conversation_id
         WHERE ci.content LIKE ?
         ORDER BY c.id DESC`,
                [`%${query}%`],
                (err, rows) => {
                    if (err) reject(err);
                    else resolve(rows as Conversation[]);
                }
            );
        });
    }

    async getLastConversationId(): Promise<number | null> {
        await this.initialize();

        return new Promise((resolve, reject) => {
            this.db.get(
                'SELECT id FROM conversations ORDER BY id DESC LIMIT 1',
                (err, row: { id: number } | undefined) => {
                    if (err) reject(err);
                    else resolve(row ? row.id : null);
                }
            );
        });
    }

    async getMessagesForLLM(conversationId: number, contextCount: number = 0): Promise<Message[]> {
        await this.initialize();

        return new Promise((resolve, reject) => {
            let query: string;
            let params: any[];

            if (contextCount > 0) {
                // Get the last N items from the conversation
                query = `
          SELECT role, content
          FROM conversation_items
          WHERE conversation_id = ?
          ORDER BY id DESC
          LIMIT ?
        `;
                params = [conversationId, contextCount];
            } else {
                // Get all items from the conversation
                query = `
          SELECT role, content
          FROM conversation_items
          WHERE conversation_id = ?
          ORDER BY id ASC
        `;
                params = [conversationId];
            }

            this.db.all(query, params, (err, rows: any[]) => {
                if (err) reject(err);
                else {
                    // If we used the DESC order for limiting, we need to reverse back
                    const sortedRows = contextCount > 0 ? rows.reverse() : rows;
                    const messages = sortedRows.map(row => ({
                        role: row.role as 'user' | 'assistant' | 'system',
                        content: row.content
                    }));
                    resolve(messages);
                }
            });
        });
    }

    async close(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.db.close((err) => {
                if (err) reject(err);
                else resolve();
            });
        });
    }
}
