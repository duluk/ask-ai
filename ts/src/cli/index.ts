#!/usr/bin/env node

import { Command } from 'commander';
import dotenv from 'dotenv';
import readline from 'node:readline';
import { loadConfig, getModelName, getDefaultModel } from '../config/index.js';
import { Database } from '../db/sqlite.js';
import { createLLMClient } from '../llm/factory.js';
import { ModelProvider, Message } from '../llm/types.js';
import { wrapText, LineWrapper, getTerminalWidth } from '../utils/linewrap.js';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import os from 'node:os';
import { Logger } from '../utils/logger.js';

// Read package.json dynamically instead of using import assertion
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const packageJsonPath = path.resolve(__dirname, '../../package.json');
const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf8'));

dotenv.config();

const config = loadConfig();
const db = new Database(config.historyFile);

const xdgConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
const logPath = path.join(xdgConfigPath, 'ask-ai', 'ask-ai.ts.log');
Logger.initialize(logPath);
const logger = Logger.getInstance();

async function recordResponse(
    conversationId: number,
    query: string,
    response: {
        content: string;
        usage?: {
            promptTokens?: number;
            completionTokens?: number;
        };
    },
    quiet: boolean = false
): Promise<void> {
    if (!quiet) {
        await db.addConversationItem(
            conversationId,
            'assistant',
            response.content,
            response.usage?.promptTokens || 0,
            response.usage?.completionTokens || 0
        );

        logger.log('info', 'AI response', {
            conversationId,
            query,
            response: response.content,
            tokens: response.usage
        });
    }
}

const program = new Command();

/*
In Commander.js:

- `program.opts()` gets options from the root program/command level
- `cmd.parent.opts()` gets options from the immediate parent command of the current subcommand

Here's an example to illustrate:

program
    .option('--global-flag <value>', 'Global flag')  // Root level option

program
    .command('parent')
    .option('--parent-flag <value>', 'Parent flag')  // Parent command option
    .command('child')                                // Subcommand
    .option('--child-flag <value>', 'Child flag')    // Child command option
    .action((childOptions, cmd) => {
        console.log({
            programOpts: program.opts(),        // Gets { globalFlag: value }
            parentOpts: cmd.parent.opts(),      // Gets { parentFlag: value }
            childOpts: childOptions             // Gets { childFlag: value }
        });
    });
*/

// Add tui command
program
    .command('tui')
    .description('Launch the Terminal User Interface for interactive chat')
    .option('-m, --model <model>', 'LLM model to use')
    .option('--debug', 'Enable debug mode')
    // 1st Argument to .action: options pased to the current command (ie tui)
    // 2nd Argument: full command object, which includes the above options plus other things such as the app name and probably options passed to the app, before the sub command
    .action(async (cmdOptions, _) => {
        try {
            const { startTUI } = await import('../tui/index.js');
            const model = cmdOptions.model || program.opts().model;
            const debug = cmdOptions.debug || program.opts().debug;
            logger.log('debug', 'TUI options in main app:', { model, debug })
            await startTUI({ model, debug });
        } catch (error) {
            console.error('Failed to start TUI:', error);
        }
    });

