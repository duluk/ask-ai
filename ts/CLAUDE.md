# Ask-AI TypeScript Development Guide

## Build & Test Commands
- Install dependencies: `npm install`
- Build project: `npm run build`
- Run development mode: `npm run dev`
- Run TUI mode: `npm run tui`
- Run tests: `npm run test`
- Run single test: `jest path/to/test.test.ts`
- Lint code: `npm run lint`
- Format code: `npm run format`
- Package binaries: `npm run build:all`
- Install globally: `npm run install-app`

## Code Style Guidelines
- Use TypeScript strict mode with explicit typing
- Follow ES module import style (import from paths with `@/` prefix)
- Use async/await for asynchronous operations
- Place interfaces and types in dedicated `.ts` files
- Use camelCase for variables/methods, PascalCase for classes/interfaces
- Implement proper error handling with try/catch blocks
- Organize code with modular architecture (LLM providers, config, utils)
- Follow Jest patterns for tests with proper mocking
- Document public functions/methods with JSDoc comments
- Use prettier for consistent formatting