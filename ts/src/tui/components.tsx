import { FC, useState, useEffect, useRef } from 'react';
// Import ink path without js extension for CommonJS mode
import { Box, Text, useInput, useApp } from 'ink';
import Spinner from 'ink-spinner';
import { Database } from '../db/sqlite.js';
import { createLLMClient } from '../llm/factory.js';
import { Message } from '../llm/types.js';

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
    // Remove duplicate scrolling logic - we'll rely on the parent component


    return (
        <Box
            flexDirection="column"
            borderStyle="single"
            borderColor="blue"
            flexGrow={1}
            padding={0}
        >
            <Box flexDirection="column" padding={1} flexGrow={1}>
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
                        <Spinner type="circleHalves" />
                        <Text color="cyan"> AI is typing...</Text>
                    </Box>
                )}
            </Box>
        </Box>
    );
};

interface StatusLineAreaProps {
    isTyping: boolean;
    modelName: string;
    totalMessages: number;
}

export const StatusLineArea: FC<StatusLineAreaProps> = ({
    isTyping,
    modelName,
    totalMessages
}) => {

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
                    color="yellow"
                    wrap="truncate"
                    paddingLeft={1}
                    paddingRight={1}
                >
                    {isTyping
                        ? 'AI is thinking...'
                        : `Model: ${modelName} | Total messages: ${totalMessages} | Ctrl+N: Newline | Esc: Quit`}
                </Text>
            </Box>
        </Box>
    );
};

interface InputAreaProps {
    value: string;
    // onChange: (value: string) => void;
    // onSubmit: () => void;
    isTyping: boolean;
}

// Component for the input area
export const InputArea: FC<InputAreaProps> = ({ value, isTyping }) => {
    // Add cursor blinking state
    const [showCursor, setShowCursor] = useState<boolean>(true);

    // Create blinking cursor effect with optimized implementation
    useEffect(() => {
        if (isTyping) {
            setShowCursor(false);
            return;
        }

        const timer = setInterval(() => {
            setShowCursor(prev => !prev);
        }, 500); // cursor blink rate in ms

        return () => clearInterval(timer);
    }, [isTyping]);

    // ▁ or ▊ or █ can also be used for cursor
    return (
        <Box flexDirection="column" width="100%">
            <Box
                borderStyle="single"
                borderColor={isTyping ? "gray" : "green"}
                minHeight={3}
                alignItems="flex-start"
                paddingX={1}
            >
                <Box flexDirection="row">
                    <Text color={isTyping ? "gray" : "cyan"}>
                        {value}
                        {!isTyping && "⎸"}
                    </Text>
                </Box>
            </Box>
        </Box>
    );
};

interface AskAITUIProps {
    // config: any;
    // logger: any;
    db: Database;
    modelName: string;
}

// Main TUI component
export const AskAITUI: FC<AskAITUIProps> = ({ db, modelName }) => {
    const { exit } = useApp(); // Move this to the top level
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
    const [isStreaming, setIsStreaming] = useState(false);
    const [scrollOffset, setScrollOffset] = useState(0);

    // Initialize conversation
    useEffect(() => {
        const initConversation = async () => {
            try {
                const id = await db.createConversation(modelName);
                setConversationId(id);
                setScrollOffset(0); // Reset scroll position
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

    // Handle input submission
    const handleSubmit = () => {
        if (!isTyping && input.trim()) {
            const query = input.trim();
            setInput('');
            sendMessage(query);
        }
    };

    // Handle keyboard input for both scrolling and text input
    useInput((input: string, key) => {
        if (key.escape) {
            exit();
        } else if (key.return) {
            handleSubmit();
        } else if (key.ctrl && input === 'n') {
            setInput(prev => prev + '\n');
        } else if (key.backspace || key.delete) {
            setInput(prev => prev.slice(0, -1));
        } else if (!key.ctrl && !key.meta && !key.upArrow && !key.downArrow) {
            // Simply append the input character without checking length
            setInput(prev => prev + input);
        }
    });

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
                    setIsStreaming(true);
                    fullResponse += chunk;
                    setStreamContent(fullResponse);
                });

                stream.on('done', async (response) => {
                    setIsStreaming(false);
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

    // Combined messages including streaming content
    const displayMessages = [
        ...messages,
        ...(isTyping && streamContent ? [{ role: 'assistant', content: streamContent }] : [])
    ];

    return (
        <Box flexDirection="column" height="100%">
            <Box>
                <ConversationOutput
                    messages={displayMessages}
                    isTyping={isTyping && !streamContent}
                />
            </Box>
            <StatusLineArea
                isTyping={isTyping}
                modelName={modelName}
                totalMessages={messages.length}
            />
            <InputArea
                value={input}
                isTyping={isTyping}
            />
        </Box>
    );
};
