#!/usr/bin/env node

import { Command } from 'commander';
import dotenv from 'dotenv';
import readline from 'readline';
import { loadConfig, getModelName, getDefaultModel } from '../config';
import { Database } from '../db/sqlite';
import { createLLMClient } from '../llm/factory';
import { Message } from '../llm/types';
import { wrapText, getTerminalWidth } from '../utils/linewrap';

// Load environment variables
dotenv.config();

// Load config
const config = loadConfig();
const db = new Database(config.historyFile);

const program = new Command();

program
  .name('ask-ai')
  .description('CLI tool for asking LLMs questions without a GUI')
  .version('0.1.0')
  .argument('[query]', 'Question to ask the AI model')
  .option('-m, --model <model>', 'LLM model to use')
  .option('-c, --continue', 'Continue the last conversation')
  .option('--context <number>', 'Use last n queries for context', '0')
  .option('--search <query>', 'Search conversation history')
  .option('--show <id>', 'Show a specific conversation')
  .option('--id <id>', 'Continue a specific conversation')
  .action(async (query, options) => {
    try {
      // Handle search
      if (options.search) {
        await handleSearch(options.search);
        return;
      }
      
      // Handle show conversation
      if (options.show) {
        await handleShow(parseInt(options.show, 10));
        return;
      }
      
      // Get model to use
      const modelName = options.model 
        ? getModelName(options.model, config) 
        : getDefaultModel(config);
      
      // Get conversation ID
      let conversationId: number | null = null;
      
      if (options.id) {
        // Use specific conversation
        conversationId = parseInt(options.id, 10);
      } else if (options.continue) {
        // Continue last conversation
        conversationId = await db.getLastConversationId();
      }
      
      // If no query provided, prompt the user
      if (!query) {
        query = await promptForQuery(modelName);
        if (!query) {
          console.log('No query provided. Exiting.');
          return;
        }
      }
      
      // Create a new conversation if needed
      if (!conversationId) {
        conversationId = await db.createConversation(modelName);
      }
      
      // Get context if needed
      const contextCount = parseInt(options.context, 10) || 0;
      const messages: Message[] = await db.getMessagesForLLM(conversationId, contextCount);
      
      // Add current query to messages
      messages.push({
        role: 'user',
        content: query
      });
      
      // Save user message to DB
      await db.addConversationItem(conversationId, 'user', query);
      
      // Send to LLM
      const llmClient = createLLMClient(modelName);
      const response = await llmClient.send(messages);
      
      // Print response
      const termWidth = getTerminalWidth() - 2;
      console.log(wrapText(response.content, termWidth));
      
      // Save response to DB
      await db.addConversationItem(
        conversationId,
        'assistant',
        response.content,
        response.usage?.promptTokens || 0,
        response.usage?.completionTokens || 0
      );
      
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