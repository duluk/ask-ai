# TODO

1. Add model flags like `--chatgpt`, `--sonnet`, etc instead of having to use
  `--model chatgpt`
1. Add a --list-models flag to list all the models available. That's not
   difficult, but I'm wondering if I can just pass one of those values in to
   the Model field for that company's API? The flag would need to know which
   API will be used - which company's models to list. (it should default to the
   specified model, which may be the default)
1. Manage stats in the conversation, such as tokens used, tokens received, etc.

## Version Sooner...
1. When using the prompt (not passing in the prompt on the CLI), don't return
   after the first answer but return to the prompt for more, continuing the
   conversation.
1. Add a TUI interface (using bubbletea?) so the output looks good
   * Markdown support would be nice
1. Track conversations by ID in the DB. (or even the YML file) This would allow
   different conversations to be used for continuation later. This would mean
   another flag to list-conversations, with ID. This would mean having the LLM
   (or some library) summarize the conversation for the one-line summary output.
1. Add a --list-conversations flag to list all the conversations available.

## Version Later...
1. Add support for --image and --file attachments for multi-modal models
1. Add --compare flag to use mulitple models and compare the results......
   * Consider something similar to the Output_Stream for the backend model
   itself. This would allow the --compare flag mentioned below to be
   implemented.
   * This is certainly a goroutine concurrent thing
   * Also moved to v2.0 because I think the TUI could provide some panes to
   make it look better and show on screen at the same time (for at least 3
   models; more may be too much)
1. Add a --server option to run a web server where the conversations can be
   searched. Very simple. Though it could turn into something like a web-based
   LLM chatbot, where conversations from the Web version and CLI version are
   stored in same place.
1. Use a package like `go-tf-idf` or `rake-go` for doing keyword analysis on
   the conversation to add to the database entry so they can be searched later.

## Structure Idea
1. Should I create multiple individual commands that do something specific,
   like ai-query and ai-review and ai-compare and ai-cmd (etc) so that the
   commands can be piped together in traditional *NIX fashion? Versus one
   command that does everything?
1. Smarter management of context length/conversation window. Instead of dumping
   the entire conversation into the prompt, consider summarizing the
   conversation as best is possible. (this is more difficult for coding I would
   assume) There is also the option of using the current prompt, then looking
   at the conversation in the database and extracting the relevant context as
   needed.
   - Some options:
      - go-summarize or go-nlp (similar to Genism or Sumy)
      - Using an LLM itself behind the scenes
      - Using a keyword extraction library like go-rake or go-tf-idf
      - Vector databases? Go: milvus-sdk-go, for milvus open source vector db
      - SQLite FTS5 - for an sqlite database
      - knowledge-graph - use a knoweldge graph to store the conversation for
        capturing relationships and concepts
        - Go: google/knowledge-graph
   - Note: for coding conversations, this can cause issues as mentioned above.
   Code conversations require retaining some detail. But there could be some
   options like keeping only the last couple of conversational turns in tact
   and summarizing what came before. Or keep the error messages/code snippets
   in tact (eg using a vector db).
