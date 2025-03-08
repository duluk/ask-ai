import { Database } from '../../src/db/sqlite.js';
import sqlite3 from 'sqlite3';

jest.mock('sqlite3');

describe('Database', () => {
    let db: Database;

    beforeEach(() => {
        jest.clearAllMocks();
        db = new Database(':memory:');
    });

    it('should initialize database correctly', () => {
        expect(sqlite3.Database).toHaveBeenCalledWith(':memory:');
    });

    it('should record conversation correctly', async () => {
        const mockRun = jest.fn((query: string, params: any[], callback: Function) => {
            callback(null);
        });
        (db as any).db = { run: mockRun };

        const conversationId = await db.createConversation('test-model');

        expect(conversationId).toBe(1);
        expect(mockRun).toHaveBeenCalledWith(
            'INSERT INTO conversations (model) VALUES (?)',
            ['test-model'],
            expect.any(Function)
        );

        expect(mockRun).toHaveBeenCalled();
    });

    it('should add conversation item correctly', async () => {
        const mockRun = jest.fn((query: string, params: any[], callback: Function) => {
            callback(null);
        });
        (db as any).db = { run: mockRun };

        await db.addConversationItem(
            1, // conversationId
            'user', // role
            'test message', // content
            10, // promptTokens
            20 // completionTokens
        );

        expect(mockRun).toHaveBeenCalledWith(
            expect.stringContaining('INSERT INTO conversation_items'),
            [1, 'user', 'test message', 10, 20],
            expect.any(Function)
        );
    });

    it('should get messages for LLM correctly', async () => {
        const mockAll = jest.fn((query: string, params: any[], callback: Function) => {
            callback(null, [
                { role: 'user', content: 'test question' },
                { role: 'assistant', content: 'test answer' }
            ]);
        });
        (db as any).db = { all: mockAll };

        const messages = await db.getMessagesForLLM(1, 2);

        expect(messages).toHaveLength(2);
        expect(mockAll).toHaveBeenCalledWith(
            expect.stringContaining('SELECT role, content FROM conversation_items'),
            [1, 2],
            expect.any(Function)
        );
    });
});
