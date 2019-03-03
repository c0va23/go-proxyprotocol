GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

STATICCHECK_VERSION := 2019.1
STATICCHECK_URL := https://github.com/dominikh/go-tools/releases/download/$(STATICCHECK_VERSION)/staticcheck_$(GOOS)_$(GOARCH)

export PATH := $(PWD)/bin:$(PATH)

bin/staticcheck_$(STATICCHECK_VERSION):
	mkdir bin
	curl -o $@ -L $(STATICCHECK_URL)

bin/staticcheck: bin/staticcheck_$(STATICCHECK_VERSION)
	ln -s $(PWD)/bin/staticcheck_$(STATICCHECK_VERSION) $@
	chmod +x $@

lint: bin/staticcheck
	which staticcheck
	staticcheck ./...

test:
	go test ./...