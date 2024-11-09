# TODO

1. Add configuration file for:
   * providing the option for storing chat results in a DB (1.5)
   * storing model and system prompt information

1. How would this app be tested?

1. Add --compare flag to use mulitple models and compare the results......
   * Consider something similar to the Output_Stream for the backend model
   itself. This would allow the --compare flag mentioned below to be
   implemented.

1. Add --system-prompt flag to allow the use of a system prompt

1. Add model flags like `--chatgpt`, `--sonnet`, etc instead of having to use
  `--model chatgpt`

## Version 1.5
1. Add support for a DB backend for storing chat logs
```sql
CREATE TABLE interactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
  prompt TEXT NOT NULL,
  response TEXT NOT NULL,
  model_name TEXT, -- Store which LLM was used (e.g., 'Gemini 1.5 Pro')
  -- Add other fields as needed (e.g., user_id, session_id, etc.)
);
```
1. Change structure of log file to better support Anthropic Role/Content style
   of context

## Version 2.0:
1. Add a TUI interface (using charm?) so the output looks good
1. Add support for --image and --file attachments for multi-modal models
