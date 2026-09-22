GO := $(if $(wildcard $(HOME)/.local/toolchains/go/bin/go),$(HOME)/.local/toolchains/go/bin/go,go)
.PHONY: test build run web-build
test:
	$(GO) test ./...
build:
	CGO_ENABLED=1 $(GO) build -buildvcs=false -o bin/panchang-api ./cmd/panchang-api
run:
	EPHE_PATH=./ephe $(GO) run ./cmd/panchang-api
web-build:
	cd web && npm run build
