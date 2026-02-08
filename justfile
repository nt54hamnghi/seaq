[group('meta')]
[private]
default:
    @just --list --justfile {{ justfile() }} --unsorted

# watch for changes, see .air.toml for configuration
[group('dev')]
watch:
    air -c .air.toml

alias w := watch

# run dev binary in ./dev/main.go
[group('dev')]
dev:
    @go run ./dev/main.go

alias d := dev

# run all tests
[group('test')]
test:
    # -count=2 - run each test twice
    # -race - run tests with race detection
    # -shuffle=on - shuffle tests to catch flakiness
    # -cover - show test coverage
    # -covermode=atomic - thread-safe coverage for race testing
    @go test ./cmd/... ./pkg/... -count=2 -race -shuffle=on -cover -covermode=atomic

alias t := test

# run all tests without coverage, race detection, and shuffle
[group('test')]
test-quick:
    @go test ./cmd/... ./pkg/...

alias tq := test-quick

# lint project source code with golangci-lint
[group('lint')]
lint:
    @golangci-lint run ./cmd/... ./pkg/...

alias l := lint

# generate and install zsh completion for seaq
[group('install')]
completion:
    @go run main.go completion zsh > "_seaq"
    @sudo mv _seaq /usr/share/zsh/site-functions/

# build and install CLI, also setup shell completion
[group('install')]
install: completion
    @go install .

alias i := install

# start dev chroma container
[group('chroma')]
up:
    @docker compose -f compose.dev.yml up -d

# stop dev chroma container
[group('chroma')]
down:
    @docker compose -f compose.dev.yml down

# restart dev chroma container
[group('chroma')]
restart:
    @docker compose -f compose.dev.yml restart
