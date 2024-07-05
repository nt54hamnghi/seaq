watch:
    air -c .air.toml

alias w := watch

dev:
    @go run ./dev/main.go

alias d := dev

test:
    @go test ./...

alias t := test