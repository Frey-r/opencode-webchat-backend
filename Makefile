.PHONY: build run test migrate dev clean docker-build docker-up docker-down

BINARY=opencode-webchat
MAIN=cmd/server/main.go
GOOSE=/opt/data/go/bin/goose
DB_URL?=postgres://postgres:postgres@localhost:5432/opencode_webchat?sslmode=disable

build:
	go build -o $(BINARY) $(MAIN)

run: build
	./$(BINARY)

dev:
	go run $(MAIN)

test:
	go test ./...

# Run migrations using goose
migrate-up:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" up

migrate-down:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" down

migrate-status:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" status

# Docker commands
docker-build:
	docker build -t opencode-webchat .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

clean:
	rm -f $(BINARY)