// Main command
program
    .name('ask-ai')
    .description('CLI tool for asking LLMs questions without a GUI')
    .version(packageJson.version)
    .argument('[query]', 'Question to ask the AI model')
    .option('-m, --model <model>', 'LLM model to use')
    .option('-c, --continue', 'Continue the last conversation')
    .option('--context <number>', 'Use last n queries for context', '0')
    .option('--search <query>', 'Search conversation history')
    .option('--show <id>', 'Show a specific conversation')
    .option('--id <id>', 'Continue a specific conversation')
    .option('--stream', 'Stream the response')
    .option('--debug', 'Enable debug mode')
    .option('-q, --quiet', 'Suppress output to console')
    .action(async (query, options) => {
        logger.log('debug', 'ask-ai CLI Query:', { query })
        logger.log('debug', 'ask-ai CLI Options:', { options })
        const pendingOperations: Promise<void>[] = [];
        try {
            logger.log('debug', 'Options:', { options })
            logger.log('info', 'New query', {
                model: options.model,
                query,
                conversationId: options.id,
                stream: options.stream,
            });

            if (options.search) {
                await handleSearch(options.search);
                return;
            }

            if (options.show) {
                await handleShow(parseInt(options.show, 10));
                return;
            }

            const modelName = options.model
                ? getModelName(options.model, config)
                : getDefaultModel(config);

            let conversationId: number | null = null;

            if (options.id) {
                conversationId = parseInt(options.id, 10);
            } else if (options.continue) {
                conversationId = await db.getLastConversationId();
            }

            if (!query) {
                query = await promptForQuery(modelName);
                if (!query) {
                    console.log('No query provided. Exiting.');
                    return;
                }
            }

            if (!conversationId) {
                conversationId = await db.createConversation(modelName);
            }

            const contextCount = parseInt(options.context, 10) || 0;
            const messages: Message[] = await db.getMessagesForLLM(conversationId, contextCount);

            messages.push({
                role: 'user',
                content: query
            });

            await db.addConversationItem(conversationId, 'user', query);

            // Query LLM
            const llmClient = createLLMClient(modelName);
            const termWidth = Math.min(getTerminalWidth() - 2, 80);

            if (options.stream) {
                const stream = llmClient.sendStream(messages);
                const wrapper = new LineWrapper(termWidth);
                let fullContent = '';

                // Create a promise that resolves when streaming is complete
                const streamComplete = new Promise<void>((resolve, reject) => {
                    stream.on('data', (chunk) => {
                        if (!options.quiet) {
                            const wrapped = wrapper.wrap(chunk);
                            process.stdout.write(wrapped);
                        }
                        fullContent += chunk;
                    });

                    stream.on('done', async (response) => {
                        if (!options.quiet) console.log('');
                        await recordResponse(conversationId, query, {
                            content: fullContent,
                            usage: response.usage
                        }, options.quiet);
                        resolve();
                    });

                    stream.on('error', (error) => {
                        reject(error);
                    });
                });

                // Store the promise for later awaiting
                pendingOperations.push(streamComplete);
            } else {
                const response = await llmClient.send(messages);
                if (!options.quiet) {
                    console.log(wrapText(response.content, termWidth));
                }

                // Store the promise for later awaiting
                pendingOperations.push(
                    recordResponse(conversationId, query, response, options.quiet)
                );
            }
        } catch (error) {
            console.error('Error:', error);
        } finally {
            // Ensure all pending operations are complete before closing
            await Promise.all(pendingOperations);
            await db.close();
        }
    });

async function promptForQuery(modelName: string): Promise<string> {
    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout
    });

    return new Promise((resolve) => {
        rl.question(`${modelName}> `, (answer) => {
            rl.close();
            resolve(answer);
        });
    });
}

async function handleSearch(searchQuery: string): Promise<void> {
    const conversations = await db.searchConversations(searchQuery);

    if (conversations.length === 0) {
        console.log('No matching conversations found.');
        return;
    }

    console.log(`Found ${conversations.length} matching conversations:`);

    for (const conv of conversations) {
        console.log(`ID: ${conv.id} | ${conv.timestamp} | Model: ${conv.model}`);

        // Show the first matching message for context
        const items = await db.getConversation(conv.id);
        const matchingItem = items.find(item =>
            item.content.toLowerCase().includes(searchQuery.toLowerCase())
        );

        if (matchingItem) {
            const preview = matchingItem.content.substring(0, 60) + (matchingItem.content.length > 60 ? '...' : '');
            console.log(`    "${preview}"`);
        }

        console.log('');
    }
}

async function handleShow(conversationId: number): Promise<void> {
    const items = await db.getConversation(conversationId);

    if (items.length === 0) {
        console.log(`Conversation with ID ${conversationId} not found.`);
        return;
    }

    const termWidth = getTerminalWidth() - 2;

    for (const item of items) {
        const prefix = item.role === 'user' ? 'You: ' : 'AI: ';
        console.log(`${prefix}${wrapText(item.content, termWidth)}\n`);
    }
}

program.parse(process.argv);
