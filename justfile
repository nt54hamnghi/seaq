watch:
    air -c .air.toml

alias w := watch

dev:
    @go run ./dev/main.go

alias d := dev

test:
    @go test ./...

alias t := test

up:
    @docker compose -f compose.dev.yml up -d

down:
    @docker compose -f compose.dev.yml down

restart:
    @docker compose -f compose.dev.yml restart