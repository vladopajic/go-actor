GO ?= go
GOBIN ?= $$($(GO) env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.54.2
TEST_COVERAGE ?= $(GOBIN)/go-test-coverage

.PHONY: install-golangcilint
get-golangcilint:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

# Runs lint on entire repo
.PHONY: lint
lint: install-golangcilint
	$(GOLANGCI_LINT) run ./...

# Runs tests on entire repo
.PHONY: test
test: 
	go test -timeout=3s -race -count=10 -shuffle=on -failfast ./...

# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@latest

# Check test coverage
.PHONY: check-coverage
check-coverage: install-go-test-coverage
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	$(TEST_COVERAGE) -config=./.testcoverage.yml