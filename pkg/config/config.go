package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	// "golang.org/x/term"
	"github.com/charmbracelet/x/term"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/duluk/ask-ai/pkg/database"
)

// Provider holds configuration for an AI provider
// The Models field captures all model entries under the provider block.
type Provider struct {
	APIKey string                 `mapstructure:"api_key"`
	Models map[string]ModelConfig `mapstructure:",remain"`
}

// ModelConfig holds configuration for a specific model
type ModelConfig struct {
	Aliases     []string `mapstructure:"aliases"`
	ModelName   string   `mapstructure:"model_name"`
	Temperature float64  `mapstructure:"temperature"`
	MaxTokens   int      `mapstructure:"max_tokens"`
	Thinking    string   `mapstructure:"thinking"`
}

// RoleConfig holds configuration for a specific role
// Prompt may be a single string or an array of strings; merged into []string
type RoleConfig struct {
	Description string `mapstructure:"description"`
	Model       string `mapstructure:"model"`
	Prompt      []string
}

// Config holds the main configuration
// It is primarily used for reading provider/model settings; logging and database
// options are read directly via viper for Options initialization.
type Config struct {
	Roles    map[string]RoleConfig
	Models   map[string]Provider `mapstructure:"models"`
	Defaults struct {
		Model    string `mapstructure:"model"`
		Provider string `mapstructure:"provider"`
		Role     string `mapstructure:"role"`
		// other defaults (e.g., max_tokens, context_length, system_prompt) are
		// read via viper directly
	} `mapstructure:"defaults"`
	// Logging block (new config uses 'log')
	Logging struct {
		File       string `mapstructure:"file"`
		Level      string `mapstructure:"level"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
	} `mapstructure:"log"`
	// Database block
	Database struct {
		File  string `mapstructure:"file"`
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
	Thinking       string
	Temperature    float32 // model temperature
	MaxTokens      int     // max tokens for response
	ContextLength  int     // context window length
	LogFileName    string
	DBFileName     string
	DBTable        string
	SystemPrompt   string
	UseTUI         bool
	NoOutput       bool
	NoRecord       bool
	Quiet          bool
	ContinueChat   bool
	ConversationID int

  SearchKeyword     string // Keyword for searching previous conversations
	ListConversations bool   // Flag to list all conversations interactively

  // Terminal dimensions and tab width for TUI or dumping
	ScreenWidth     int // total terminal width
	ScreenTextWidth int // usable text width (terminal width minus pad, capped)
	ScreenHeight    int // total terminal height
	TabWidth        int // tab width for text wrapping
	// Loaded full configuration, including providers and models
	Config *Config
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
	// viper uses a '.' for concatenating keys, so it can't be in a YAML key
	// (eg gemini-2.5-pro). Changing the delimiter to '|' is a workaround.
	// Note: can also just not use the '.' in the YAML key; the model_name
	// string is what's passed to the API.
	// NOTE: this function isn't working so just removing the '.' from the
	// config key
	// viper.KeyDelimiter("|")

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
	// Compute usable text width (terminal width minus pad, capped)
	textWidth := width - widthPad
	if textWidth > MaxTermTextWidth {
		textWidth = MaxTermTextWidth
	}

	// Runtime options flags
	pflag.StringP("model", "m", "", "Model to use")
	pflag.StringP("provider", "p", "", "Provider to use")
	// Temperature controls randomness in responses
	pflag.Float64P("temperature", "t", 0.7, "Temperature for model responses")
	// Maximum tokens for a single response (default 512)
	pflag.IntP("max-tokens", "M", 512, "Maximum tokens for response")
	pflag.StringP("thinking-effort", "e", "medium", "Reasoning effort for model responses")
	pflag.BoolP("continue", "c", false, "Continue last conversation")
	pflag.IntP("id", "i", 0, "Conversation ID to continue")
	pflag.String("search", "", "Search previous conversations for keyword")
	pflag.BoolP("list", "l", false, "List all conversations interactively")
	pflag.BoolP("tui", "T", false, "Use TUI interface")
	pflag.BoolP("no-output", "n", false, "Disable direct terminal output")
	pflag.BoolP("quiet", "q", false, "Suppress non-essential output")
	// Disable conversation recording
	pflag.Bool("no-record", false, "Disable recording conversations to database")
	pflag.BoolP("dump-config", "d", false, "Dump configuration and exit")
	pflag.BoolP("show-keys", "k", false, "Show API keys in config dump")
	// Additional runtime flags
	// Context window length for prompts (default 2048)
	pflag.Int("context-length", 2048, "Context window length for model responses")
	// System prompt override
	pflag.String("system-prompt", "", "System prompt to send to model")
	// Role selection override (use role prompts from config)
	pflag.StringP("role", "r", "", "Role to use for system prompt (as defined in config)")

	// Bind flags to viper and parse CLI
	viper.BindPFlags(pflag.CommandLine)
	// Parse CLI flags to populate values
	pflag.Parse()

	// Set default configuration values
	// Set default configuration values
	viper.SetDefault("defaults.provider", "openai")
	// Default model (fallback when not specified in config or CLI)
	viper.SetDefault("defaults.model", "ollama")
	// Default log file location
	viper.SetDefault("log.file", filepath.Join(configDir, "ask-ai.log"))
	// Default database file and table
	viper.SetDefault("database.file", filepath.Join(configDir, "ask-ai.db"))
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
	// Parse roles section if present
	if rawRoles := viper.Get("roles"); rawRoles != nil {
		if rm, ok := rawRoles.(map[string]any); ok {
			config.Roles = make(map[string]RoleConfig, len(rm))
			for name, entry := range rm {
				em, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				var rc RoleConfig
				if d, ok := em["description"].(string); ok {
					rc.Description = d
				}
				// Optional model override for this role
				if mVal, ok := em["model"].(string); ok {
					rc.Model = mVal
				}
				if p, ok := em["prompt"]; ok {
					switch v := p.(type) {
					case string:
						rc.Prompt = []string{v}
					case []any:
						for _, item := range v {
							if s, ok := item.(string); ok {
								rc.Prompt = append(rc.Prompt, s)
							}
						}
					}
				}
				config.Roles[name] = rc
			}
		}
	}

	// Create options from config and flags
	opts := &Options{}
	// Attach full parsed config for model lookup
	opts.Config = &config

	// Basic flags/booleans
	opts.ConfigDir = configDir
	opts.DumpConfig = viper.GetBool("dump-config")
	opts.ShowAPIKeys = viper.GetBool("show-keys")
	opts.UseTUI = viper.GetBool("tui")
	opts.NoOutput = viper.GetBool("no-output")
	// Respect no-record flag to skip saving conversations
	opts.NoRecord = viper.GetBool("no-record")
	opts.Quiet = viper.GetBool("quiet")
	opts.ContinueChat = viper.GetBool("continue")
	opts.ConversationID = viper.GetInt("id")
	opts.SearchKeyword = viper.GetString("search")
	opts.ListConversations = viper.GetBool("list")
	// Terminal size and tab width
	opts.ScreenWidth = width
	opts.ScreenTextWidth = textWidth
	opts.ScreenHeight = height
	opts.TabWidth = TabWidth

	// Determine model selection: CLI flag > old-style config block > new-style defaults
	modelConf := viper.Sub("model")
	if m := viper.GetString("model"); m != "" {
		opts.Model = m
	} else if modelConf != nil && modelConf.IsSet("default") {
		opts.Model = modelConf.GetString("default")
	} else {
		opts.Model = viper.GetString("defaults.model")
	}
	// Provider selection: CLI flag > new-style default
	if p := viper.GetString("provider"); p != "" {
		opts.Provider = p
	} else {
		opts.Provider = viper.GetString("defaults.provider")
	}

	// MaxTokens: CLI flag > old-style config block > new-style defaults > flag default
	if modelConf != nil && modelConf.IsSet("max_tokens") {
		opts.MaxTokens = modelConf.GetInt("max_tokens")
	} else if mt := viper.GetInt("defaults.max_tokens"); mt != 0 {
		opts.MaxTokens = mt
	} else {
		opts.MaxTokens = viper.GetInt("max-tokens")
	}

	// Thiking Effort: CLI flag > old-style config block > new-style defaults > flag default
	if modelConf != nil && modelConf.IsSet("thinking") {
		opts.Thinking = modelConf.GetString("thinking")
	} else if th := viper.GetString("defaults.thinking"); th != "" {
		opts.Thinking = th
	} else {
		opts.Thinking = viper.GetString("thinking")
	}

	// ContextLength: CLI flag > old-style config block > new-style defaults > flag default
	if modelConf != nil && modelConf.IsSet("context_length") {
		opts.ContextLength = modelConf.GetInt("context_length")
	} else if cl := viper.GetInt("defaults.context_length"); cl != 0 {
		opts.ContextLength = cl
	} else {
		opts.ContextLength = viper.GetInt("context-length")
	}

	// Temperature: CLI flag > old-style config block > new-style defaults > flag default
	if modelConf != nil && modelConf.IsSet("temperature") {
		opts.Temperature = float32(modelConf.GetFloat64("temperature"))
	} else if t := viper.GetFloat64("defaults.temperature"); t != 0 {
		opts.Temperature = float32(t)
	} else {
		opts.Temperature = float32(viper.GetFloat64("temperature"))
	}

	// SystemPrompt: CLI flag > role selection > old-style config block > new-style defaults
	if sp := viper.GetString("system-prompt"); sp != "" {
		opts.SystemPrompt = sp
	} else if roleName := viper.GetString("role"); roleName != "" {
		if rc, ok := config.Roles[roleName]; ok {
			opts.SystemPrompt = strings.Join(rc.Prompt, "\n")
			// Override model if specified for this role and not set via CLI
			if rc.Model != "" && viper.GetString("model") == "" {
				opts.Model = rc.Model
			}
		} else {
			return nil, fmt.Errorf("role %q not found in config", roleName)
		}
	} else if defaultRole := viper.GetString("defaults.role"); defaultRole != "" {
		if rc, ok := config.Roles[defaultRole]; ok {
			opts.SystemPrompt = strings.Join(rc.Prompt, "\n")
			if rc.Model != "" && viper.GetString("model") == "" {
				opts.Model = rc.Model
			}
		} else {
			return nil, fmt.Errorf("default role %q not found in config", defaultRole)
		}
	} else if modelConf != nil && modelConf.IsSet("system_prompt") {
		opts.SystemPrompt = modelConf.GetString("system_prompt")
	} else if sp := viper.GetString("defaults.system_prompt"); sp != "" {
		opts.SystemPrompt = sp
	}

	// Log and database settings
	opts.LogFileName = os.ExpandEnv(viper.GetString("log.file"))
	opts.DBFileName = os.ExpandEnv(viper.GetString("database.file"))
	opts.DBTable = viper.GetString("database.table")

	// Validations
	for _, provider := range config.Models {
		for modelName, modelConfig := range provider.Models {
			if modelConfig.Thinking != "" {
				if err := validateThinking(modelConfig.Thinking); err != nil {
					return nil, fmt.Errorf("model %s: %w", modelName, err)
				}
			}
		}
	}

	return opts, nil
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
		if slices.Contains(m.Aliases, model) {
			return &m, nil
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

func validateThinking(thinking string) error {
	allowedValues := []string{"low", "medium", "high"}
	if slices.Contains(allowedValues, thinking) {
		return nil
	}
	return fmt.Errorf("invalid Thinking value: %s", thinking)
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
		cfg_len := len("--config=")
		if len(arg) > 8 && arg[:cfg_len] == "--config=" {
			return arg[cfg_len:]
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
	// fmt.Printf("checkConfigFlag return: %s\n", cfgFile)

	if cfgFile != "" {
		fmt.Printf("Setting config file: %s\n", cfgFile)
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
