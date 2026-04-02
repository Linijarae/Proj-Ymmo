.PHONY: run build clean tidy docker-build docker-up docker-down docker-logs docker-ps

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

## docker-build: build application and database images
docker-build:
	docker compose build

## docker-up: start all containers in background
docker-up:
	docker compose up -d --build

## docker-down: stop and remove containers
docker-down:
	docker compose down

## docker-logs: show recent container logs
docker-logs:
	docker compose logs --tail=150

## docker-ps: display running service status
docker-ps:
	docker compose ps
