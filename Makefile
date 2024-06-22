run:
	go run cmd/service/main.go

lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run --fix -c .golangci.yml ./...

test: up
	go test -v ./...

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose down
	docker compose up -d
