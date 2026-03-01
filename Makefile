.PHONY: build run test

build:
	go build -o bin/pgcheckpoint .

run:
	go run . $(ARGS)

test:
	gotestsum

