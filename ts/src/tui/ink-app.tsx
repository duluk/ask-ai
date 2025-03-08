import React, { useState, useEffect, useRef } from 'react';
import { render, Box, Text, useInput, useApp } from 'ink';
import type { Config } from '../config/index.js';
import type { Database, Conversation, ConversationItem } from '../db/sqlite.js';
import type { Logger } from '../utils/logger.js';
import { createLLMClient } from '../llm/factory.js';
import type { LLMClient, Message } from '../llm/types.js';

interface HistoryItem {
  id: number;
  text: string;
}

interface AppProps {
  config: Config;
  db: Database;
  modelName: string;
  logger: Logger;
}

const App: React.FC<AppProps> = ({ config, db, modelName, logger }) => {
  const { exit } = useApp();
  const [messages, setMessages] = useState<Message[]>([
    { role: 'system', content: 'Welcome to Ask AI. Type your question and press Enter to send.' },
  ]);
  const [input, setInput] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [currentResponse, setCurrentResponse] = useState('');
  const [activeTab, setActiveTab] = useState<'chat' | 'history'>('chat');
  const [historyItems, setHistoryItems] = useState<HistoryItem[]>([]);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [currentConversationId, setCurrentConversationId] = useState<number | null>(null);
  const llmClientRef = useRef<LLMClient | null>(null);

  // Initialize LLM client
  useEffect(() => {
    try {
      llmClientRef.current = createLLMClient(modelName);
      logger.log('info', `Initialized LLM client for model: ${modelName}`, { modelName });
      
      // Load history
      loadConversations();
    } catch (error) {
      logger.log('error', 'Error initializing LLM client', { error: String(error) });
      setMessages(prev => [
        ...prev,
        { 
          role: 'system', 
          content: `Error initializing model: ${error instanceof Error ? error.message : String(error)}` 
        }
      ]);
    }
  }, [modelName]);

  const loadConversations = async () => {
    try {
      const recentConversations = await db.getRecentConversations(10);
      setConversations(recentConversations);
      
      // Create conversation items for display
      const items: HistoryItem[] = recentConversations.map((conv) => ({
        id: conv.id,
        text: `Conversation ${conv.id} (${conv.model})` 
      }));
      
      setHistoryItems(items);
    } catch (error) {
      logger.log('error', 'Error loading conversations', { error: String(error) });
    }
  };

  // Handle keyboard input
  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      exit();
      return;
    }

    if (key.tab) {
      setActiveTab(prev => prev === 'chat' ? 'history' : 'chat');
      return;
    }

    if (activeTab === 'history') {
      if (key.upArrow || key.downArrow) {
        // Navigate history
        return;
      }
      
      if (key.return) {
        // Load selected history item
        // For now just switch back to chat
        setActiveTab('chat');
        return;
      }
      
      return;
    }

    // Chat tab input handling
    if (key.return && input.trim() && !isGenerating) {
      const userInput = input.trim();
      sendMessage(userInput);
      setInput('');
      return;
    }

    if (key.backspace || key.delete) {
      setInput(prev => prev.slice(0, -1));
      return;
    }

    if (!key.ctrl && !key.meta && !key.return && input.length < 500) {
      setInput(prev => prev + input);
    }
  });

  const sendMessage = async (content: string) => {
    if (!content.trim() || isGenerating) return;
    
    try {
      // Add user message
      const newMessage: Message = { role: 'user', content };
      setMessages(prev => [...prev, newMessage]);
      setIsGenerating(true);
      setCurrentResponse('');
      
      if (!llmClientRef.current) {
        throw new Error('LLM client not initialized');
      }
      
      // Create or use existing conversation
      let conversationId = currentConversationId;
      if (!conversationId) {
        conversationId = await db.createConversation(modelName);
        setCurrentConversationId(conversationId);
      }
      
      // Add message to database
      await db.addConversationItem(conversationId, newMessage.role, newMessage.content);
      
      // Update history
      await loadConversations();
      
      // Prepare conversation history for the LLM
      const conversationHistory = messages
        .filter(msg => msg.role !== 'system')
        .slice(-6); // Keep last 6 messages for context
      
      // Send to LLM
      const stream = llmClientRef.current.sendStream([...conversationHistory, newMessage]);
      
      stream.on('data', (chunk: string) => {
        setCurrentResponse(prev => prev + chunk);
      });
      
      stream.on('done', async () => {
        // Add the assistant's response to the conversation in the database
        if (conversationId) {
          await db.addConversationItem(conversationId, 'assistant', currentResponse);
        }
        
        setMessages(prev => [...prev, { role: 'assistant', content: currentResponse }]);
        setIsGenerating(false);
      });
      
      stream.on('error', (error) => {
        logger.log('error', 'Stream error', { error: String(error) });
        setMessages(prev => [
          ...prev, 
          { 
            role: 'system', 
            content: `Error: ${error instanceof Error ? error.message : String(error)}` 
          }
        ]);
        setIsGenerating(false);
      });
    } catch (error) {
      logger.log('error', 'Error sending message', { error: String(error) });
      setMessages(prev => [
        ...prev, 
        { 
          role: 'system', 
          content: `Error: ${error instanceof Error ? error.message : String(error)}` 
        }
      ]);
      setIsGenerating(false);
    }
  };

  // Render chat messages
  const renderMessages = () => {
    const messagesToRender = [...messages];
    
    // Add the currently generating response
    if (isGenerating && currentResponse) {
      messagesToRender.push({ role: 'assistant', content: currentResponse });
    }
    
    return messagesToRender.map((message, index) => {
      const prefix = message.role === 'user' ? 'You: ' : 
                     message.role === 'assistant' ? 'AI: ' : 
                     'System: ';
      
      const color = message.role === 'user' ? 'green' : 
                    message.role === 'assistant' ? 'cyan' : 
                    'yellow';
      
      return (
        <Box key={index} flexDirection="column" marginBottom={1}>
          <Text color={color} bold>{prefix}</Text>
          <Text>{message.content}</Text>
        </Box>
      );
    });
  };

  // Render history tab
  const renderHistory = () => {
    return historyItems.length > 0 ? (
      historyItems.map((item, index) => (
        <Box key={item.id} marginTop={1} marginBottom={1}>
          <Text color="gray">{index + 1}.</Text>
          <Text> {item.text}</Text>
        </Box>
      ))
    ) : (
      <Text color="gray">No history items</Text>
    );
  };

  return (
    <Box flexDirection="column" height="100%">
      {/* Header */}
      <Box borderStyle="single" borderColor="gray" padding={1} marginBottom={1}>
        <Text bold color="green">Ask AI Terminal UI</Text>
        <Text color="gray"> | </Text>
        <Text color="cyan">Model: {modelName}</Text>
        <Text color="gray"> | </Text>
        <Text color={activeTab === 'chat' ? 'white' : 'gray'} underline={activeTab === 'chat'}>Chat</Text>
        <Text color="gray"> | </Text>
        <Text color={activeTab === 'history' ? 'white' : 'gray'} underline={activeTab === 'history'}>History</Text>
        <Text color="gray"> (Tab to switch)</Text>
      </Box>

      {/* Main content area */}
      <Box flexDirection="column" flexGrow={1} padding={1}>
        {activeTab === 'chat' ? renderMessages() : renderHistory()}
      </Box>

      {/* Input area */}
      <Box borderStyle="single" borderColor="gray" padding={1} marginTop={1}>
        <Text bold>{isGenerating ? 'AI is thinking...' : '> '}</Text>
        {!isGenerating && (
          <Text>{input}<Text color="gray" backgroundColor="gray">█</Text></Text>
        )}
      </Box>

      {/* Footer with help */}
      <Box marginTop={1}>
        <Text color="gray" dimColor>
          Press Tab to switch tabs | Ctrl+C to exit
        </Text>
      </Box>
    </Box>
  );
};

/**
 * Start the Ink-based TUI application
 */
export async function startInkApp(
  config: Config,
  db: Database,
  modelName: string,
  logger: Logger
): Promise<void> {
  logger.log('info', 'Starting Ink app with model', { modelName });

  try {
    // Initialize the database
    await db.initialize();
    
    // Render the Ink app
    const { waitUntilExit } = render(
      <App config={config} db={db} modelName={modelName} logger={logger} />
    );

    // Wait until the user exits the app
    await waitUntilExit();
    logger.log('info', 'Ink app exited normally');
  } catch (error) {
    logger.log('error', 'Error in Ink app', { error: String(error) });
    throw error;
  }
}