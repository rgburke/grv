GOCMD=go
GOLINT=golint

BINARY?=grv
SOURCE_DIR=./cmd/grv
BUILD_FLAGS=--tags static

GRV_DIR:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
GOPATH_DIR:=$(GRV_DIR)../../../..
GOBIN_DIR:=$(GOPATH_DIR)/bin
GIT2GO_DIR:=$(GOPATH_DIR)/src/gopkg.in/libgit2/git2go.v25

all: $(BINARY)

$(BINARY): build-libgit2
	$(GOCMD) build $(BUILD_FLAGS) -o $(BINARY) $(SOURCE_DIR)

.PHONY: install
install: $(BINARY)
	install -m755 $(BINARY) $(GOBIN_DIR)

.PHONY: update
update:
	$(GOCMD) get -d ./...

.PHONY: update-test
update-test:
	$(GOCMD) get github.com/golang/lint/golint
	$(GOCMD) get github.com/stretchr/testify/mock

.PHONY: build-libgit2
build-libgit2: update
	cd $(GIT2GO_DIR) && git submodule update --init;
	make -C $(GIT2GO_DIR) install-static

.PHONY: test
test: $(BINARY) update-test
	$(GOCMD) test $(BUILD_FLAGS) $(SOURCE_DIR)
	$(GOCMD) vet $(SOURCE_DIR)
	$(GOLINT) $(SOURCE_DIR)

.PHONY: clean
clean:
	rm -f $(BINARY)
