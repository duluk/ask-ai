// Package apikey resolves API-key values from configuration.
package apikey

import (
	"fmt"
	"os/exec"
	"strings"
)

// Resolve returns value directly unless it contains whitespace, in which case
// it runs value as a shell command and returns its trimmed standard output.
// Configuration files that use commands must therefore be trusted.
func Resolve(value string) (string, error) {
	if !strings.ContainsAny(value, " \t") {
		return value, nil
	}
	out, err := exec.Command("sh", "-c", value).CombinedOutput()
	if err != nil {
		if detail := strings.TrimSpace(string(out)); detail != "" {
			return "", fmt.Errorf("%w: %s", err, detail)
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
