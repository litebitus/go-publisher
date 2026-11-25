COVERAGE_PROFILE = /tmp/go-publisher-coverage.out
COVERAGE_FLOOR   = 88

.PHONY: build test

build: test
	go build -o bin/go-publisher .

test:
	go vet ./...
	@unformatted="$$(gofmt -l .)"; [ -z "$$unformatted" ] || { echo "not gofmt-formatted:"; echo "$$unformatted"; exit 1; }
	LC_ALL=C go test -race -count=1 -failfast -coverpkg=./... -coverprofile=$(COVERAGE_PROFILE) -timeout=15m ./...
	@go tool cover -func=$(COVERAGE_PROFILE) | awk -v floor=$(COVERAGE_FLOOR) '/^total:/ { \
		total = $$3 + 0; printf "coverage %s, floor %s%%\n", $$3, floor; \
		if (total < floor) { print "coverage is below the floor"; exit 1 } }'
	npm install --no-audit --no-fund
	node --test --test-timeout=5000 --test-force-exit schemas/
