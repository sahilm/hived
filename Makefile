.PHONY: all
all: setup lint test

.PHONY: test
test: setup
	go test

sources = $(shell find . -name '*.go' -not -path './vendor/*')
.PHONY: goimports
goimports: setup
	goimports -w $(sources)

.PHONY: lint
lint: setup
	gometalinter ./... --enable=goimports --enable=gosimple \
	--enable=unparam --enable=unused --disable=gotype --vendor -t

.PHONY: check
check: setup
	gometalinter ./... --disable-all --enable=vet --enable=vetshadow \
	--enable=goimports --vendor -t

.PHONY: cover
cover: setup
	mkdir -p coverage
	gocov test ./... | gocov-html > coverage/index.html
	open coverage/index.html

.PHONY: ci
ci: setup check test

.PHONY: install
install: setup
	go install

.PHONY: build
build: setup
	go build

BIN_DIR := $(GOPATH)/bin
GOIMPORTS := $(BIN_DIR)/goimports
GOMETALINTER := $(BIN_DIR)/gometalinter
DEP := $(BIN_DIR)/dep
GOCOV := $(BIN_DIR)/gocov
GOCOV_HTML := $(BIN_DIR)/gocov-html

$(GOIMPORTS):
	go get -u golang.org/x/tools/cmd/goimports

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

$(GOCOV):
	go get -u github.com/axw/gocov/gocov

$(GOCOV_HTML):
	go get -u gopkg.in/matm/v1/gocov-html

tools: $(GOIMPORTS) $(GOMETALINTER) $(DEP) $(GOCOV) $(GOCOV_HTML)

vendor: $(DEP)
	dep ensure

setup: tools vendor

updatedeps:
	dep ensure -update

BINARY := hived
VERSION ?= latest

.PHONY: linux
linux: setup
	mkdir -p $(CURDIR)/release
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" \
	-o release/$(BINARY)-$(VERSION)-linux-amd64

.PHONY: release
release: linux
