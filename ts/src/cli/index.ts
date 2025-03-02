#!/usr/bin/env node

import { Command } from 'commander';
import dotenv from 'dotenv';
import readline from 'readline';
import { loadConfig, getModelName, getDefaultModel } from '../config';
import { Database } from '../db/sqlite';
import { createLLMClient } from '../llm/factory';
import { Message } from '../llm/types';
import { wrapText, getTerminalWidth } from '../utils/linewrap';
import { version } from '../../package.json';

import { Logger } from '../utils/logger';
import path from 'path';
import os from 'os';

// Environment variables
dotenv.config();

const config = loadConfig();
const db = new Database(config.historyFile);

// Initialize logger
const xdgConfigPath = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
const logPath = path.join(xdgConfigPath, 'ask-ai', 'ask-ai.log');
Logger.initialize(logPath);
const logger = Logger.getInstance();

const program = new Command();

program
  .name('ask-ai')
  .description('CLI tool for asking LLMs questions without a GUI')
  .version(version)
  .argument('[query]', 'Question to ask the AI model')
  .option('-m, --model <model>', 'LLM model to use')
  .option('-c, --continue', 'Continue the last conversation')
  .option('--context <number>', 'Use last n queries for context', '0')
  .option('--search <query>', 'Search conversation history')
  .option('--show <id>', 'Show a specific conversation')
  .option('--id <id>', 'Continue a specific conversation')
  .action(async (query, options) => {
    try {
      logger.log('info', 'New query', {
        model: options.model,
        query,
        conversationId: options.id
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

      // Save user message to DB
      await db.addConversationItem(conversationId, 'user', query);

      // Get the response from the LLM
      const llmClient = createLLMClient(modelName);
      const response = await llmClient.send(messages);

      const termWidth = getTerminalWidth() - 2;
      console.log(wrapText(response.content, termWidth));

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
    } catch (error) {
      console.error('Error:', error);
    } finally {
      db.close();
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
