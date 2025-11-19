.PHONY: build run tidy

## run: Run the app.
run:
	@go run ./cmd/server

tidy:
	go mod tidy

# build current or local machine
build: tidy
	go build -o dist/server ./cmd/server