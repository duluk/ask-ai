package apikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	key, err := Resolve("literal-key")
	require.NoError(t, err)
	assert.Equal(t, "literal-key", key)

	key, err = Resolve("printf 'command-key\\n'")
	require.NoError(t, err)
	assert.Equal(t, "command-key", key)
}

func TestResolveIncludesCommandErrorOutput(t *testing.T) {
	_, err := Resolve("echo command failed >&2; exit 1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
}
