watch:
    air -c .air.toml

alias w := watch

dev:
    @go run ./dev/main.go

alias d := dev

test:
    @go test ./cmd/... ./pkg/...

alias t := test

lint:
    @golangci-lint run ./cmd/... ./pkg/...

alias l := lint

completion:
    @go run main.go completion zsh > "_seaq"
    @sudo mv _seaq /usr/share/zsh/site-functions/

install: completion
    @go install .

alias i := install

up:
    @docker compose -f compose.dev.yml up -d

down:
    @docker compose -f compose.dev.yml down

restart:
    @docker compose -f compose.dev.yml restart