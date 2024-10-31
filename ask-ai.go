package main

// TODO:
// - Add a flag to specify the model to use
// - Add a flag to specify the chat log file
// - Read the chat log for context possibilities
//   - That is, could add a flag to read the last n messages for context
// - Create an output class/struct or something that can receive different
//   'stream' objects so that one output functoin can be called, then it will
//   send the output to all attached streams. (eg, stdout, log file, etc)

import (
	"bufio"
	"fmt"
	"os"

	"github.com/duluk/ask-ai/LLM"
)

func main() {
	key := os.Getenv("OPENAI_API_KEY")
	home := os.Getenv("HOME")
	if key == "" {
		file, err := os.Open(home + "/.config/ask-ai/openai-api-key")
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

	log, err := os.OpenFile(home+"/.config/ask-ai/ask-ai.chat.log", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening/creating chat log file: ", err)
		fmt.Println("CHAT WILL NOT BE SAVED (but we're forging on)")
	}

	var prompt string
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	} else {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		prompt, _ = reader.ReadString('\n')
		fmt.Println()
	}
	log.WriteString("User: " + prompt + "\n\n")

	client := LLM.New_Client(key)
	err = LLM.Chat(client, prompt, log)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
