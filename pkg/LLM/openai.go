package LLM

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func NewOpenAI(apiLLC string, apiURL string) *OpenAI {
	apiKey := getClientKey(apiLLC)
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(apiURL),
	)

	return &OpenAI{APIKey: apiKey, Client: client}
}

func (cs *OpenAI) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, error) {
	client := cs.Client

	var msgCtx string

	for _, msg := range args.Context {
		msg.Role = strings.ToLower(msg.Role)
		switch msg.Role {
		case "user":
			msgCtx += "User: " + msg.Content + "\n"
		case "assistant":
			msgCtx += "Assistant: " + msg.Content + "\n"
		}
	}

	const ChatModelGrokBeta openai.ChatModel = "grok-beta"
	const ChatModelGrok2 openai.ChatModel = "grok-2-latest"
	const ChatModelDeepSeekChat openai.ChatModel = "deepseek-chat"
	const ChatModelDeepSeekReasoner openai.ChatModel = "deepseek-reasoner"

	var model openai.ChatModel
	switch *args.Model {
	case "chatgpt":
		// N.B. - the o1 models are a little weird, or I'm using them wrong.
		// They don't support temperature and they generate a ton of text,
		// immediately, just cutting off at max tokens. Switching to GPT-4o and
		// asking the same prompts resulted in much more reasonable responses.
		// model = openai.ChatModelO1Preview2024_09_12
		// model = openai.ChatModelO1Mini // Tailored for coding and math
		model = openai.ChatModelGPT4o
	case "grok":
		model = ChatModelGrok2
	case "deepseek":
		model = ChatModelDeepSeekReasoner
	}

	myInputEstimate := EstimateTokens(msgCtx + *args.Prompt + *args.SystemPrompt)
	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(*args.SystemPrompt),
				openai.AssistantMessage(msgCtx),
				openai.UserMessage(*args.Prompt),
			}),
			// N.B. - not all openai models support these, eg the o1 models (as of 20241125) and temperature
			// Seed:        openai.Int(1), // Same seed/parameters will attempt to return the same results
			Model:               openai.F(model),
			MaxCompletionTokens: openai.Int(int64(*args.MaxTokens)),
			Temperature:         openai.Float(float64(*args.Temperature)), // Controls randomness (0.0 to 2.0)
			StreamOptions: openai.F(openai.ChatCompletionStreamOptionsParam{
				IncludeUsage: openai.Bool(true),
			}),

			// TopP:             openai.Float(1.0),                // Controls diversity via nucleus sampling; alter this or Temperature but not both
			// N:                openai.Int(1),                    // Number of completions to generate
			// ResponseFormat:   openai.ChatResponseFormatDefault, // Format of the response
			// Stop:             openai.String("\n"),              // Stop completion at this token
		},
	)

	// Apparently what happens with stream is that the server chunks the
	// response according to its own internal desires and whims, presenting the
	// result as if it's a stream of responses, which looks more
	// conversational.
	var resp string
	var usage *openai.CompletionUsage
	// wrapper := linewrap.NewLineWrapper(termWidth, tabWidth, os.Stdout)
	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			data := evt.Choices[0].Delta.Content
			// if _, err := wrapper.Write([]byte(data)); err != nil {
			// 	fmt.Printf("Error writing to wrapper: %s\n", err)
			// 	return ClientResponse{}, err
			// }
			resp += data
		}
		usage = &evt.Usage
	}

	var stats *openai.CompletionUsage
	if usage != nil {
		stats = usage
	}

	if stream.Err() != nil {
		fmt.Printf("Error: %s\n", stream.Err())
		return ClientResponse{}, stream.Err()
	}

	r := ClientResponse{
		Text:         resp,
		InputTokens:  int32(stats.PromptTokens),
		OutputTokens: int32(stats.CompletionTokens), // Is this correct?
		MyEstInput:   myInputEstimate,
	}

	return r, nil
}

// StreamChat implements streaming response for OpenAI providers
func (cs *OpenAI) StreamChat(args ClientArgs, termWidth int, tabWidth int, callback func(chunk string)) (ClientResponse, error) {
	client := cs.Client

	var msgCtx string

	for _, msg := range args.Context {
		msg.Role = strings.ToLower(msg.Role)
		switch msg.Role {
		case "user":
			msgCtx += "User: " + msg.Content + "\n"
		case "assistant":
			msgCtx += "Assistant: " + msg.Content + "\n"
		}
	}

	const ChatModelGrokBeta openai.ChatModel = "grok-beta"
	const ChatModelGrok2 openai.ChatModel = "grok-2-latest"
	const ChatModelDeepSeekChat openai.ChatModel = "deepseek-chat"
	const ChatModelDeepSeekReasoner openai.ChatModel = "deepseek-reasoner"

	var model openai.ChatModel
	switch *args.Model {
	case "chatgpt":
		model = openai.ChatModelGPT4o
	case "grok":
		model = ChatModelGrok2
	case "deepseek":
		model = ChatModelDeepSeekReasoner
	}

	myInputEstimate := EstimateTokens(msgCtx + *args.Prompt + *args.SystemPrompt)
	ctx := context.Background()
	stream := client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(*args.SystemPrompt),
				openai.AssistantMessage(msgCtx),
				openai.UserMessage(*args.Prompt),
			}),
			Model:               openai.F(model),
			MaxCompletionTokens: openai.Int(int64(*args.MaxTokens)),
			Temperature:         openai.Float(float64(*args.Temperature)),
			StreamOptions: openai.F(openai.ChatCompletionStreamOptionsParam{
				IncludeUsage: openai.Bool(true),
			}),
		},
	)

	var resp string
	var usage *openai.CompletionUsage
	var needsSpace bool

	// Helper function to check if a rune is punctuation that shouldn't have spaces before it
	isPunctuation := func(r rune) bool {
		return strings.ContainsRune(",.!?:;\"')", r)
	}

	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 && evt.Choices[0].Delta.Content != "" {
			content := evt.Choices[0].Delta.Content

			// Add space if needed, but not before punctuation
			if needsSpace &&
				!strings.HasPrefix(content, " ") &&
				!strings.HasPrefix(content, "\n") &&
				len(content) > 0 &&
				!isPunctuation([]rune(content)[0]) {
				callback(" ")
				resp += " "
			}

			callback(content)
			resp += content

			// Set flag for next chunk if this one ends with a word
			if len(content) > 0 {
				lastRune := []rune(content)[len([]rune(content))-1]
				needsSpace = !strings.HasSuffix(content, " ") &&
					!strings.HasSuffix(content, "\n") &&
					!strings.HasSuffix(content, "\t") &&
					!isPunctuation(lastRune)
			}
		}
		usage = &evt.Usage
	}

	var stats *openai.CompletionUsage
	if usage != nil {
		stats = usage
	}

	if stream.Err() != nil {
		return ClientResponse{}, stream.Err()
	}

	r := ClientResponse{
		Text:         resp,
		InputTokens:  int32(stats.PromptTokens),
		OutputTokens: int32(stats.CompletionTokens),
		MyEstInput:   myInputEstimate,
	}

	return r, nil
}
