# Developer entry points for ponder-media-catalog.
#
# Every Go operation runs inside the Docker dev container so no host Go
# toolchain is needed. HOST_UID/HOST_GID make the container run as the invoking
# user, so files it writes into the bind mount stay owned by the developer.

export HOST_UID := $(shell id -u)
export HOST_GID := $(shell id -g)

COMPOSE := docker compose
GO      := $(COMPOSE) run --rm dev go

.PHONY: test test-race vet fmt-check tidy-check lint build clean shell tidy

test:
	$(GO) test ./...

test-race:
	$(COMPOSE) run --rm -e CGO_ENABLED=1 dev go test -race ./...

vet:
	$(GO) vet ./...

# Local mirrors of the CI quality gate (see .github/workflows/ci.yml). The
# formatter and tidiness checks have no Go toolchain equivalent in the existing
# targets; the lint tool is pinned to the release recorded in .golangci.yml so a
# local run and a CI run report identical findings.
fmt-check:
	$(COMPOSE) run --rm dev bash -c 'test -z "$$(gofmt -l .)"'

tidy-check:
	$(GO) mod tidy && git diff --exit-code -- go.mod go.sum

lint:
	docker run --rm -v "$(CURDIR)":/src -w /src golangci/golangci-lint:v2.14.0 golangci-lint run

build:
	$(GO) build -o bin/ponder-catalog ./cmd/ponder-catalog/

clean:
	rm -rf bin/

shell:
	$(COMPOSE) run --rm dev bash

tidy:
	$(GO) mod tidy
