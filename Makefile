
run:
	go run ./cmd/bookcrossing

dev:
	air

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...


tidy:
	go mod tidy