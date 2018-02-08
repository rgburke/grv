GRV_VERSION=$(shell git describe --long --tags --dirty --always 2>/dev/null || echo 'Unknown')
GRV_HEAD_OID=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'Unknown')
GRV_BUILD_DATETIME=$(shell date '+%Y-%m-%d %H:%M:%S %Z')

GOCMD=go
GOLINT=golint

BINARY?=grv
GRV_SOURCE_DIR=./cmd/grv
GRV_LDFLAGS=-X 'main.version=$(GRV_VERSION)' -X 'main.headOid=$(GRV_HEAD_OID)' -X 'main.buildDateTime=$(GRV_BUILD_DATETIME)'
GRV_STATIC_LDFLAGS=-extldflags '-lncurses -ltinfo -lgpm -static'
GRV_BUILD_FLAGS=--tags static -ldflags "$(GRV_LDFLAGS)"
GRV_STATIC_BUILD_FLAGS=--tags static -ldflags "$(GRV_LDFLAGS) $(GRV_STATIC_LDFLAGS)"

GRV_DIR:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
GOPATH_DIR:=$(shell go env GOPATH)
GOBIN_DIR:=$(GOPATH_DIR)/bin
GIT2GO_DIR:=$(GRV_SOURCE_DIR)/vendor/gopkg.in/libgit2/git2go.v25
GIT2GO_PATCH=git2go.v25.patch

export PKG_CONFIG=$(GRV_DIR)/pkg-config-wrapper.sh

all: $(BINARY)

$(BINARY): build-libgit2
	$(GOCMD) build $(GRV_BUILD_FLAGS) -o $(BINARY) $(GRV_SOURCE_DIR)

.PHONY: build-only
build-only:
	make -C $(GIT2GO_DIR) install-static
	$(GOCMD) build $(GRV_BUILD_FLAGS) -o $(BINARY) $(GRV_SOURCE_DIR)

.PHONY: build-libgit2
build-libgit2: apply-git2go-patch
	make -C $(GIT2GO_DIR) install-static

.PHONY: install
install: $(BINARY)
	install -m755 -d $(GOBIN_DIR)
	install -m755 $(BINARY) $(GOBIN_DIR)

.PHONY: update
update:
	-@ git submodule foreach --recursive git reset --hard >/dev/null 2>&1
	git submodule update --init --recursive

.PHONY: update-test
update-test:
	$(GOCMD) get github.com/golang/lint/golint
	$(GOCMD) get github.com/stretchr/testify/mock

.PHONY: apply-git2go-patch
apply-git2go-patch: update
	if patch --dry-run -N -d $(GIT2GO_DIR) -p1 < $(GIT2GO_PATCH) >/dev/null; then \
		patch -d $(GIT2GO_DIR) -p1 < $(GIT2GO_PATCH); \
	fi

# Only tested on Ubuntu.
# Requires dependencies static library versions to be present alongside dynamic ones
.PHONY: static
static: build-libgit2
	$(GOCMD) build $(GRV_STATIC_BUILD_FLAGS) -o $(BINARY) $(GRV_SOURCE_DIR)

.PHONY: test
test: $(BINARY) update-test
	$(GOCMD) test $(GRV_BUILD_FLAGS) $(GRV_SOURCE_DIR)
	$(GOCMD) vet $(GRV_SOURCE_DIR)
	$(GOLINT) -set_exit_status $(GRV_SOURCE_DIR)

.PHONY: clean
clean:
	rm -f $(BINARY)
