# Runs lint on entire repo
.PHONY: lint
lint: 
	golangci-lint run

# Runs tests on entire repo
.PHONY: test
test: 
	go test ./...

# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...