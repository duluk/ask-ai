package ollama

import (
	"bytes"
	"fmt"
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

func TestChatCompletionStream_Success(t *testing.T) {
	// Prepare raw JSON chunks
	raw1 := `{"choices":[{"delta":{"content":"chunk1"}}]}`
	raw2 := `{"choices":[{"delta":{"content":"chunk2"}}]}`
	// Build SSE-like stream data
	data := fmt.Sprintf("data: %s\n", raw1) + fmt.Sprintf("data: %s\n", raw2) + "data: [DONE]\n"

	// Set up client with stub transport
	client := NewClient("key", "unused")
	client.HTTPClient = &http.Client{Transport: &stubTransport{fn: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(data))),
			Header:     make(http.Header),
		}, nil
	}}}

	var collected []ChatCompletionChunk
	err := client.ChatCompletionStream(ChatCompletionRequest{}, func(chunk ChatCompletionChunk) {
		collected = append(collected, chunk)
	})
	assert.NoError(t, err)
	assert.Len(t, collected, 2)
	assert.Equal(t, "chunk1", collected[0].Choices[0].Delta.Content)
	assert.Equal(t, "chunk2", collected[1].Choices[0].Delta.Content)
}

func TestChatCompletionStream_HTTPError(t *testing.T) {
	// Stub transport returns 500 error with body
	client := NewClient("key", "unused")
	client.HTTPClient = &http.Client{Transport: &stubTransport{fn: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewReader([]byte("server error"))),
			Header:     make(http.Header),
		}, nil
	}}}

	err := client.ChatCompletionStream(ChatCompletionRequest{}, func(chunk ChatCompletionChunk) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 500: server error")
}

func TestChatCompletionStream_UnmarshalError(t *testing.T) {
	// Stub transport returns invalid JSON
	data := "data: {invalid json}\n"
	client := NewClient("key", "unused")
	client.HTTPClient = &http.Client{Transport: &stubTransport{fn: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(data))),
			Header:     make(http.Header),
		}, nil
	}}}

	err := client.ChatCompletionStream(ChatCompletionRequest{}, func(chunk ChatCompletionChunk) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshaling chunk")
}
