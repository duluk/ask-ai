package websearch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBraveSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.Header.Get("X-Subscription-Token"))
		assert.Equal(t, "latest Go release", r.URL.Query().Get("q"))
		assert.Equal(t, "3", r.URL.Query().Get("count"))
		assert.Equal(t, "moderate", r.URL.Query().Get("safesearch"))
		_, _ = w.Write([]byte(`{"web":{"results":[{"title":"Go","url":"https://go.dev","description":"The Go programming language."}]}}`))
	}))
	defer server.Close()

	client := NewBraveClient("test-key")
	client.endpoint = server.URL
	results, err := client.Search(context.Background(), "latest Go release", 3)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, Result{Title: "Go", URL: "https://go.dev", Description: "The Go programming language."}, results[0])
}

func TestFormatResults(t *testing.T) {
	formatted := FormatResults("latest Go release", []Result{{Title: "Go", URL: "https://go.dev", Description: "Current release notes"}})
	assert.Contains(t, formatted, "untrusted reference text")
	assert.Contains(t, formatted, "[1] Go")
	assert.Contains(t, formatted, "URL: https://go.dev")
}
