# Developer entry points for ponder-media-catalog.
#
# Every Go operation runs inside the Docker dev container so no host Go
# toolchain is needed. HOST_UID/HOST_GID make the container run as the invoking
# user, so files it writes into the bind mount stay owned by the developer.

export HOST_UID := $(shell id -u)
export HOST_GID := $(shell id -g)

COMPOSE := docker compose
GO      := $(COMPOSE) run --rm dev go

.PHONY: test test-race vet build clean shell tidy

test:
	$(GO) test ./...

test-race:
	$(COMPOSE) run --rm -e CGO_ENABLED=1 dev go test -race ./...

vet:
	$(GO) vet ./...

build:
	$(GO) build -o bin/ponder-catalog ./cmd/ponder-catalog/

clean:
	rm -rf bin/

shell:
	$(COMPOSE) run --rm dev bash

tidy:
	$(GO) mod tidy
