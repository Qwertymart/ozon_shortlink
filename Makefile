.PHONY: run test build docker-up docker-down clean help

run:
	go run cmd/shortener/main.go

test:
	go test -v -cover ./...

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v