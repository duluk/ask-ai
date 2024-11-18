package LLM

import (
	"context"
	"fmt"
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

func NewGoogle(maxTokens int) *Google {
	apiKey := getClientKey("google")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}

	return &Google{APIKey: apiKey, Tokens: maxTokens, Client: client, Context: ctx}
}

func (cs *Google) SimpleChat(args ClientArgs) error {
	client := cs.Client
	ctx := cs.Context

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetMaxOutputTokens(int32(*args.MaxTokens))
	resp, err := model.GenerateContent(ctx, genai.Text(*args.Prompt))
	if err != nil {
		return err
	}

	fmt.Printf("%s", resp.Candidates[0].Content.Parts[0])
	fmt.Printf("\n<------>\n")

	return nil
}

func (cs *Google) Chat(args ClientArgs) (string, error) {
	client := cs.Client
	ctx := cs.Context

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetTemperature(*args.Temperature)
	model.SetMaxOutputTokens(int32(*args.MaxTokens))
	model.SystemInstruction = genai.NewUserContent(genai.Text(*args.SystemPrompt))
	// model.SetTopP(0.9)
	// model.SetTopK(40)
	// model.ResponseMIMEType = "application/json"

	var resp_str string
	prompt := buildPrompt(args.Context, *args.Prompt)
	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	wrapper := linewrap.NewLineWrapper(TermWidth, TabWidth)
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", err
		}

		r := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
		resp_str += r
		wrapper.Write([]byte(r))
	}

	return resp_str, nil
}
