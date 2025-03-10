import React, { FC, useState, useEffect } from 'react';
// Import ink path without js extension for CommonJS mode
import { Box, Text, useInput, useApp } from 'ink';
import { Database } from '../db/sqlite.js';
import { createLLMClient } from '../llm/factory.js';
import { Message, ModelProvider } from '../llm/types.js';

// Define message type
export interface ChatMessage {
    role: string;
    content: string;
}

interface ConversationOutputProps {
    messages: ChatMessage[];
    isTyping: boolean;
}

// Component for displaying conversation messages
export const ConversationOutput: FC<ConversationOutputProps> = ({ messages, isTyping }) => {
    return (
        <Box
            flexDirection="column"
            height="100%"
            borderStyle="single"
            borderColor="blue"
            flexGrow={1}
            padding={1}
            overflow="auto"
        >
            {messages.map((msg, index) => (
                <Box key={index} flexDirection="column" marginBottom={1}>
                    <Text bold color={msg.role === 'user' ? 'green' : 'cyan'}>
                        {msg.role === 'user' ? 'You:' : msg.role === 'assistant' ? 'AI:' : 'System:'}
                    </Text>
                    <Text wrap="wrap" paddingLeft={1}>{msg.content}</Text>
                </Box>
            ))}
            {isTyping && (
                <Box>
                    <Text color="cyan">AI is typing...</Text>
                </Box>
            )}
        </Box>
    );
};

interface StatusLineAreaProps {
    isTyping: boolean;
    modelName: string;
}

export const StatusLineArea: FC<StatusLineAreaProps> = ({ isTyping, modelName }) => {
    return (
        <Box
            flexDirection="row"
            justifyContent="space-between"
            width="100%"
        >
            <Box
                width="100%"
                paddingLeft={1}
                paddingRight={1}
            >
                <Text
                    // backgroundColor="blue"
                    color="yellow"
                    wrap="truncate"
                    paddingLeft={1}
                    paddingRight={1}
                >
                    {isTyping
                        ? 'AI is thinking...'
                        : `Model: ${modelName} | Use Ctrl+N for newline | Press Esc to quit`}
                </Text>
            </Box>
        </Box >
    );
};

interface InputAreaProps {
    value: string;
    onChange: (value: string) => void;
    onSubmit: () => void;
    isTyping: boolean;
}

// Component for the input area
export const InputArea: FC<InputAreaProps> = ({ value, onChange, onSubmit, isTyping }) => {
    const { exit } = useApp();

    useInput((input: string, key) => {
        if (key.escape) {
            exit();
        } else if (key.return) {
            if (!isTyping && value.trim() !== '') {
                onSubmit();
            }
        } else if (key.ctrl && input === 'n') {
            onChange(value + '\n');
        } else if (key.backspace || key.delete) {
            onChange(value.slice(0, -1));
        } else if (input && !key.ctrl && !key.meta) {
            onChange(value + input);
        }
    });

    return (
        <Box flexDirection="column" width="100%">
            <Box
                borderStyle="single"
                borderColor={isTyping ? "gray" : "green"}
                minHeight={3}
                alignItems="flex-end"
            >
                <Text>{value}</Text>
            </Box>
        </Box>
    );
};

interface AskAITUIProps {
    config: any;
    db: Database;
    modelName: string;
    logger: any;
}

// Main TUI component
export const AskAITUI: FC<AskAITUIProps> = ({ config, db, modelName, logger }) => {
    const [conversationId, setConversationId] = useState<number | null>(null);
    const [messages, setMessages] = useState<ChatMessage[]>([
        {
            role: 'system',
            content: 'Welcome to Ink 5 Ask AI Terminal UI\nType your question below and press Enter\nPress Ctrl+N to insert a new line'
        }
    ]);
    const [input, setInput] = useState<string>('');
    const [isTyping, setIsTyping] = useState<boolean>(false);
    const [streamContent, setStreamContent] = useState<string>('');

    // Initialize conversation
    useEffect(() => {
        const initConversation = async () => {
            try {
                const id = await db.createConversation(modelName);
                setConversationId(id);
            } catch (error) {
                setMessages(prev => [...prev, {
                    role: 'system',
                    content: `Error initializing conversation: ${error}`
                }]);
            }
        };

        initConversation();

        return () => {
            // Clean up function to close the database when component unmounts
            db.close().catch(console.error);
        };
    }, [db, modelName]);

    // Function to send a message to the AI
    const sendMessage = async (query: string) => {
        try {
            if (!conversationId) {
                return; // Wait until we have a conversation ID
            }

            if (!query.trim()) {
                return;
            }

            // Add user message to UI
            setMessages(prev => [...prev, { role: 'user', content: query }]);
            setIsTyping(true);
            setStreamContent('');

            // Add to database
            await db.addConversationItem(conversationId, 'user', query);

            // Prepare the context
            const contextMessages: Message[] = await db.getMessagesForLLM(conversationId, 10);

            // Create LLM client
            const llmClient = createLLMClient(modelName);

            try {
                // Get streaming response
                const stream = llmClient.sendStream(contextMessages);
                let fullResponse = '';

                stream.on('data', (chunk) => {
                    fullResponse += chunk;
                    setStreamContent(fullResponse);
                });

                stream.on('done', async (response) => {
                    // Add AI response to UI
                    setMessages(prev => [
                        ...prev,
                        { role: 'assistant', content: fullResponse }
                    ]);

                    // Record the response to the database
                    await db.addConversationItem(
                        conversationId,
                        'assistant',
                        fullResponse,
                        response.usage?.promptTokens || 0,
                        response.usage?.completionTokens || 0
                    );

                    setIsTyping(false);
                    setStreamContent('');
                });

                stream.on('error', (error) => {
                    setMessages(prev => [...prev, {
                        role: 'system',
                        content: `Error from AI: ${error}`
                    }]);
                    setIsTyping(false);
                    setStreamContent('');
                });
            } catch (error) {
                setMessages(prev => [...prev, {
                    role: 'system',
                    content: `Error communicating with AI: ${error}`
                }]);
                setIsTyping(false);
            }
        } catch (error) {
            setMessages(prev => [...prev, {
                role: 'system',
                content: `Error: ${error}`
            }]);
            setIsTyping(false);
        }
    };

    // Handle input submission
    const handleSubmit = () => {
        if (!isTyping && input.trim()) {
            const query = input.trim();
            setInput('');
            sendMessage(query);
        }
    };

    // Combined messages including streaming content
    const displayMessages = [
        ...messages,
        ...(isTyping && streamContent ? [{ role: 'assistant', content: streamContent }] : [])
    ];

    return (
        <Box flexDirection="column" height="100%">
            <Box flexGrow={1} minHeight={10}>
                <ConversationOutput
                    messages={displayMessages}
                    isTyping={isTyping && !streamContent}
                />
            </Box>
            <StatusLineArea
                isTyping={isTyping}
                modelName={modelName}
            />
            <InputArea
                value={input}
                onChange={setInput}
                onSubmit={handleSubmit}
                isTyping={isTyping}
            />
        </Box >
    );
};
