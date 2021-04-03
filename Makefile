.DEFAULT_GOAL = test
.PHONY: FORCE

# Build
build: mangaconv
.PHONY: build

clean:
	rm -f mangaconv
.PHONY: clean

# Test
test:
	go test -v -race -cover ./...
bench:
	go test -bench=. -benchmem ./...
.PHONY: test bench

# Lint
lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix
.PHONY: lint lint-fix

# Non-PHONY targets (real files)
mangaconv: FORCE
	go build -o $@ ./cmd/mangaconv

go.mod: FORCE
	go mod tidy
	go mod verify
go.sum: go.mod
