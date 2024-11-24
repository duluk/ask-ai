package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Options struct {
	Model         string
	Context       int
	ContextLength int
	ContinueChat  bool
	LogFileName   string
	DBFileName    string
	SystemPrompt  string
	MaxTokens     int
	Temperature   float32
}

const version = "0.3.0"

var (
	commit = "Unknown"
	date   = "Unknown"
)

func Initialize() (*Options, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ask-ai")

	// Set baseline defaults
	viper.SetDefault("model.default", "claude")
	viper.SetDefault("model.context_length", 2048)
	viper.SetDefault("model.max_tokens", 512)
	viper.SetDefault("model.temperature", 0.7)
	viper.SetDefault("model.system_prompt", "")
	viper.SetDefault("log.file", filepath.Join(configDir, "ask-ai.chat.yml"))
	viper.SetDefault("database.file", filepath.Join(configDir, "ask-ai.db"))

	// Handle config option separately so we can use it for defaults
	configFlags := pflag.NewFlagSet("config", pflag.ContinueOnError)
	configFlags.StringP("config", "C", "", "Configuration file")
	configFlags.ParseErrorsWhitelist.UnknownFlags = true
	err := configFlags.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("error parsing config flags: %v", err)
	}
	viper.BindPFlag("config", configFlags.Lookup("config"))

	// Set up and read config file
	if err := setupConfigFile(); err != nil {
		return nil, fmt.Errorf("error setting up config: %w", err)
	}

	args := removeConfigFlag(os.Args[1:])

	// Now define the rest of the flags using values from viper (which now has config file values)
	pflag.StringP("model", "m", viper.GetString("model.default"), "Which LLM to use (claude|chatgpt|gemini|grok)")
	pflag.IntP("context", "n", 0, "Use previous n messages for context")
	pflag.IntP("context-length", "l", viper.GetInt("model.context_length"), "Maximum context length")
	pflag.BoolP("continue", "c", false, "Continue previous conversation")
	pflag.StringP("log", "L", viper.GetString("log.file"), "Chat log file")
	pflag.StringP("database", "d", viper.GetString("database.file"), "Database file")
	pflag.StringP("system-prompt", "s", viper.GetString("model.system_prompt"), "System prompt for LLM")
	pflag.IntP("max-tokens", "t", viper.GetInt("model.max_tokens"), "Maximum tokens to generate")
	pflag.Float32P("temperature", "T", float32(viper.GetFloat64("model.temperature")), "Temperature for generation")
	pflag.BoolP("version", "v", false, "Print version and exit")
	pflag.BoolP("full-version", "V", false, "Print full version information and exit")

	// Parse all flags again
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.CommandLine.Parse(args)

	// Handle version flags early
	viper.BindPFlag("version", pflag.Lookup("version"))
	viper.BindPFlag("full-version", pflag.Lookup("full-version"))
	if handleVersionFlags() {
		os.Exit(0)
	}

	// Bind all flags to viper
	viper.BindPFlag("context", pflag.Lookup("context"))
	viper.BindPFlag("continue", pflag.Lookup("continue"))
	viper.BindPFlag("log.file", pflag.Lookup("log"))
	viper.BindPFlag("database.file", pflag.Lookup("database"))
	viper.BindPFlag("model.system_prompt", pflag.Lookup("system-prompt"))
	viper.BindPFlag("model.max_tokens", pflag.Lookup("max-tokens"))
	viper.BindPFlag("model.context_length", pflag.Lookup("context-length"))
	viper.BindPFlag("model.temperature", pflag.Lookup("temperature"))

	// Create and return options
	return &Options{
		Model:         pflag.Lookup("model").Value.String(),
		Context:       viper.GetInt("context"),
		ContextLength: viper.GetInt("model.context_length"),
		ContinueChat:  viper.GetBool("continue"),
		LogFileName:   os.ExpandEnv(viper.GetString("log.file")),
		DBFileName:    os.ExpandEnv(viper.GetString("database.file")),
		SystemPrompt:  viper.GetString("model.system_prompt"),
		MaxTokens:     viper.GetInt("model.max_tokens"),
		Temperature:   float32(viper.GetFloat64("model.temperature")),
	}, nil
}

func handleVersionFlags() bool {
	if viper.GetBool("version") {
		fmt.Println("ask-ai version:", version)
		return true
	}
	if viper.GetBool("full-version") {
		fmt.Printf("Version: %s\nCommit:  %s\nDate:    %s\n", version, commit, date)
		return true
	}
	return false
}

func removeConfigFlag(args []string) []string {
	result := make([]string, 0, len(args))
	skip := false
	for _, arg := range args {
		if skip {
			skip = false
			continue
		}
		if arg == "--config" || arg == "-c" {
			skip = true
			continue
		}
		if strings.HasPrefix(arg, "--config=") {
			continue
		}
		result = append(result, arg)
	}
	return result
}

func setupConfigFile() error {
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
		viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "ask-ai"))
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}
	return nil
}

func DumpConfig(cfg *Options) {
	fmt.Printf("Model: %s\n", cfg.Model)
	fmt.Printf("Context: %d\n", cfg.Context)
	fmt.Printf("ContextLength: %d\n", cfg.ContextLength)
	fmt.Printf("ContinueChat: %t\n", cfg.ContinueChat)
	fmt.Printf("LogFileName: %s\n", cfg.LogFileName)
	fmt.Printf("DBFileName: %s\n", cfg.DBFileName)
	fmt.Printf("SystemPrompt: %s\n", cfg.SystemPrompt)
	fmt.Printf("MaxTokens: %d\n", cfg.MaxTokens)
	fmt.Printf("Temperature: %f\n", cfg.Temperature)
}
