.PHONY: run test build docker-up

run:
	go run cmd/shortener/main.go

test:
	go test -v -cover ./...

docker-up:
	docker-compose up --build