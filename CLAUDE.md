# Ask-AI Development Guide

## Build & Test Commands

### Go Version
- Build: `make` or `make build`
- Test all: `make test` 
- Test verbose: `make test VERBOSE=1`
- Test single: `go test -v ./pkg/[package]/[file]_test.go`
- Lint/Check: `make check` (runs go vet)
- Format code: `make fmt` (runs go fmt ./...)
- Install: `make install`
- Run TUI mode: `ask-ai --tui` or with flag `-tui`

### TypeScript Version
- Install dependencies: `npm install`
- Build: `npm run build`
- Run directly: `npm run start -- [args]`
- Run TUI mode: `npm run tui`
- Lint code: `npm run lint`
- Format code: `npm run format`
- Install globally: `npm install -g`

## Code Style Guidelines

### Go
- Imports: Standard library first, third-party second, project imports last
- Error handling: Early returns with meaningful error messages
- Types: Define interfaces in pkg/LLM/types.go for client interactions
- Naming: Go standard camelCase variables, PascalCase for exported entities
- Comments: Document public functions/types with meaningful descriptions
- Keep functions focused and under 50 lines when possible
- Use contexts for operation control where appropriate
- Favor composition over inheritance
- Follow Go standard project layout with cmd/ and pkg/ directories

### TypeScript
- Use TypeScript types for all variables and functions
- Follow standard ES module imports
- Use async/await for asynchronous operations
- Maintain consistent error handling patterns
- Follow Jest patterns for tests with proper mocking

## TUI Mode
- Special commands in TUI:
  - `/exit`, `/quit` - Exit the application
  - `/help`, `/?` - Show help message
  - `/model` - Show current model
  - `/model NAME` - Change model to NAME
  - `/id` - Show conversation ID
  - `/clear` - Clear conversation history
  - `/context` - Show current context

## Project Structure
- cmd/ask-ai/main.go: Entry point with CLI handling for Go version
- pkg/: Core functionality organized by provider and responsibility
- ts/: TypeScript implementation
- Modular design enables adding new LLM providers with minimal changes