GOCMD=go
GOLINT=golint

BINARY=grv
SOURCE_DIR=./src
GIT2GO_DIR=$(GOPATH)/src/gopkg.in/libgit2/git2go.v25

SOURCES!=find $(SOURCE_DIR) -maxdepth 1 -name '*.go' ! -name '*_test.go' -type f
BUILD_FLAGS=-v --tags static

all: $(BINARY)

$(BINARY): build-libgit2 $(SOURCES)
	$(GOCMD) build $(BUILD_FLAGS) -o $(BINARY) $(SOURCE_DIR)

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
