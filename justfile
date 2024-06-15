watch:
    air -c .air.toml

alias w := watch

dev:
    @go run ./dev/main.go

alias d := dev

run:
    @go run ./todo/main.go

alias r := run