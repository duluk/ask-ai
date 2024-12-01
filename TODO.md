# TODO

1. Add model flags like `--chatgpt`, `--sonnet`, etc instead of having to use
  `--model chatgpt`
1. Add a --list-models flag to list all the models available. That's not
   difficult, but I'm wondering if I can just pass one of those values in to
   the Model field for that company's API? The flag would need to know which
   API will be used - which company's models to list. (it should default to the
   specified model, which may be the default)

## Version Sooner...
1. Add a TUI interface (using bubbletea?) so the output looks good
   * Markdown support would be nice
1. Add a --list-conversations flag to list all the conversations available.
1. Assuming the following is remotely correct:
```
### **2. How `max_tokens` Relates to Prompt and Context Size**

- **Setting `max_tokens`:** This parameter specifically limits the length of the
generated response. However, the effective maximum you can set for `max_tokens`
depends on the length of your prompt and any previous context included in the conversation.

- **Total Token Calculation:** The sum of tokens in the prompt and the `max_tokens`
setting must not exceed the model's context window. Exceeding this limit will result
in errors or truncated outputs.

 **Example:**
 - **Model Context Window:** 4096 tokens
 - **Prompt Length:** 1500 tokens
 - **Available for Response (`max_tokens`):** 4096 - 1500 = 2596 tokens

### **3. Practical Implications**

- **Optimizing Prompts:** Keep your prompts concise to maximize the potential length
of responses. Redundant or unnecessary information in the prompt consumes valuable
tokens that could otherwise be used for generating more detailed responses.

- **Managing Conversations:** In multi-turn conversations, earlier parts of the
dialogue consume tokens. As the conversation progresses, the available tokens for
new responses decrease unless you implement strategies like summarizing past exchanges.
```
  Then I need to manage the tokens better. I could rename max_tokens in the
  config to response tokens, then estimate the number of tokens used in the
  prompt (with context), and add that to the response_tokens setting, to
  attempt to allow the intended number of tokens in the response. If I'm able
  to gather stats from the models' response, I can use the real number of
  tokens used in the context and estimate just the prompt.

  Another option, or additional, is the note below about smarter context
  management, summarizing context as is reasonable (eg not code). That will
  also reduce the context size. I need to see how other tools handle this.

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
1. Maybe allow commands (/ commands?) in the stdin prompt so that the user
   could do something like switch the model 'on the fly' in the middle of a
   conversation.
1. Add a --search option that sends the query to Google or DuckDuckGo, parses
   the results and takes the top 2 or 3 links, fetches the content of those,
   then sends that to the LLM to generate a response. This would be a way to
   more current information.
   - Or the option could be --web, with --search being for searching the DB

## Other Support Apps
1. Add an ask-ai-db app (or do something like sub-commnds with ask-ai) that
   allows the user to search the database of conversations. This could be
   useful for finding a conversation that was had before, or for finding
   conversations that are similar to the one the user is having now.
   - When the TUI is added, this could be a pane in the TUI
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
