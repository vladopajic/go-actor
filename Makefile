GOBIN ?= $$(go env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.14.0 # LINT_VERSION: update version in other places
SYNCTEST_GO_VERSION ?= 1.27 # SYNCTEST_GO_VERSION: must match test_synctest.yml
TEST_COVERAGE ?= $(GOBIN)/go-test-coverage

.PHONY: install-golangcilint
install-golangcilint:
	test -f "$(GOLANGCI_LINT)" || curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$(GOBIN)" $(GOLANGCI_LINT_VERSION)
	test -x "$(GOLANGCI_LINT)"

# Runs lint on entire repo
.PHONY: lint
lint: install-golangcilint
	$(GOLANGCI_LINT) run ./...

# Runs benchmark on entire repo
.PHONY: benchmark
benchmark:
	go version
	go test -bench=. ./actor -run=^# -count 5 -benchmem

# Runs tests on entire repo
.PHONY: test
test: 
	go test -timeout=10s -race -count=10 -shuffle=on -failfast ./...

# Runs deterministic concurrency tests (requires Go 1.25+)
.PHONY: test-synctest
test-synctest:
	@orig=$$(go list -m -f '{{.GoVersion}}'); \
	trap 'go mod edit -go "$$orig"' EXIT INT TERM HUP; \
	go mod edit -go $(SYNCTEST_GO_VERSION); \
	go test -tags "synctest" -run Synctest$$ -timeout=20s -race -count=10 -shuffle=on -failfast -v ./...

# Code tidy
.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: install-go-test-coverage
install-go-test-coverage:
	test -f $(TEST_COVERAGE) || go install github.com/vladopajic/go-test-coverage/v2@latest


# Generates test coverage profile
.PHONY: generate-coverage
generate-coverage:
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...

# Runs test coverage check
.PHONY: check-coverage
check-coverage: generate-coverage install-go-test-coverage
	$(TEST_COVERAGE) -config=./.testcoverage.yml

# View coverage profile
.PHONY: view-coverage
view-coverage: generate-coverage
	go tool cover -html=cover.out -o=cover.html
	xdg-open cover.html
