package LLM

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func buildPrompt(msgCtx []LLMConversations, newPrompt string) string {
	var prompt strings.Builder
	for _, msg := range msgCtx {
		prompt.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}
	prompt.WriteString(fmt.Sprintf("user: %s", newPrompt))
	return prompt.String()
}

func NewGoogle() *Google {
	apiKey, err := getClientKey("google")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}

	return &Google{APIKey: apiKey, Client: client, Context: ctx}
}

func (cs *Google) SimpleChat(args ClientArgs) error {
	client := cs.Client
	ctx := cs.Context

	// Use configured model name (e.g., gemini-2.0-flash-001) instead of hardcoded value
	modelName := *args.Model
	model := client.GenerativeModel(modelName)
	// model := client.GenerativeModel("gemini-exp-1114")
	model.SetMaxOutputTokens(int32(*args.MaxTokens))
	resp, err := model.GenerateContent(ctx, genai.Text(*args.Prompt))
	if err != nil {
		return err
	}

	fmt.Printf("%s", resp.Candidates[0].Content.Parts[0])
	fmt.Printf("\n<------>\n")

	return nil
}

func (cs *Google) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, <-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse)

	var resp ClientResponse
	var err error
	go func() {
		defer close(responseChan)

		// Use the streaming implementation
		resp, err = cs.ChatStream(args, termWidth, tabWidth, responseChan)
		if err != nil {
			responseChan <- StreamResponse{
				Content: "",
				Done:    true,
				Error:   err,
			}
		}
	}()

	return resp, responseChan, nil
}

func (cs *Google) ChatStream(args ClientArgs, termWidth int, tabWidth int, stream chan<- StreamResponse) (ClientResponse, error) {
	client := cs.Client
	ctx := cs.Context

	// Use configured model name provided via args.Model
	modelName := *args.Model
	model := client.GenerativeModel(modelName)
	model.SetTemperature(*args.Temperature)
	model.SetMaxOutputTokens(int32(*args.MaxTokens))
	model.SystemInstruction = genai.NewUserContent(genai.Text(*args.SystemPrompt))
	// model.SetTopP(0.9)
	// model.SetTopK(40)
	// model.ResponseMIMEType = "application/json"

	var resp_str string
	var usage *genai.UsageMetadata
	prompt := buildPrompt(args.Context, *args.Prompt)
	myInputEstimate := EstimateTokens(prompt + *args.SystemPrompt)

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return ClientResponse{}, err
		}

		r := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
		resp_str += r

		stream <- StreamResponse{
			Content: r,
			Done:    false,
			Error:   nil,
		}

		usage = resp.UsageMetadata
	}

	// TODO: do we need to check for errors?
	stream <- StreamResponse{
		Content: "",
		Done:    true,
		Error:   nil,
	}

	// I believe the stats object will be usable even if the response is empty
	// TODO: confirm this is working and passed back
	r := ClientResponse{
		Text:         resp_str,
		InputTokens:  usage.PromptTokenCount,
		OutputTokens: usage.CandidatesTokenCount,
		MyEstInput:   myInputEstimate,
	}

	return r, nil
}
