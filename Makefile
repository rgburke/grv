GOCMD=go
GOLINT=golint

BINARY=grv
SOURCE_DIR=./src

SOURCES!=find $(SOURCE_DIR) -maxdepth 1 -name '*.go' ! -name '*_test.go' -type f

all: $(BINARY)

$(BINARY): $(SOURCES)
	$(GOCMD) build -o $(BINARY) $(SOURCE_DIR)

.PHONY: update
update:
	$(GOCMD) get ./...

.PHONY: test
test: update $(BINARY)
	$(GOCMD) test $(SOURCE_DIR)
	$(GOCMD) vet $(SOURCE_DIR)
	$(GOLINT) $(SOURCE_DIR)

.PHONY: clean
clean:
	rm -f $(BINARY)
