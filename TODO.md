# TODO

1. How would this app be tested?

1. Add model flags like `--chatgpt`, `--sonnet`, etc instead of having to use
  `--model chatgpt`

1. Add a --list-models flag to list all the models available. That's not
   difficult, but I'm wondering if I can just pass one of those values in to
   the Model field for that company's API? The flag would need to know which
   API will be used - which company's models to list.

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
   * Markdown support would be nice
1. Add support for --image and --file attachments for multi-modal models
1. Add --compare flag to use mulitple models and compare the results......
   * Consider something similar to the Output_Stream for the backend model
   itself. This would allow the --compare flag mentioned below to be
   implemented.
   * This is certainly a goroutine concurrent thing
   * Also moved to v2.0 because I think the TUI could provide some panes to
   make it look better and show on screen at the same time (for at least 3
   models; more may be too much)

