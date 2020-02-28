#!/bin/sh
GOOS=js GOARCH=wasm go build -o ./web/main.wasm main.go
go build -o run cmd/run/main.go

