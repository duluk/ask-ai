#!/usr/bin/env node

/**
 * Standalone TUI implementation that doesn't use Ink
 * This avoids the top-level await issue in yoga-wasm-web
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const readline = require('readline');
const util = require('util');
const { spawn } = require('child_process');

// Import our compiled modules
const config = require('./dist/src/config');

// Terminal colors
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',
  underscore: '\x1b[4m',
  blink: '\x1b[5m',
  reverse: '\x1b[7m',
  hidden: '\x1b[8m',
  
  black: '\x1b[30m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  
  bgBlack: '\x1b[40m',
  bgRed: '\x1b[41m',
  bgGreen: '\x1b[42m',
  bgYellow: '\x1b[43m',
  bgBlue: '\x1b[44m',
  bgMagenta: '\x1b[45m',
  bgCyan: '\x1b[46m',
  bgWhite: '\x1b[47m'
};

// Clear the terminal
const clearScreen = () => {
  process.stdout.write('\x1b[2J');
  process.stdout.write('\x1b[0f');
};

// Set cursor position
const setCursor = (x, y) => {
  process.stdout.write(`\x1b[${y};${x}H`);
};

// Get terminal dimensions
const getTerminalSize = () => {
  return {
    width: process.stdout.columns || 80,
    height: process.stdout.rows || 24
  };
};

// Draw a box around content
const drawBox = (x, y, width, height, title = '') => {
  // Top border with title
  setCursor(x, y);
  process.stdout.write('┌' + title + '─'.repeat(width - 2 - title.length) + '┐');
  
  // Sides
  for (let i = 1; i < height - 1; i++) {
    setCursor(x, y + i);
    process.stdout.write('│' + ' '.repeat(width - 2) + '│');
  }
  
  // Bottom border
  setCursor(x, y + height - 1);
  process.stdout.write('└' + '─'.repeat(width - 2) + '┘');
};

// Write text within a box
const writeInBox = (x, y, width, text, color = colors.white) => {
  const lines = [];
  let currentLine = '';
  
  // Word wrap the text
  const words = text.split(' ');
  for (const word of words) {
    if (currentLine.length + word.length + 1 <= width - 4) {
      currentLine += (currentLine ? ' ' : '') + word;
    } else {
      lines.push(currentLine);
      currentLine = word;
    }
  }
  if (currentLine) {
    lines.push(currentLine);
  }
  
  // Write the lines
  for (let i = 0; i < lines.length; i++) {
    setCursor(x + 2, y + 1 + i);
    process.stdout.write(color + lines[i] + colors.reset);
  }
  
  return lines.length; // Return the number of lines written
};

// Create an interactive readline interface
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  prompt: '> '
});

// Basic chat application
class SimpleChatTUI {
  constructor() {
    this.messages = [{ role: 'system', content: 'Welcome to Ask AI Terminal UI. Type your question and press Enter.' }];
    this.input = '';
    this.isTyping = false;
    this.configData = config.loadConfig();
    this.modelName = config.getDefaultModel(this.configData);
    this.setRawMode(true);
    this.render();
    this.setupHandlers();
  }
  
  setRawMode(isRaw) {
    if (process.stdin.isTTY) {
      process.stdin.setRawMode(isRaw);
    }
  }
  
  setupHandlers() {
    // Handle input events
    process.stdin.on('data', (data) => {
      const key = data.toString();
      
      if (key === '\u0003') { // Ctrl+C
        this.exit();
      } else if (key === '\r' || key === '\n') {
        if (!this.isTyping && this.input.trim()) {
          this.handleSubmit();
        }
      } else if (key === '\u007F' || key === '\b') { // Backspace
        this.input = this.input.slice(0, -1);
        this.render();
      } else if (key >= ' ' && key <= '~') { // Printable ASCII
        this.input += key;
        this.render();
      }
    });
    
    // Handle terminal resize
    process.stdout.on('resize', () => {
      this.render();
    });
  }
  
  async handleSubmit() {
    const query = this.input.trim();
    this.input = '';
    this.messages.push({ role: 'user', content: query });
    this.isTyping = true;
    this.render();
    
    try {
      // Start a child process to get the answer
      // This avoids the module loading issues
      const child = spawn('node', [
        '-e',
        `
        const { createLLMClient } = require('./dist/src/llm/factory');
        const llmClient = createLLMClient("${this.modelName}");
        const message = { role: "user", content: ${JSON.stringify(query)} };
        
        const stream = llmClient.sendStream([message]);
        let response = "";
        
        stream.on('data', (chunk) => {
          response += chunk;
          process.stdout.write(chunk);
        });
        
        stream.on('done', () => {
          process.exit(0);
        });
        
        stream.on('error', (error) => {
          console.error(error);
          process.exit(1);
        });
        `
      ], {
        stdio: ['ignore', 'pipe', 'pipe'],
        env: process.env
      });
      
      let response = '';
      
      child.stdout.on('data', (data) => {
        response += data.toString();
        this.messages[this.messages.length - 1] = { 
          role: 'assistant', 
          content: response 
        };
        this.render();
      });
      
      child.stderr.on('data', (data) => {
        console.error(`Error: ${data}`);
      });
      
      child.on('close', (code) => {
        this.isTyping = false;
        if (code !== 0) {
          this.messages.push({ 
            role: 'system', 
            content: `Error: Command exited with code ${code}` 
          });
        }
        this.render();
      });
    } catch (error) {
      this.isTyping = false;
      this.messages.push({ role: 'system', content: `Error: ${error.message}` });
      this.render();
    }
  }
  
  render() {
    clearScreen();
    const { width, height } = getTerminalSize();
    
    // Draw main chat area
    drawBox(1, 1, width - 2, height - 4, ' Ask AI Terminal UI ');
    
    // Display chat messages
    let y = 2;
    for (const msg of this.messages) {
      let prefix = '';
      let color = colors.white;
      
      if (msg.role === 'user') {
        prefix = 'You: ';
        color = colors.green;
      } else if (msg.role === 'assistant') {
        prefix = 'AI: ';
        color = colors.cyan;
      } else {
        prefix = 'System: ';
        color = colors.yellow;
      }
      
      setCursor(3, y);
      process.stdout.write(color + prefix + colors.reset);
      
      const content = msg.content.replace(/\\n/g, '\n');
      const usedLines = writeInBox(prefix.length + 2, y, width - 6, content, color);
      y += usedLines + 2;
      
      if (y >= height - 5) break;
    }
    
    // Draw input area
    drawBox(1, height - 3, width - 2, 3, ` Model: ${this.modelName} `);
    setCursor(3, height - 2);
    process.stdout.write(this.input);
    
    // Position cursor at end of input
    setCursor(3 + this.input.length, height - 2);
  }
  
  exit() {
    clearScreen();
    this.setRawMode(false);
    process.exit(0);
  }
}

// Start the application
try {
  new SimpleChatTUI();
} catch (error) {
  console.error('Error starting TUI:', error);
  process.exit(1);
}