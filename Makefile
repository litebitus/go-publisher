COVERAGE_PROFILE = /tmp/go-publisher-coverage.out

.PHONY: build test

build: test
	go build -o bin/go-publisher .

test:
	go vet ./...
	@unformatted="$$(gofmt -l .)"; [ -z "$$unformatted" ] || { echo "not gofmt-formatted:"; echo "$$unformatted"; exit 1; }
	LC_ALL=C go test -race -count=1 -failfast -coverpkg=./... -coverprofile=$(COVERAGE_PROFILE) -timeout=15m ./...
	@go tool cover -func=$(COVERAGE_PROFILE) | awk '/^total:/ { printf "coverage %s\n", $$3 }'
	npm ci --prefix schemas --no-audit --no-fund
	node --test --test-timeout=5000 --test-force-exit schemas/
