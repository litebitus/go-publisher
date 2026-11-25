COVERAGE_PROFILE = /tmp/go-publisher-coverage.out
COVERAGE_FLOOR   = 88

.PHONY: build test test-action

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
	node --test --test-timeout=5000 --test-force-exit 'schemas/*.test.js'

test-action:
	@temp="$$(mktemp -d)"; trap 'rm -rf "$$temp"' EXIT; \
	case "$$(uname -s)" in Darwin) os=macOS ;; Linux) os=Linux ;; *) os="$$(uname -s)" ;; esac; \
	case "$$(uname -m)" in arm64|aarch64) arch=ARM64 ;; x86_64) arch=X64 ;; *) arch="$$(uname -m)" ;; esac; \
	VERSION="$$(awk '$$1 == "VERSION:" { print $$2 }' action.yml)" RUNNER_OS_NAME="$$os" RUNNER_ARCHITECTURE="$$arch" \
		RUNNER_TEMP="$$temp" GITHUB_PATH="$$temp/github-path" bash scripts/install-go-publisher.sh && \
	"$$(cat "$$temp/github-path")/go-publisher" check
