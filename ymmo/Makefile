.PHONY: run build clean tidy

## run: start the development server
run:
	go run ./cmd/server/

## build: compile the binary
build:
	go build -o ymmo.exe ./cmd/server/

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove build artifacts and database
clean:
	Remove-Item -Force -ErrorAction SilentlyContinue ymmo.exe
	Remove-Item -Force -ErrorAction SilentlyContinue ymmo.db
