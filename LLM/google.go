package LLM

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

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
	out := args.Out

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetMaxOutputTokens(int32(args.Max_Tokens))
	resp, err := model.GenerateContent(ctx, genai.Text(args.Prompt))
	if err != nil {
		return err
	}

	out.Printf("%s", resp.Candidates[0].Content.Parts[0])
	out.Printf("\n<------>\n")

	return nil
}

// Some configuration options for the model:
// model.SetTopP(0.9)
// model.SetTopK(40)
// model.SystemInstruction = genai.NewUserContent(genai.Text("You are Yoda from Star Wars."))
// model.ResponseMIMEType = "application/json"
func (cs *Google) Chat(args Client_Args) error {
	client := cs.Client
	ctx := cs.Context
	out := args.Out

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetTemperature(0.3)
	model.SetMaxOutputTokens(int32(args.Max_Tokens))

	out.Printf("Assistant: ")
	iter := model.GenerateContentStream(ctx, genai.Text(args.Prompt))
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		out.Printf("%s", resp.Candidates[0].Content.Parts[0])
	}

	return nil
}
