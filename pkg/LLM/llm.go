package LLM

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
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

func getClientKey(llm string) string {
	keyUpper := strings.ToUpper(llm) + "_API_KEY"
	keyLower := strings.ToLower(llm) + "-api-key"

	key := os.Getenv(keyUpper)
	// TODO this should attempt XDG_CONFIG_HOME first, then HOME
	home := os.Getenv("HOME")
	if key == "" {
		file, err := os.Open(home + "/.config/ask-ai/" + keyLower)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			key = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	}
	return key
}
