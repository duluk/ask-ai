package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/duluk/ask-ai/pkg/LLM"
)

type plannerTestClient struct {
	response string
}

func (c plannerTestClient) Chat(LLM.ClientArgs, int, int) (LLM.ClientResponse, <-chan LLM.StreamResponse, error) {
	stream := make(chan LLM.StreamResponse, 1)
	stream <- LLM.StreamResponse{Content: c.response}
	close(stream)
	return LLM.ClientResponse{}, stream, nil
}

func (c plannerTestClient) ChatStream(LLM.ClientArgs, int, int, chan<- LLM.StreamResponse) (LLM.ClientResponse, error) {
	return LLM.ClientResponse{}, nil
}

func TestChooseWebSearchQuery(t *testing.T) {
	prompt := "What is the current Go release?"
	systemPrompt := "You are helpful."
	maxTokens := 512
	temperature := float32(0.7)
	args := LLM.ClientArgs{
		Prompt:       &prompt,
		SystemPrompt: &systemPrompt,
		MaxTokens:    &maxTokens,
		Temperature:  &temperature,
	}

	query, err := chooseWebSearchQuery(plannerTestClient{response: `{"search":true,"query":"latest Go release"}`}, args)
	require.NoError(t, err)
	assert.Equal(t, "latest Go release", query)

	query, err = chooseWebSearchQuery(plannerTestClient{response: `{"search":false}`}, args)
	require.NoError(t, err)
	assert.Empty(t, query)
}

func TestChooseWebSearchQueryRejectsInvalidResponse(t *testing.T) {
	prompt := "Hello"
	systemPrompt := "You are helpful."
	maxTokens := 512
	temperature := float32(0.7)
	args := LLM.ClientArgs{Prompt: &prompt, SystemPrompt: &systemPrompt, MaxTokens: &maxTokens, Temperature: &temperature}

	_, err := chooseWebSearchQuery(plannerTestClient{response: "search for it"}, args)
	require.Error(t, err)
}
