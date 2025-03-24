# ask-ai Project Guidelines

## Build and Run Commands
- Build project: `make` or `go build cmd/ask-ai/main.go`
- Run tests: `make test` (all tests) or `go test ./pkg/[module]` (specific module)
- Run specific test: `go test -run [TestName] ./pkg/[module]`
- Format code: `make fmt` or `go fmt ./...`
- Check code: `make check` or `go vet ./...`
- Install: `make install`
- Run: `bin/ask-ai [args]` or `make run`

## Code Style
- **Naming**: Use CamelCase for public functions, lowerCamelCase for private functions
- **Packages**: Organize code in domain-specific packages under `pkg/`
- **Imports**: Group standard library, external libraries, and project imports with blank line separation
- **Error Handling**: Always check errors and return them with context (`fmt.Errorf("context: %w", err)`)
- **Tests**: Include test files in same package with `_test.go` suffix 
- **Comments**: Document public functions with meaningful descriptions 
- **Line Length**: Keep lines under 80 characters where possible
- **Logging**: Use structured, context-rich logging messages
- **Types**: Declare interface types for dependency injection and testability