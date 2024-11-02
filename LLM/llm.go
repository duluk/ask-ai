package LLM

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func get_client_key(llm string) string {
	key_upper := strings.ToUpper(llm) + "_API_KEY"
	key_lower := strings.ToLower(llm) + "-api-key"

	key := os.Getenv(key_upper)
	// TODO this should attempt XDG_CONFIG_HOME first, then HOME
	home := os.Getenv("HOME")
	if key == "" {
		file, err := os.Open(home + "/.config/ask-ai/" + key_lower)
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
