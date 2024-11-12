package LLM

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func build_prompt(msg_context []LLM_Conversations, new_prompt string) string {
	var prompt strings.Builder
	for _, msg := range msg_context {
		prompt.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}
	prompt.WriteString(fmt.Sprintf("user: %s", new_prompt))
	return prompt.String()
}

func New_Google(max_tokens int) *Google {
	api_key := get_client_key("google")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(api_key))
	if err != nil {
		panic(err)
	}

	return &Google{API_Key: api_key, Tokens: max_tokens, Client: client, Context: ctx}
}

func (cs *Google) Simple_Chat(args Client_Args) error {
	client := cs.Client
	ctx := cs.Context

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetMaxOutputTokens(int32(*args.Max_Tokens))
	resp, err := model.GenerateContent(ctx, genai.Text(*args.Prompt))
	if err != nil {
		return err
	}

	fmt.Printf("%s", resp.Candidates[0].Content.Parts[0])
	fmt.Printf("\n<------>\n")

	return nil
}

func (cs *Google) Chat(args Client_Args) (string, error) {
	client := cs.Client
	ctx := cs.Context

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetTemperature(*args.Temperature)
	model.SetMaxOutputTokens(int32(*args.Max_Tokens))
	model.SystemInstruction = genai.NewUserContent(genai.Text(*args.System_Prompt))
	// model.SetTopP(0.9)
	// model.SetTopK(40)
	// model.ResponseMIMEType = "application/json"

	var resp_str string
	prompt := build_prompt(args.Context, *args.Prompt)
	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", err
		}

		resp_str += fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
		fmt.Printf("%s", resp.Candidates[0].Content.Parts[0])
	}

	return resp_str, nil
}
