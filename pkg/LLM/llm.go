package LLM

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/viper"
)

// Gemini created this function, along with tokenizeWord. It's not perfect by
// any means but it provided a decent estimate, compared to what the LLMs
// returned for the same prompt.
func EstimateTokens(text string) int32 {
	var tokenCount int32
	words := strings.Fields(text)

	for _, word := range words {
		tokenCount += tokenizeWord(word)
	}

	return tokenCount
}

func tokenizeWord(word string) int32 {
	var tokens int32
	var currentToken string

	for _, char := range word {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			currentToken += string(char)
		} else if unicode.IsPunct(char) {
			if currentToken != "" {
				tokens++
				currentToken = ""
			}
			tokens++
		} else if unicode.IsSpace(char) {
			if currentToken != "" {
				tokens++
				currentToken = ""
			}
		} else {
			if currentToken != "" {
				tokens++
				currentToken = ""
			}
			tokens++
		}
	}

	if currentToken != "" {
		tokens++
	}

	return tokens
}

// TODO: what should the precedence be? Which should be used first, env or config?
func getClientKey(llm string) string {
	// 1) First try environment variable
	keyEnv := strings.ToUpper(llm) + "_API_KEY"
	if key := os.Getenv(keyEnv); key != "" {
		return key
	}

	// 2) Try API key from config file via viper
	cfgKey := fmt.Sprintf("models.%s.api_key", llm)
	if key := viper.GetString(cfgKey); key != "" {
		// If the api_key contains whitespace, treat it as a shell command to run
		if strings.ContainsAny(key, " \t") {
			out, err := exec.Command("sh", "-c", key).Output()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error executing API key command for %s: %v\n", llm, err)
				os.Exit(1)
			}
			return strings.TrimSpace(string(out))
		}
		return key
	}

	// 3) Fallback: read key from file under XDG or HOME config dir
	cfgHome := os.Getenv("XDG_CONFIG_HOME")
	if cfgHome == "" {
		cfgHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	keyFile := strings.ToLower(llm) + "-api-key"
	filePath := filepath.Join(cfgHome, "ask-ai", keyFile)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading API key for %s: %v\n", llm, err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading API key for %s: %v\n", llm, err)
		os.Exit(1)
	}
	return ""
}
