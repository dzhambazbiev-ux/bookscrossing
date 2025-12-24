
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

install-hook:
	@chmod +x scripts/pre-commit
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook установлен"
	@echo ""
	@echo "Не забудьте установить OPENAI_API_KEY:"
	@echo "  export OPENAI_API_KEY='your-api-key-here'"

test-hook:
	@go run ./scripts/ai-precommit.go	