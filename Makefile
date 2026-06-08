.PHONY: build run test clean docker-build docker-run

APP_NAME ?= geekbot
BIN_DIR  ?= bin

build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/bot

run:
	go run ./cmd/bot

test:
	go test ./... -count=1 -race

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)/

docker-build:
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		--build-arg no_proxy \
		-t $(APP_NAME) .

docker-run:
	docker compose up -d --build

docker-stop:
	docker compose down

docker-logs:
	docker compose logs -f
