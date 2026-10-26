GRV_VERSION=$(shell git describe --long --tags --dirty --always --match=v*.*.* 2>/dev/null || echo 'Unknown')
GRV_BUILD_DATETIME=$(shell date '+%Y-%m-%d %H:%M:%S %Z')

GOCMD?=go

BINARY?=grv
GRV_SOURCE_DIR=./cmd/grv
GRV_DIR:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
GOBIN_DIR:=$(shell $(GOCMD) env GOPATH)/bin

# libgit2 v0.27 (required by git2go v27) is built from source and linked
# statically. pkg-config is pointed at the local install so git2go finds it.
LIBGIT2_VERSION?=0.27.10
LIBGIT2_PREFIX?=$(GRV_DIR)build/libgit2
LIBGIT2_PC=$(LIBGIT2_PREFIX)/lib/pkgconfig/libgit2.pc
export PKG_CONFIG_PATH:=$(LIBGIT2_PREFIX)/lib/pkgconfig$(if $(PKG_CONFIG_PATH),:$(PKG_CONFIG_PATH))

GRV_TAGS=static system_libgit2
GRV_LDFLAGS=-X 'main.version=$(GRV_VERSION)' -X 'main.buildDateTime=$(GRV_BUILD_DATETIME)'
GRV_STATIC_LDFLAGS=-linkmode external -extldflags '-static'
GRV_BUILD_FLAGS=-tags '$(GRV_TAGS)' -ldflags "$(GRV_LDFLAGS)"
GRV_STATIC_BUILD_FLAGS=-tags '$(GRV_TAGS)' -ldflags "$(GRV_LDFLAGS) $(GRV_STATIC_LDFLAGS)"

# On macOS readline comes from Homebrew (the system one is libedit)
ifeq ($(shell uname -s),Darwin)
BREW_READLINE:=$(shell brew --prefix readline 2>/dev/null)
ifneq ($(BREW_READLINE),)
export CGO_CFLAGS:=$(CGO_CFLAGS) -I$(BREW_READLINE)/include
export CGO_LDFLAGS:=$(CGO_LDFLAGS) -L$(BREW_READLINE)/lib
endif
endif

all: $(BINARY)

.PHONY: $(BINARY)
$(BINARY): $(LIBGIT2_PC)
	$(GOCMD) build $(GRV_BUILD_FLAGS) -o $(BINARY) $(GRV_SOURCE_DIR)

$(LIBGIT2_PC):
	LIBGIT2_VERSION=$(LIBGIT2_VERSION) ./script/build-libgit2.sh "$(LIBGIT2_PREFIX)"

.PHONY: build-libgit2
build-libgit2: $(LIBGIT2_PC)

.PHONY: install
install: $(BINARY)
	install -m755 -d $(GOBIN_DIR)
	install -m755 $(BINARY) $(GOBIN_DIR)

# Fully static binary. Linux only: requires the static versions of
# ncurses, readline and gpm to be installed (e.g. the -dev packages on Ubuntu).
.PHONY: static
static: $(LIBGIT2_PC)
	CGO_LDFLAGS="$(CGO_LDFLAGS) -lncursesw -ltinfo" \
		$(GOCMD) build $(GRV_STATIC_BUILD_FLAGS) -o $(BINARY) $(GRV_SOURCE_DIR)

.PHONY: test
test: $(LIBGIT2_PC)
	$(GOCMD) vet -tags '$(GRV_TAGS)' $(GRV_SOURCE_DIR)
	$(GOCMD) test -tags '$(GRV_TAGS)' $(GRV_SOURCE_DIR)

.PHONY: doc
doc: $(BINARY)
	@GRV_GENERATE_DOCUMENTATION=1 ./$(BINARY)

.PHONY: clean
clean:
	rm -f $(BINARY)

.PHONY: distclean
distclean: clean
	rm -rf $(GRV_DIR)build
