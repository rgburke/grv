GOCMD=go
GOLINT=golint

BINARY=grv
SOURCE_DIR=./cmd/grv
GRV_DIR:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
GIT2GO_DIR=$(GRV_DIR)../../../gopkg.in/libgit2/git2go.v25
BUILD_FLAGS=--tags static

all: $(BINARY)

$(BINARY): build-libgit2
	$(GOCMD) build $(BUILD_FLAGS) -o $(BINARY) $(SOURCE_DIR)

.PHONY: install
install: build-libgit2
	$(GOCMD) install $(BUILD_FLAGS) $(SOURCE_DIR)

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
