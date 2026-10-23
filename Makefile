GO ?= go
.PHONY: build test race vet docs demo cross
build:
	$(GO) build -o bin/reconcilelab ./cmd/reconcilelab
test:
	$(GO) test ./...
race:
	$(GO) test -race ./...
vet:
	$(GO) vet ./...
docs:
	python3 scripts/check_docs.py
demo: build
	python3 scripts/record_demo.py
cross:
	python3 scripts/cross.py
