#!/usr/bin/env node

import blessed from 'blessed';
import { loadConfig, getModelName, getDefaultModel } from '../config';
import { Database } from '../db/sqlite';
import { createLLMClient } from '../llm/factory';
import { Message } from '../llm/types';
import { LineWrapper } from '../utils/linewrap';
import { Logger } from '../utils/logger';
import path from 'path';
import os from 'os';

interface TUIOptions {
    model?: string;
    debug?: boolean;
}

export async function startTUI(options: TUIOptions = {}) {
    // Initialize config, database and logger
    const config = loadConfig();
    const db = new Database(config.historyFile);

    const xdgConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
    const logPath = path.join(xdgConfigPath, 'ask-ai', 'ask-ai.ts.log');
    Logger.initialize(logPath);
    const logger = Logger.getInstance();

    // Set the model name from options or default
    let modelName: string = options.model
        ? getModelName(options.model, config)
        : getDefaultModel(config);

    // Create a screen object
    const screen = blessed.screen({
        smartCSR: true,
        title: 'Ask AI',
        dockBorders: true,
        fullUnicode: true,
    });

    // Create a box for displaying the AI responses
    const outputBox = blessed.box({
        top: 0,
        left: 0,
        width: '100%',
        height: '100%-4', // Leave room for status and min input height
        scrollable: true,
        alwaysScroll: true,
        scrollbar: {
            ch: ' ',
            track: {
                bg: 'gray'
            },
            style: {
                inverse: true
            }
        },
        tags: true,
        keys: true,
        vi: true,
        mouse: true,
        border: {
            type: 'line'
        },
        style: {
            border: {
                fg: 'blue'
            }
        }
    });

    // Create a text input for the user's questions
    const inputBox = blessed.textarea({
        bottom: 0,
        left: 0,
        height: 3,
        width: '100%',
        keys: true,
        inputOnFocus: true,
        vi: true, // Enable VI mode for better key handling
        border: {
            type: 'line'
        },
        style: {
            border: {
                fg: 'blue'
            },
            focus: {
                border: {
                    fg: 'green'
                }
            }
        }
    });

    // Add escape handler for inputBox to exit immediately (currently requires hitting Esc twice)
    // inputBox.key('escape', () => {
    //     process.exit(0);
    // });

    // Register explicit key binding for inserting newlines with Ctrl+N
    inputBox.key('C-n', () => {
        const currentValue = inputBox.getValue();
        const cursorPos = (inputBox as any)._caret?.position || 0;

        inputBox.setValue(currentValue);

        logger.log('debug', 'C-n event:', {
            cursorPos: cursorPos,
            oldValue: currentValue,
            newValue: inputBox.getValue(),
        });

        const lines = (currentValue + '\n').split('\n').length;
        const newHeight = Math.min(Math.max(3, lines + 2), 10); // Min 3, Max 10 lines

        if (newHeight !== inputBox.height) {
            inputBox.height = newHeight;
            statusLine.bottom = newHeight;
            // +1 for the statusLine
            outputBox.height = `100%-${newHeight + 1}`;
        }

        inputBox.setValue(inputBox.getValue() + '\n');

        // Set new cursor position after the newline
        if ((inputBox as any)._caret) {
            (inputBox as any)._caret.position = cursorPos + 1;
        }

        inputBox.focus();
        screen.render();
    });

    // Status line positioned relative to input
    const statusLine = blessed.text({
        bottom: 3,
        left: 0,
        width: '100%',
        height: 1,
        content: 'Use Ctrl+N for newline | Press Esc to quit',
        style: {
            fg: 'white',
            bg: 'blue'
        }
    });

    // Add all elements to the screen
    screen.append(outputBox);
    screen.append(inputBox);
    screen.append(statusLine);

    // Set focus on the input box
    inputBox.focus();

    // Handle key events
    screen.key(['escape', 'C-c'], async () => {
        await db.close();
        return process.exit(0);
    });

    // Current conversation ID
    let conversationId: number | null = null;
    //    let modelName: string = getDefaultModel(config);

    // Add a message to the output box
    function appendMessage(role: string, content: string): void {
        if (role === 'user') {
            outputBox.pushLine('You:');
            outputBox.pushLine(content);
            outputBox.pushLine('');
        } else {
            outputBox.pushLine('AI:');
            outputBox.pushLine(content);
            outputBox.pushLine('');
        }
        outputBox.setScrollPerc(100);
        screen.render();
    }

    // Initialize the conversation
    async function initConversation(): Promise<void> {
        try {
            // Create a new conversation
            conversationId = await db.createConversation(modelName);

            // Update status line
            statusLine.setContent(`Model: ${modelName} | Use Ctrl+N for newline | Press Esc to quit`);
            screen.render();
        } catch (error) {
            outputBox.pushLine(`Error initializing conversation: ${error}`);
            screen.render();
        }
    }

    // Send a message to the AI and display the response
    async function sendMessage(query: string): Promise<void> {
        try {
            if (!conversationId) {
                await initConversation();
            }

            if (!query.trim()) {
                return;
            }

            // Display user message
            appendMessage('user', query);

            // Add to database
            await db.addConversationItem(conversationId!, 'user', query);

            // Prepare the context
            const messages: Message[] = await db.getMessagesForLLM(conversationId!, 10); // Get last 10 messages for context

            // Query LLM
            const llmClient = createLLMClient(modelName);

            // Show "thinking" indicator
            statusLine.setContent('AI is thinking...');
            screen.render();

            try {
                // Use streaming for more interactive feel
                const stream = llmClient.sendStream(messages);

                // Start with "AI:" header
                outputBox.pushLine('AI:');

                // We'll keep track of the complete response
                let responseText = '';

                // Track the lines we've added so far
                let responseLines = 0;

                // Buffer for collecting text
                let buffer = '';

                // Track the width of the output box for better wrapping
                const boxWidth = typeof outputBox.width === 'number'
                    ? outputBox.width - 2
                    : parseInt(outputBox.width as string, 10) - 2;
                // If using  perecentage based width, convert to number
                // const boxWidth = screen.width - 4;

                // Split the response into paragraphs and lines for better display
                stream.on('data', (chunk) => {
                    // Add to the full response
                    responseText += chunk;

                    // Process the chunk
                    if (chunk.includes('\n')) {
                        // If the chunk contains newlines, add each part as a separate line
                        const newLines = chunk.split('\n');

                        // Handle the first part (append to current buffer)
                        buffer += newLines[0];

                        // If this is the first time we're adding text
                        if (responseLines === 0) {
                            // Add a new line with the buffer content
                            outputBox.pushLine(buffer);
                            responseLines++;
                        } else {
                            // Update the last line with the new buffer content
                            const lastLineIndex = outputBox.getLines().length - 1;
                            const currentText = outputBox.getLine(lastLineIndex);
                            outputBox.setLine(lastLineIndex, currentText + buffer);
                        }

                        // Add each remaining part as a new line
                        for (let i = 1; i < newLines.length; i++) {
                            outputBox.pushLine(newLines[i]);
                            responseLines++;
                        }

                        // Reset buffer
                        buffer = '';
                    } else {
                        // Add to buffer and display when appropriate
                        buffer += chunk;

                        // Display buffer if it contains sentence endings or gets long enough
                        if (buffer.includes('.') || buffer.includes('!') || buffer.includes('?') || buffer.length >= 30) {
                            // If this is the first time we're adding text
                            if (responseLines === 0) {
                                // Add a new line with the buffer content
                                outputBox.pushLine(buffer);
                                responseLines++;
                            } else {
                                // Update the last line with the new buffer content
                                const lastLineIndex = outputBox.getLines().length - 1;
                                const currentText = outputBox.getLine(lastLineIndex);

                                // If adding this buffer would make the line too long, start a new line
                                if ((currentText.length + buffer.length) > boxWidth) {
                                    outputBox.pushLine(buffer);
                                    responseLines++;
                                } else {
                                    // Append to the current line
                                    outputBox.setLine(lastLineIndex, currentText + buffer);
                                }
                            }

                            // Reset buffer
                            buffer = '';
                        }
                    }

                    // Ensure we're scrolled to the bottom
                    outputBox.setScrollPerc(100);
                    screen.render();
                });

                stream.on('done', async (response) => {
                    // If there's any remaining buffer content, add it
                    if (buffer.length > 0) {
                        // Add as a new line
                        outputBox.pushLine(buffer);
                        buffer = '';
                    }

                    // Record the response to the database
                    await db.addConversationItem(
                        conversationId!,
                        'assistant',
                        responseText,
                        response.usage?.promptTokens || 0,
                        response.usage?.completionTokens || 0
                    );

                    // Add an empty line after the AI response for separation
                    outputBox.pushLine('');

                    // Reset status line
                    statusLine.setContent(`Model: ${modelName} | Use Ctrl+N for newline | Press Esc to quit`);
                    screen.render();
                });

                stream.on('error', (error) => {
                    outputBox.pushLine(`Error from AI: ${error}`);
                    statusLine.setContent(`Model: ${modelName} | Use Ctrl+N for newline | Press Esc to quit`);
                    screen.render();
                });
            } catch (error) {
                outputBox.pushLine(`Error communicating with AI: ${error}`);
                statusLine.setContent(`Model: ${modelName} | Use Ctrl+N for newline | Press Esc to quit`);
                screen.render();
            }
        } catch (error) {
            outputBox.pushLine(`Error: ${error}`);
            screen.render();
        }
    }

    // Handle input box content changes
    inputBox.on('keypress', async (ch, key: any) => {
        // Handle Enter key
        if (key.name === 'enter' || key.name === 'return') {
            // Add logging to debug key event
            logger.log('debug', 'Key event:', {
                key: JSON.stringify(key),
                char: ch ? ch.charCodeAt(0) : 'null',
                shift: key.shift,
                meta: key.meta,
                ctrl: key.ctrl
            });

            // Regular Enter submits
            const query = inputBox.getValue().trim();
            if (!query) return;

            inputBox.setValue('');
            inputBox.height = 3;
            statusLine.bottom = 3;
            outputBox.height = '100%-4';
            screen.render();
            await sendMessage(query);
            inputBox.focus();
            return;
        }

        // Wait for next tick to get updated content
        process.nextTick(() => {
            const content = inputBox.getValue();
            const lines = content.split('\n').length;
            const newHeight = Math.min(Math.max(3, lines + 2), 10); // Min 3, Max 10 lines

            if (newHeight !== inputBox.height) {
                inputBox.height = newHeight;
                statusLine.bottom = newHeight;
                // +1 for the statusLine
                outputBox.height = `100%-${newHeight + 1}`;
                screen.render();
            }
        });
    });

    // Initialize
    (async () => {
        try {
            await initConversation();

            // Display welcome message
            outputBox.pushLine('Welcome to Ask AI Terminal UI');
            outputBox.pushLine('Type your question below and press Enter');
            outputBox.pushLine('Press Ctrl+N to insert a new line');
            outputBox.pushLine('');

            screen.render();
        } catch (error) {
            console.error('Failed to initialize:', error);
            process.exit(1);
        }
    })();

    // Render the screen
    screen.render();
}

// Move your initialization code inside startTUI
