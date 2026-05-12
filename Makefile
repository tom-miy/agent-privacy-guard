GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod
GOENV := GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)

.PHONY: test vet cover tidy check

test:
	$(GOENV) go test ./...

vet:
	$(GOENV) go vet ./...

cover:
	$(GOENV) go test -cover ./...

tidy:
	$(GOENV) go mod tidy

check: test vet
