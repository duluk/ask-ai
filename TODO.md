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
  temperature REAL, -- Store the temperature used for the response
);
```
1. Use a package like `go-tf-idf` or `rake-go` for doing keyword analysis on
   the conversation to add to the database entry so they can be searched later.

## Version 2.0:
1. Add a TUI interface (using bubbletea?) so the output looks good
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
1. Track conversations by ID in the DB. (or even the YML file) This would allow
   different conversations to be used for continuation later. This would mean
   another flag to list-conversations, with ID. This would mean having the LLM
   (or some library) summarize the conversation for the one-line summary output.
1. Add a --list-conversations flag to list all the conversations available.
1. Add a --server option to run a web server where the conversations can be
   searched. Very simple. Though it could turn into something like a web-based
   LLM chatbot, where conversations from the Web version and CLI version are
   stored in same place.

## Structure Idea
1. Should I create multiple individual commands that do something specific,
   like ai-query and ai-review and ai-compare and ai-cmd (etc) so that the
   commands can be piped together in traditional *NIX fashion? Versus one
   command that does everything?
