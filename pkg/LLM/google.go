package LLM

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/duluk/ask-ai/pkg/linewrap"

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
	apiKey := getClientKey("google")
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

	model := client.GenerativeModel("gemini-1.5-pro")
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

func (cs *Google) Chat(args ClientArgs, termWidth int, tabWidth int) (ClientResponse, error) {
	client := cs.Client
	ctx := cs.Context

	model := client.GenerativeModel("gemini-1.5-pro")
	// model := client.GenerativeModel("gemini-exp-1114")
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
	wrapper := linewrap.NewLineWrapper(termWidth, tabWidth, os.Stdout)
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
		wrapper.Write([]byte(r))

		usage = resp.UsageMetadata
	}

	// I believe the stats object will be usable even if the response is empty
	r := ClientResponse{
		Text:         resp_str,
		InputTokens:  usage.PromptTokenCount,
		OutputTokens: usage.CandidatesTokenCount,
		MyEstInput:   myInputEstimate,
	}

	return r, nil
}

// Add this method to the Google struct
func (cs *Google) ChatStream(args ClientArgs, termWidth int, tabWidth int, stream chan<- StreamResponse) error {
	// Not yet implemented - just use the non-streaming version for now
	resp, err := cs.Chat(args, termWidth, tabWidth)
	if err != nil {
		stream <- StreamResponse{
			Content: "",
			Done:    true,
			Error:   err,
		}
		return err
	}

	// Send the full response as one chunk
	stream <- StreamResponse{
		Content: resp.Text,
		Done:    false,
		Error:   nil,
	}

	// Signal completion
	stream <- StreamResponse{
		Content: "",
		Done:    true,
		Error:   nil,
	}

	return nil
}
