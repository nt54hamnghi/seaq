# Repository Guidelines

## Project Structure & Module Organization
- `cmd/`: CLI subcommands and entrypoints.
- `pkg/`: core library packages (LLM providers, loaders, utilities).
- `dev/`: local dev helpers and prototypes (e.g., `dev/main.go`).
- `main.go`: production CLI entrypoint.
- `imgs/`: documentation assets (README media).
- `compose.dev.yml`: local ChromaDB stack.

## Tech Stack

### Go-specific
- Go
- `cobra`
- `viper`
- `bubbletea`
- `langchaingo`

### Others
- `docker`
- `chromadb`

## Build, Test, and Development Commands
- `just dev`: run the development binary from `dev/main.go`.
- `just watch`: live-reload using `air` and `.air.toml`.
- `just test`: full test suite with race, shuffle, and coverage.
- `just test-quick`: fast tests without coverage/race/shuffle.
- `just lint`: run `golangci-lint` on `cmd/` and `pkg/`.
- `just up` / `just down` / `just restart`: manage the dev ChromaDB container.
- `just install`: install the CLI and zsh completion.

## Coding Style & Naming Conventions
- Go formatting: use `gofmt` (standard Go style).
- Indentation: tabs in Go files, 4 spaces in YAML.
- Exported identifiers: `CamelCase`, unexported: `camelCase`.
- Keep comments concise; doc comments should start with the identifier name.
- Target Go version: 1.24.0 or newer.
- Prefer idiomatic, standard-library-first solutions; keep APIs simple and secure.

## Testing Guidelines
- Tests live alongside code in `*_test.go` files.
- Use `go test ./cmd/... ./pkg/...` (see `just test`).
- Use table-driven tests and clear case names.
- Use `testify` for assertions in parameterized tests.
- Use `testify/suite` for setup & teardown behaviors.
- Cover edge cases and error paths; test plan should be explicit for complex features.

## Commit & Pull Request Guidelines
- Commit messages follow Conventional Commits (e.g., `fix: ...`, `chore(deps): ...`).
- PR titles often use sentence case and may include issue numbers (e.g., `Preserve case in model identifiers (#17)`).
- PRs should include a short summary, motivation, and testing notes.

## Review & Planning Expectations
- Code review focus: logic correctness and security issues first, then maintainability.
- For new features, consider security implications, implementation complexity, and UX impact before coding.

## Configuration & Environment Tips
- Config lives in `~/.config/seaq.yaml` by default.
- `seaq` reads API keys from environment variables (e.g., `OPENAI_API_KEY`, `HUGGINGFACE_API_KEY`, `CHROMA_URL`).
- See `README.md` for the full list of supported env vars.
