# Makefile
.PHONY: run build migrate

run:
	go run cmd/main.go

build:
	go build -o bin/nid-backend cmd/main.go

migrate:
	psql $(DATABASE_URL) -f schema.sql
