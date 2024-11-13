# TODO

1. How would this app be tested?

1. Add --compare flag to use mulitple models and compare the results......
   * Consider something similar to the Output_Stream for the backend model
   itself. This would allow the --compare flag mentioned below to be
   implemented.

1. Add model flags like `--chatgpt`, `--sonnet`, etc instead of having to use
  `--model chatgpt`

1. Add a --continue flag to allow the user to continue a conversation with the
   same context (and model?). The issue is how to 'mark' a conversational 'turn'
   as the same conversation. Or how to mark the end of a conversation. Maybe
   something like, if the current prompt does not use `--continue`, then it's a
   new conversation and the previous one is over.

## Version 1.5
1. Add support for a DB backend for storing chat logs
```sql
CREATE TABLE interactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
  prompt TEXT NOT NULL,
  response TEXT NOT NULL,
  model_name TEXT, -- Store which LLM was used (e.g., 'Gemini 1.5 Pro')
);
```
1. Use a package like `go-tf-idf` or `rake-go` for doing keyword analysis on
   the conversation to add to the database entry so they can be searched later.

## Version 2.0:
1. Add a TUI interface (using charm?) so the output looks good
1. Add support for --image and --file attachments for multi-modal models
