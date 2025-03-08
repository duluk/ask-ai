#!/usr/bin/env node

import blessed from 'blessed';
import { loadConfig, getModelName, getDefaultModel } from '../config';
import { Database } from '../db/sqlite';
import { createLLMClient } from '../llm/factory';
import { Message } from '../llm/types';
// import { LineWrapper } from '../utils/linewrap';
import { Logger } from '../utils/logger';

// import path from 'path';
// import os from 'os';

interface TUIOptions {
    model?: string;
    debug?: boolean;
}

export async function startTUI(options: TUIOptions = {}) {
    // Initialize config, database and logger
    const config = loadConfig();
    const db = new Database(config.historyFile);

    // The logger is initialzed in cli/index.ts so commenting out for now
    // const xdgConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
    // const logPath = path.join(xdgConfigPath, 'ask-ai', 'ask-ai.ts.log');
    // Logger.initialize(logPath);
    const logger = Logger.getInstance();
    logger.log('debug', 'TUI options in startTUI:', options);
    logger.log('debug', 'Config:', config);

    const statusHelp = 'Ctrl+A/E:Start/End | Ctrl+B/F:Back/Forward | Alt+B/F:Prev/Next word | Ctrl+P/J:Prev/Next line | Ctrl+K:Kill | Ctrl+W:Del word | Ctrl+N:Newline | Ctrl+D/Esc:Quit';

    // Set the model name from options or default
    let modelName: string = options.model
        ? getModelName(options.model, config)
        : getDefaultModel(config);

    // Create a screen object
    const screen = blessed.screen({
        smartCSR: true,
        dockBorders: true,
        fullUnicode: true,
        title: 'Ask AI',
    });

    // Create a box for displaying the AI responses
    const outputBox = blessed.box({
        parent: screen,
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
        parent: screen,
        bottom: 0,
        left: 0,
        height: 3,
        width: '100%',
        tags: true,
        keys: true,
        inputOnFocus: true,
        input: true,
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
    inputBox.key('C-d', async () => {
        await db.close();
        process.exit(0);
    });

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

    // Emacs-style keybindings

    // Navigation keybindings
    inputBox.key('C-a', () => { // Move to beginning of line
        if ((inputBox as any)._caret) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;
            const lines = text.split('\n');

            // Find current line and position within that line
            let charCount = 0;
            let lineStart = 0;

            for (let i = 0; i < lines.length; i++) {
                const lineLength = lines[i].length;
                if (charCount + lineLength >= curPos) {
                    lineStart = charCount;
                    break;
                }
                // +1 for the newline character
                charCount += lineLength + 1;
            }

            (inputBox as any)._caret.position = lineStart;
            screen.render();
        }
    });

    inputBox.key('C-e', () => { // Move to end of line
        if ((inputBox as any)._caret) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;
            const lines = text.split('\n');

            // Find current line and position within that line
            let charCount = 0;

            for (let i = 0; i < lines.length; i++) {
                const lineLength = lines[i].length;
                if (charCount + lineLength >= curPos) {
                    (inputBox as any)._caret.position = charCount + lineLength;
                    break;
                }
                // +1 for the newline character
                charCount += lineLength + 1;
            }

            screen.render();
        }
    });

    inputBox.key('C-b', () => { // Move back one character
        if ((inputBox as any)._caret && (inputBox as any)._caret.position > 0) {
            (inputBox as any)._caret.position -= 1;
            screen.render();
        }
    });

    inputBox.key('C-f', () => { // Move forward one character
        if ((inputBox as any)._caret) {
            const textLength = inputBox.getValue().length;
            if ((inputBox as any)._caret.position < textLength) {
                (inputBox as any)._caret.position += 1;
                screen.render();
            }
        }
    });

    inputBox.key('C-p', () => { // Move to previous line
        if ((inputBox as any)._caret) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;
            const lines = text.split('\n');

            let charCount = 0;
            let currentLine = 0;
            let positionInLine = 0;

            // Find the current line and position within that line
            for (let i = 0; i < lines.length; i++) {
                const lineLength = lines[i].length;
                if (charCount + lineLength >= curPos) {
                    currentLine = i;
                    positionInLine = curPos - charCount;
                    break;
                }
                // +1 for the newline character
                charCount += lineLength + 1;
            }

            // Move to previous line if possible
            if (currentLine > 0) {
                const prevLineLength = lines[currentLine - 1].length;
                const newPosInLine = Math.min(positionInLine, prevLineLength);

                // Calculate new absolute position
                let newPosition = 0;
                for (let i = 0; i < currentLine - 1; i++) {
                    newPosition += lines[i].length + 1;
                }
                newPosition += newPosInLine;

                (inputBox as any)._caret.position = newPosition;
                screen.render();
            }
        }
    });

    inputBox.key('C-j', () => { // Move to next line (Emacs-style)
        if ((inputBox as any)._caret) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;
            const lines = text.split('\n');

            let charCount = 0;
            let currentLine = 0;
            let positionInLine = 0;

            // Find the current line and position within that line
            for (let i = 0; i < lines.length; i++) {
                const lineLength = lines[i].length;
                if (charCount + lineLength >= curPos) {
                    currentLine = i;
                    positionInLine = curPos - charCount;
                    break;
                }
                // +1 for the newline character
                charCount += lineLength + 1;
            }

            // Move to next line if possible
            if (currentLine < lines.length - 1) {
                const nextLineLength = lines[currentLine + 1].length;
                const newPosInLine = Math.min(positionInLine, nextLineLength);

                // Calculate new absolute position
                let newPosition = 0;
                for (let i = 0; i <= currentLine; i++) {
                    newPosition += lines[i].length + 1;
                }
                newPosition += newPosInLine;

                (inputBox as any)._caret.position = newPosition;
                screen.render();
            }
        }
    });

    // Editing keybindings
    inputBox.key('C-k', () => { // Kill line from cursor to end
        if ((inputBox as any)._caret) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;
            const lines = text.split('\n');

            // Find current line and position within that line
            let charCount = 0;
            let currentLine = 0;
            let positionInLine = 0;

            for (let i = 0; i < lines.length; i++) {
                const lineLength = lines[i].length;
                if (charCount + lineLength >= curPos) {
                    currentLine = i;
                    positionInLine = curPos - charCount;
                    break;
                }
                // +1 for the newline character
                charCount += lineLength + 1;
            }

            // Remove text from cursor to end of line
            const beforeCursor = text.substring(0, curPos);
            const lineEnd = charCount + lines[currentLine].length;
            let afterCursor = text.substring(lineEnd);

            if (positionInLine < lines[currentLine].length) {
                // We're in the middle of a line, remove to end of this line
                afterCursor = text.substring(lineEnd);
            } else {
                // We're at the end of a line, remove the newline
                afterCursor = text.substring(lineEnd + 1);
            }

            inputBox.setValue(beforeCursor + afterCursor);
            screen.render();
        }
    });

    inputBox.key('C-w', () => { // Delete word backward
        if ((inputBox as any)._caret && (inputBox as any)._caret.position > 0) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;

            // Find the start of the previous word
            let wordStart = curPos - 1;

            // Skip any whitespace directly before cursor
            while (wordStart > 0 && /\s/.test(text[wordStart])) {
                wordStart--;
            }

            // Find the start of the word
            while (wordStart > 0 && !/\s/.test(text[wordStart - 1])) {
                wordStart--;
            }

            // Delete from word start to cursor
            const newText = text.substring(0, wordStart) + text.substring(curPos);
            inputBox.setValue(newText);
            (inputBox as any)._caret.position = wordStart;

            screen.render();
        }
    });

    inputBox.key('C-y', () => { // Not a true Yank, but demonstrates the concept
        // In a full implementation, this would paste from a kill ring
        logger.log('debug', 'C-y pressed: Yank functionality would go here');
    });

    inputBox.key('C-h', () => { // Delete character backward (backspace)
        if ((inputBox as any)._caret && (inputBox as any)._caret.position > 0) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;

            const newText = text.substring(0, curPos - 1) + text.substring(curPos);
            inputBox.setValue(newText);
            (inputBox as any)._caret.position = curPos - 1;

            screen.render();
        }
    });

    inputBox.key('C-d', (ch, key) => { // Delete character forward (unless at EOF, then exit)
        if ((inputBox as any)._caret) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;

            // If we're at the end of the file, and there's no text, exit
            if (curPos >= text.length && text.length === 0) {
                db.close().then(() => process.exit(0));
                return;
            }

            // Otherwise delete the character at cursor
            if (curPos < text.length) {
                const newText = text.substring(0, curPos) + text.substring(curPos + 1);
                inputBox.setValue(newText);
                screen.render();
            }
        }
    });

    inputBox.key('M-b', () => { // Move backward one word
        if ((inputBox as any)._caret && (inputBox as any)._caret.position > 0) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;

            // Find the start of the current/previous word
            let newPos = curPos - 1;

            // Skip any whitespace directly before cursor
            while (newPos > 0 && /\s/.test(text[newPos])) {
                newPos--;
            }

            // Find the start of the word
            while (newPos > 0 && !/\s/.test(text[newPos - 1])) {
                newPos--;
            }

            (inputBox as any)._caret.position = newPos;
            screen.render();
        }
    });

    inputBox.key('M-f', () => { // Move forward one word
        if ((inputBox as any)._caret) {
            const text = inputBox.getValue();
            const curPos = (inputBox as any)._caret.position;

            if (curPos >= text.length) return;

            // Find the end of the current word
            let newPos = curPos;

            // Skip any whitespace directly at or after cursor
            while (newPos < text.length && /\s/.test(text[newPos])) {
                newPos++;
            }

            // Move to the end of the word
            while (newPos < text.length && !/\s/.test(text[newPos])) {
                newPos++;
            }

            (inputBox as any)._caret.position = newPos;
            screen.render();
        }
    });

    // Status line positioned relative to input
    const statusLine = blessed.text({
        parent: screen,
        bottom: 3,
        left: 0,
        width: '100%',
        height: 1,
        content: statusHelp,
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

    // Handle key events; I think you have to hit it twice because it's
    // attached to the screen and not inputBox
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

    // Note to self: declared as async and returns Promise<void>. Thus the
    // await handles the Promise from db.createConversation, and the default
    // return value is Promise<void> even without an explicit return statement.
    // Promise<void> is essentially void, or returning nothing.
    async function initConversation(): Promise<void> {
        try {
            // Create a new conversation
            conversationId = await db.createConversation(modelName);

            // Update status line
            statusLine.setContent(`Model: ${modelName} | ${statusHelp}`);
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
            statusLine.setContent('{cyan-fg}AI is thinking...{/cyan-fg}');
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
                    statusLine.setContent(`Model: ${modelName} | ${statusHelp}`);
                    screen.render();
                });

                stream.on('error', (error) => {
                    outputBox.pushLine(`Error from AI: ${error}`);
                    statusLine.setContent(`Model: ${modelName} | ${statusHelp}`);
                    screen.render();
                });
            } catch (error) {
                outputBox.pushLine(`Error communicating with AI: ${error}`);
                statusLine.setContent(`Model: ${modelName} | ${statusHelp}`);
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
            const currentValue = inputBox.getValue();

            // Handle /commands
            if (currentValue.startsWith('/')) {
                const command = currentValue.trim();
                if (command === '/exit' || command === '/quit') {
                    await db.close();
                    process.exit(0);
                }
                if (command.startsWith('/model')) {
                    // Implement model switching logic here
                    const newModel = command.split(' ')[1];
                    logger.log('debug', 'Switching to model from command:', { newModel, command });
                    modelName = getModelName(newModel, config)
                    if (!modelName) {
                        outputBox.pushLine('Error: No model provided or invalid model. Using default.');
                        outputBox.pushLine('');
                        modelName = getDefaultModel(config);
                    }
                    statusLine.setContent(`Model: ${modelName} | ${statusHelp}`);
                    inputBox.setValue('');
                    screen.render();
                    inputBox.focus();
                    return;
                }
            }

            // Add logging to debug key event
            logger.log('debug', 'Key event:', {
                key: JSON.stringify(key),
                char: ch?.charCodeAt(0) ?? 'null',
                shift: key.shift,
                meta: key.meta,
                ctrl: key.ctrl
            });

            // Regular Enter submits
            const query = currentValue.trim();
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
            outputBox.pushLine('{blue}Welcome to Ask AI Terminal UI{/blue}');
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
