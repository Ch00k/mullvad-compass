.PHONY: lint test test-verbose test-one test-ci build run release release-patch release-minor release-major

.EXPORT_ALL_VARIABLES:

MULLVAD_COMPASS_EXECUTABLE_FILENAME ?= mullvad-compass
MULLVAD_COMPASS_BUILD_ARTIFACTS_DIR ?= dist
MULLVAD_COMPASS_VERSION ?= dev

lint:
	golangci-lint run --fix

test:
	gotestsum --format testname -- ./...

test-verbose:
	gotestsum --format standard-verbose -- -v -count=1 ./...

test-one:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-one TEST=TestName"; \
		exit 1; \
	fi
	gotestsum --format standard-verbose -- -v -count=1 -run "^$(TEST)$$" ./...

test-ci:
	go run gotest.tools/gotestsum@latest --format testname -- -race "-coverprofile=coverage.txt" "-covermode=atomic" ./...

build:
	go build -trimpath -ldflags="-s -w -X main.Version=${MULLVAD_COMPASS_VERSION}" -o ./${MULLVAD_COMPASS_BUILD_ARTIFACTS_DIR}/${MULLVAD_COMPASS_EXECUTABLE_FILENAME} ./cmd/mullvad-compass

run:
	go run ./cmd/mullvad-compass

release:
	@echo "Available release types:"
	@echo "  make release-patch  # Patch version (x.y.Z)"
	@echo "  make release-minor  # Minor version (x.Y.0)"
	@echo "  make release-major  # Major version (X.0.0)"

release-patch:
	./release.sh patch

release-minor:
	./release.sh minor

release-major:
	./release.sh major
