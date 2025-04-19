package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stubTransport implements http.RoundTripper for testing
type stubTransport struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return s.fn(req)
}

func TestCreateChatCompletion_Success(t *testing.T) {
	// Prepare a successful response
	respObj := ChatCompletionResponse{
		ID:      "1",
		Object:  "chat.completion",
		Created: 123,
		Choices: []Choice{{
			Index:        0,
			Message:      Message{Role: "assistant", Content: "hello"},
			FinishReason: "stop",
		}},
		Usage: Usage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3},
	}
	b, err := json.Marshal(respObj)
	assert.NoError(t, err)

	client := NewClient("test-key")
	client.HTTPClient = &http.Client{Transport: &stubTransport{fn: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(b)),
			Header:     make(http.Header),
		}, nil
	}}}

	result, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "1", result.ID)
	assert.Len(t, result.Choices, 1)
	assert.Equal(t, "hello", result.Choices[0].Message.Content)
}

func TestCreateChatCompletion_HTTPError(t *testing.T) {
	client := NewClient("key")
	client.HTTPClient = &http.Client{Transport: &stubTransport{fn: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewReader([]byte("bad request"))),
			Header:     make(http.Header),
		}, nil
	}}}

	_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400: bad request")
}

func TestCreateChatCompletion_TransportError(t *testing.T) {
	client := NewClient("key")
	client.HTTPClient = &http.Client{Transport: &stubTransport{fn: func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network failure")
	}}}

	_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestCreateChatCompletion_DecodeError(t *testing.T) {
	client := NewClient("key")
	client.HTTPClient = &http.Client{Transport: &stubTransport{fn: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("not a json"))),
			Header:     make(http.Header),
		}, nil
	}}}

	_, err := client.CreateChatCompletion(context.Background(), ChatCompletionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestChatCompletionRequest_MarshalUnmarshal(t *testing.T) {
	req := ChatCompletionRequest{
		Model:       "m1",
		Messages:    []Message{{Role: "user", Content: "hi"}},
		Temperature: 0.5,
		MaxTokens:   10,
		Stream:      true,
	}
	data, err := json.Marshal(req)
	assert.NoError(t, err)
	var other ChatCompletionRequest
	err = json.Unmarshal(data, &other)
	assert.NoError(t, err)
	assert.Equal(t, req, other)
}
