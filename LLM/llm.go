package LLM

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Output_Stream struct {
	writers map[string]io.Writer
}

func New_Output_Stream() *Output_Stream {
	return &Output_Stream{writers: make(map[string]io.Writer)}
}

func (ostr *Output_Stream) Add_Stream(name string, writer io.Writer) {
	ostr.writers[name] = writer
}

func (ostr *Output_Stream) Remove_Stream(name string) {
	delete(ostr.writers, name)
}

func (ostr *Output_Stream) Write(p []byte) (n int, err error) {
	for _, writer := range ostr.writers {
		n, err = writer.Write(p)
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (ostr *Output_Stream) Printf(format string, a ...interface{}) (n int, err error) {
	fmtstr := fmt.Sprintf(format, a...)

	for _, writer := range ostr.writers {
		n, err = writer.Write([]byte(fmtstr))
		if err != nil {
			return n, err
		}
	}

	return len(fmtstr), nil
}

func (ostr *Output_Stream) Nl() (n int, err error) {
	return fmt.Fprintln(ostr)
}

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
