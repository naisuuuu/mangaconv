.DEFAULT_GOAL = test
.PHONY: FORCE

# Test
test:
	go test -v -race -cover ./...
.PHONY: test

bench:
	go test -bench=. -benchmem ./...
.PHONY: bench

lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix
.PHONY: lint lint_fix

# Non-PHONY targets (real files)
go.mod: FORCE
	go mod tidy
	go mod verify
go.sum: go.mod
