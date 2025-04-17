package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	// "golang.org/x/term"
	"github.com/charmbracelet/x/term"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/duluk/ask-ai/pkg/database"
)

// Provider holds configuration for an AI provider
type Provider struct {
	APIKey string                 `mapstructure:"api_key"`
	Models map[string]ModelConfig `mapstructure:"models"`
}

// ModelConfig holds configuration for a specific model
type ModelConfig struct {
	Aliases     []string `mapstructure:"aliases"`
	ModelName   string   `mapstructure:"model_name"`
	Temperature float64  `mapstructure:"temperature"`
	MaxTokens   int      `mapstructure:"max_tokens"`
}

// Config holds the main configuration
type Config struct {
	Models   map[string]Provider `mapstructure:"models"`
	Defaults struct {
		Model    string `mapstructure:"model"`
		Provider string `mapstructure:"provider"`
	} `mapstructure:"defaults"`
	Logging struct {
		File       string `mapstructure:"file"`
		Level      string `mapstructure:"level"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
	} `mapstructure:"logging"`
	Database struct {
		Path  string `mapstructure:"path"`
		Table string `mapstructure:"table"`
	} `mapstructure:"database"`
}

// Add this to your Options struct
// type Options struct {
// 	Model           string
// 	ContextLength   int
// 	ContinueChat    bool
// 	DumpConfig      bool
// 	LogFileName     string
// 	DBFileName      string
// 	DBTable         string
// 	SystemPrompt    string
// 	MaxTokens       int
// 	Temperature     float32
// 	ConversationID  int
// 	ScreenWidth     int
// 	ScreenTextWidth int
// 	ScreenHeight    int
// 	TabWidth        int
// 	Quiet           bool
// 	NoRecord        bool
// 	UseTUI          bool // Add this field
// 	NoOutput        bool
// }

// Options holds runtime configuration options
type Options struct {
	ConfigDir      string
	DumpConfig     bool
	ShowAPIKeys    bool
	Model          string
	Provider       string
	Temperature    float64
	MaxTokens      int
	LogFileName    string
	DBFileName     string
	DBTable        string
	SystemPrompt   string
	UseTUI         bool
	NoOutput       bool
	ContinueChat   bool
	ConversationID int
}

const Version = "0.3.3"

const (
	MaxTermTextWidth = 80
	widthPad         = 5
	TabWidth         = 4
)

var (
	commit = "Unknown"
	date   = "Unknown"
)

func Initialize() (*Options, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ask-ai")

	// TODO: though I've put so much effort into the config file to read it
	// first so that the values can be used as defaults (eg in --help), I'm
	// starting to wonder if that's even what I want...
	// Figure out the config file and read it first if there
	pflag.StringP("config", "C", "", "Configuration file")
	viper.BindPFlags(pflag.CommandLine)
	if err := setupConfigFile(); err != nil {
		return nil, fmt.Errorf("error setting up config: %w", err)
	}

	width, height := determineScreenSize()

	pflag.StringP("model", "m", "", "Model to use")
	pflag.StringP("provider", "p", "", "Provider to use")
	pflag.Float64P("temperature", "t", 0.7, "Temperature for model responses")
	pflag.IntP("max-tokens", "M", 1024, "Maximum tokens for response")
	pflag.BoolP("continue", "c", false, "Continue last conversation")
	pflag.IntP("id", "i", 0, "Conversation ID to continue")
	pflag.BoolP("tui", "T", false, "Use TUI interface")
	pflag.BoolP("no-output", "n", false, "Disable direct terminal output")
	pflag.BoolP("dump-config", "d", false, "Dump configuration and exit")
	pflag.BoolP("show-keys", "k", false, "Show API keys in config dump")

	viper.BindPFlags(pflag.CommandLine)

	// Set default configuration values
	viper.SetDefault("defaults.provider", "openai")
	viper.SetDefault("defaults.model", "chatgpt-4o-latest")
	viper.SetDefault("logging.file", filepath.Join(configDir, "ask-ai.log"))
	viper.SetDefault("database.path", filepath.Join(configDir, "ask-ai.db"))
	viper.SetDefault("database.table", "conversations")

	// Read config file
	if configFile := viper.GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	// Create options from config and flags
	opts := &Options{
		ConfigDir:      configDir,
		DumpConfig:     viper.GetBool("dump-config"),
		ShowAPIKeys:    viper.GetBool("show-keys"),
		Model:          viper.GetString("model"),
		Provider:       viper.GetString("provider"),
		Temperature:    viper.GetFloat64("temperature"),
		MaxTokens:      viper.GetInt("max-tokens"),
		LogFileName:    os.ExpandEnv(config.Logging.File),
		DBFileName:     os.ExpandEnv(config.Database.Path),
		DBTable:        config.Database.Table,
		UseTUI:         viper.GetBool("tui"),
		NoOutput:       viper.GetBool("no-output"),
		ContinueChat:   viper.GetBool("continue"),
		ConversationID: viper.GetInt("id"),
	}

	// If model not specified on command line, use default from config
	if opts.Model == "" {
		opts.Model = config.Defaults.Model
	}
	if opts.Provider == "" {
		opts.Provider = config.Defaults.Provider
	}

	return opts, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetModelConfig returns the configuration for a specific model
func GetModelConfig(config *Config, provider, model string) (*ModelConfig, error) {
	p, ok := config.Models[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", provider)
	}

	// Check direct model name
	if m, ok := p.Models[model]; ok {
		return &m, nil
	}

	// Check aliases
	for _, m := range p.Models {
		for _, alias := range m.Aliases {
			if alias == model {
				return &m, nil
			}
		}
	}

	return nil, fmt.Errorf("model %s not found for provider %s", model, provider)
}

// Maybe this shouldn't be in config...
func searchForConversation(search string) {
	if viper.GetString("database.file") == "" {
		fmt.Println("Database file not set")
		os.Exit(1)
	}
	if viper.GetString("database.table") == "" {
		fmt.Println("Database table not set")
		os.Exit(1)
	}

	db, err := database.InitializeDB(os.ExpandEnv(viper.GetString("database.file")), viper.GetString("database.table"))
	if err != nil {
		fmt.Printf("Error opening database: %s", err)
		os.Exit(1)
	}
	defer db.Close()

	ids, err := db.SearchForConversation(search)
	if err != nil {
		fmt.Printf("Error searching for conversation: %s", err)
	}

	uniqIDs := make([]int, 0)
	unique := make(map[int]bool)
	for _, id := range ids {
		if !unique[id] {
			unique[id] = true
			uniqIDs = append(uniqIDs, id)
		}
	}

	fmt.Printf("Found %d conversations: ", len(uniqIDs))
	for i, id := range uniqIDs {
		fmt.Printf("%d", id)
		if i < len(uniqIDs)-1 {
			fmt.Printf(", ")
		}
	}
	fmt.Println()

	os.Exit(0)
}

func determineScreenSize() (int, int) {
	width, height, err := term.GetSize(0)
	if err != nil {
		return 80, 24
	}

	return width, height
}

func showConversation(convID int) {
	if viper.GetString("database.file") == "" {
		fmt.Println("Database file not set")
		os.Exit(1)
	}
	if viper.GetString("database.table") == "" {
		fmt.Println("Database table not set")
		os.Exit(1)
	}

	db, err := database.InitializeDB(os.ExpandEnv(viper.GetString("database.file")), viper.GetString("database.table"))
	if err != nil {
		fmt.Printf("Error opening database: %s", err)
		os.Exit(1)
	}
	defer db.Close()

	db.ShowConversation(convID)
	os.Exit(0)
}

func handleVersionFlags() bool {
	if viper.GetBool("version") {
		fmt.Println("ask-ai version:", Version)
		return true
	}
	if viper.GetBool("full-version") {
		fmt.Printf("Version: %s\nCommit:  %s\nDate:    %s\n", Version, commit, date)
		return true
	}
	return false
}

// This is pretty bad but it's a quick hack to get the config file because
// nothing else is working. I need to parse and read the config file before
// the rest of the options so that I can use the values from the config file
// to set the defaults for the other options.
func checkConfigFlag() string {
	for i, arg := range os.Args {
		if arg == "--config" || arg == "-C" {
			if i+1 < len(os.Args) {
				return os.Args[i+1]
			}
		}
		if len(arg) > 8 && arg[:8] == "--config=" {
			return arg[8:]
		}
	}
	return ""
}

// If there's an error getting the user, just returning the path unmodified
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~") {
		currentUser, err := user.Current()
		if err != nil {
			// I'm always trying to sneak a goto in just to trigger :)
			goto oopsies
		}
		return filepath.Join(currentUser.HomeDir, path[1:])
	}
oopsies:
	return path
}

func setupConfigFile() error {
	cfgFile := checkConfigFlag()

	if cfgFile != "" {
		viper.SetConfigFile(expandHomePath(cfgFile))
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
	fmt.Printf("ContextLength: %d\n", cfg.ContextLength)
	fmt.Printf("ContinueChat: %t\n", cfg.ContinueChat)
	fmt.Printf("LogFileName: %s\n", cfg.LogFileName)
	fmt.Printf("DBFileName: %s\n", cfg.DBFileName)
	fmt.Printf("DBTable: %s\n", cfg.DBTable)
	fmt.Printf("SystemPrompt: %s\n", cfg.SystemPrompt)
	fmt.Printf("MaxTokens: %d\n", cfg.MaxTokens)
	fmt.Printf("Temperature: %f\n", cfg.Temperature)
	fmt.Printf("ConversationID: %d\n", cfg.ConversationID)
	fmt.Printf("ScreenWidth: %d\n", cfg.ScreenWidth)
	fmt.Printf("ScreenHeight: %d\n", cfg.ScreenHeight)
	fmt.Printf("TabWidth: %d\n", cfg.TabWidth)
}
