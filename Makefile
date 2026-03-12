GO ?= go
NPM ?= npm

.PHONY: fmt test gateway cli web-install web-dev web-build web-lint clean

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

gateway:
	$(GO) run ./cmd/agentfence

cli:
	$(GO) run ./cmd/agentfence-cli

web-install:
	cd web && $(NPM) install

web-dev:
	cd web && $(NPM) run dev

web-build:
	cd web && $(NPM) run build

web-lint:
	cd web && $(NPM) run lint

clean:
	$(GO) clean ./...
