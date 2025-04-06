package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const OllamaBaseURL = "http://bamf.midgaard.xyz:11434/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
	Usage struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
		TotalTokens      int32 `json:"total_tokens"`
	} `json:"usage"`
}

type Client struct {
	APIKey     string
	HTTPClient *http.Client
	BaseURL    string
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
		BaseURL:    baseURL,
	}
}

type ChatCompletionChunk struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason any `json:"finish_reason"`
	} `json:"choices"`
}

type StreamHandler func(ChatCompletionChunk)

func (c *Client) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	requestData := ChatCompletionRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url, err := url.Parse(c.BaseURL)
	if err != nil {
		log.Fatal(err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		url.String(),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("User-Agent", "ask-ai/0.0.3")
	httpReq.Header.Set("Content-Type", "application/json")
	// If this is empty, it's fine (and probably the default)
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make client request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ChatCompletionResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			fmt.Printf("Error decoding error response: %v\n", err)
			return nil, fmt.Errorf("error decoding client response: %v", err)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorResp.Error.Message)
	}

	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

func (c *Client) ChatCompletionStream(req ChatCompletionRequest, handler StreamHandler) error {
	req.Stream = true // Ensure streaming is enabled

	// Use the same request structure as the non-streaming version
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	url, err := url.Parse(OllamaBaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		url.String(),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("User-Agent", "ask-ai/0.0.3")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream: %v", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Create a file and write the line to it
		// f, err := os.Create("raw_data.json")
		// if err != nil {
		// 	return fmt.Errorf("error creating file: %v", err)
		// }
		// defer f.Close()
		// f.Write(line) // I don't care if it fails

		// Strip "data: " prefix from SSE format
		if bytes.HasPrefix(line, []byte("data: ")) {
			line = bytes.TrimPrefix(line, []byte("data: "))
		}

		// Skip any "data: [DONE]" messages
		if string(line) == "[DONE]" {
			continue
		}

		var chunk ChatCompletionChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			return fmt.Errorf("error unmarshaling chunk: %v\nRaw data: %s", err, string(line))
		}

		handler(chunk)
	}

	return nil
}

// Add this new method
// func (c *Client) ChatCompletionStream(req ChatCompletionRequest, handler StreamHandler) error {
// 	req.Stream = true // Ensure streaming is enabled
//
// 	requestBody, err := json.Marshal(req)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal request: %v", err)
// 	}
//
// 	resp, err := c.HTTPClient.Post(OllamaBaseURL, "application/json", bytes.NewBuffer(requestBody))
// 	if err != nil {
// 		return fmt.Errorf("failed to send request: %v", err)
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
// 	}
//
// 	reader := bufio.NewReader(resp.Body)
// 	for {
// 		line, err := reader.ReadBytes('\n')
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return fmt.Errorf("error reading stream: %v", err)
// 		}
//
// 		line = bytes.TrimSpace(line)
// 		if len(line) == 0 {
// 			continue
// 		}
//
// 		var chunk ChatCompletionChunk
// 		if err := json.Unmarshal(line, &chunk); err != nil {
// 			return fmt.Errorf("error unmarshaling chunk: %v", err)
// 		}
//
// 		handler(chunk)
// 	}
//
// 	return nil
// }
