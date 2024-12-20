package LLM

// I originally intended to use some sort of summarization API like tf-ipf or
// RAKE, but it proved to be less than impressive...to me at least. I could
// perhaps have spent more time tweaking the functions and finding the right
// package (or generating my own). However, I decided that since LLMs are
// excellent at text summarization, I might as well just use one.

// Thus, the summarizer currently just uses OpenAI's gpt-3.5-turbo, so an
// OpenAI API key is required, in addition to whatever other model is being
// used.

// Perhaps one day I'll revisit this and try to implement a more traditional
// text summarization method. Or allow other APIs to be used.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type GPTRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type GPTResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Summarizer struct {
	APIKey string
}

func NewSummarizer(apiClient string) *Summarizer {
	apiKey := getClientKey(apiClient)
	return &Summarizer{APIKey: apiKey}
}

func (s *Summarizer) GenerateSummary(text string, prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	if prompt == "" {
		prompt = "Summarize the following text in one line, capturing the general topic of the text:\n\n" + text
	}

	requestBody := GPTRequest{
		Model: DefaultSummaryModel,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{{Role: "user", Content: prompt}},
		MaxTokens:   50,
		Temperature: 0.7,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New("API error: " + string(body))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var gptResponse GPTResponse
	if err := json.Unmarshal(responseBody, &gptResponse); err != nil {
		return "", err
	}

	if len(gptResponse.Choices) == 0 {
		return "", errors.New("no summary generated")
	}

	return gptResponse.Choices[0].Message.Content, nil
}
