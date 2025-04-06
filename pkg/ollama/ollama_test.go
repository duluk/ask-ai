package ollama

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey, "http://example.com")
	if client.APIKey != apiKey {
		t.Errorf("expected API key to be %q, got %q", apiKey, client.APIKey)
	}
	if client.HTTPClient == nil {
		t.Errorf("expected HTTP client to be non-nil")
	}
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name          string
		request       ChatCompletionRequest
		mockResponse  string
		mockStatus    int
		expectedError string
	}{
		{
			name: "successful request",
			request: ChatCompletionRequest{
				Model:    "test-model",
				Messages: []Message{{Role: "user", Content: "test"}},
			},
			mockResponse:  `{"choices": [{"message": {"content": "response"}}]}`,
			mockStatus:    http.StatusOK,
			expectedError: "",
		},
		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.
				ResponseWriter, r *http.Request,
			) {
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			client := NewClient("test-key", server.URL)
			client.HTTPClient = server.Client()

			resp, err := client.ChatCompletion(tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestChatCompletionRequestMarshaling(t *testing.T) {
	req := ChatCompletionRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role:    "user",
				Content: "test message",
			},
		},
		MaxTokens:   100,
		Temperature: 0.5,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Errorf("failed to marshal request: %v", err)
	}

	var unmarshaledReq ChatCompletionRequest
	if err := json.Unmarshal(jsonData, &unmarshaledReq); err != nil {
		t.Errorf("failed to unmarshal request: %v", err)
	}

	if !reflect.DeepEqual(req, unmarshaledReq) {
		t.Errorf("marshaled and unmarshaled requests do not match")
	}
}

func TestChatCompletionSuccess(t *testing.T) {
	apiKey := "test-api-key"

	req := ChatCompletionRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role:    "user",
				Content: "test message",
			},
		},
		MaxTokens:   100,
		Temperature: 0.5,
	}

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %q", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer "+apiKey {
			t.Errorf("expected Authorization header to be %q, got %q", "Bearer "+apiKey,
				r.Header.Get("Authorization"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type header to be %q, got %q",
				"application/json", r.Header.Get("Content-Type"))
		}

		var unmarshaledReq ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&unmarshaledReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if !reflect.DeepEqual(req, unmarshaledReq) {
			t.Errorf("marshaled and unmarshaled requests do not match")
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices": [{"message": {"content": "test response"}}]}`)
	}))

	defer httpServer.Close()

	client := NewClient(apiKey, httpServer.URL)
	client.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	resp, err := client.ChatCompletion(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Errorf("expected response to be non-nil")
	}

	if len(resp.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "test response" {
		t.Errorf("expected choice content to be %q, got %q", "test response", resp.
			Choices[0].Message.Content)
	}
}

func TestChatCompletionError(t *testing.T) {
	apiKey := "test-api-key"

	req := ChatCompletionRequest{
		Model: "test-model",
		Messages: []Message{
			{
				Role:    "user",
				Content: "test message",
			},
		},
		MaxTokens:   100,
		Temperature: 0.5,
	}

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %q", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer "+apiKey {
			t.Errorf("expected Authorization header to be %q, got %q", "Bearer "+apiKey,
				r.Header.Get("Authorization"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type header to be %q, got %q",
				"application/json", r.Header.Get("Content-Type"))
		}

		var unmarshaledReq ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&unmarshaledReq); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if !reflect.DeepEqual(req, unmarshaledReq) {
			t.Errorf("marshaled and unmarshaled requests do not match")
		}

		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error": {"message": "test error"}}`)
	}))

	defer httpServer.Close()

	client := NewClient(apiKey, httpServer.URL)
	client.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	_, err := client.ChatCompletion(req)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if err.Error() != "API request failed with status 404: test error" {
		t.Errorf("expected error to be %q, got %q", "API request failed with status 404: test error", err.Error())
	}
}
