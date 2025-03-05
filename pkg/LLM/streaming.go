package LLM

import (
	"context"
)

// This file contains default implementations for StreamChat to make it easier
// to implement the Client interface for providers that don't have native streaming support

// DefaultStreamChat provides a fallback implementation of StreamChat for LLM providers
// that don't natively support streaming. It simply calls Chat and returns the entire
// response as a single chunk.
func DefaultStreamChat(client Client, args ClientArgs, termWidth int, tabWidth int, callback func(chunk string)) (ClientResponse, error) {
	// Create a goroutine that can be cancelled
	type result struct {
		resp ClientResponse
		err  error
	}

	// Use the context from args if provided, otherwise use background context
	ctx := context.Background()
	if args.Ctx != nil {
		ctx = args.Ctx
	}

	// Create a channel for results
	resChan := make(chan result, 1)

	// Call the standard Chat method in a separate goroutine
	go func() {
		resp, err := client.Chat(args, termWidth, tabWidth)
		select {
		case resChan <- result{resp, err}:
			// Result sent successfully
		case <-ctx.Done():
			// Context was cancelled, no need to send result
		}
	}()

	// Wait for either the result or context cancellation
	select {
	case res := <-resChan:
		if res.err != nil {
			return ClientResponse{}, res.err
		}
		// Send the full response as a single chunk
		callback(res.resp.Text)
		// Return the same response
		return res.resp, nil
	case <-ctx.Done():
		// Context was cancelled
		return ClientResponse{}, ctx.Err()
	}
}
