set positional-arguments

default:
  just --list

build:
  CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/denv cmd/denv/main.go

run *args:
  go run cmd/denv/main.go "$@"