#!/bin/sh
#
# Download and build a static copy of libgit2 for git2go v27.
#
# git2go v27 only supports libgit2 v0.27.x, which is no longer shipped by
# current Linux distributions or Homebrew, so it is built from source and
# installed into a local prefix. The Makefile points pkg-config at that prefix.
#
# Usage: script/build-libgit2.sh [install-prefix]

set -eu

LIBGIT2_VERSION="${LIBGIT2_VERSION:-0.27.10}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PREFIX="${1:-$ROOT_DIR/build/libgit2}"
WORK_DIR="$ROOT_DIR/build/libgit2-src"
SRC_DIR="$WORK_DIR/libgit2-$LIBGIT2_VERSION"
TARBALL="$WORK_DIR/libgit2-$LIBGIT2_VERSION.tar.gz"
URL="https://github.com/libgit2/libgit2/archive/refs/tags/v$LIBGIT2_VERSION.tar.gz"

if [ -f "$PREFIX/lib/pkgconfig/libgit2.pc" ] &&
	grep -q "^Version: $LIBGIT2_VERSION\$" "$PREFIX/lib/pkgconfig/libgit2.pc"; then
	echo "libgit2 $LIBGIT2_VERSION already installed in $PREFIX"
	exit 0
fi

mkdir -p "$WORK_DIR"

if [ ! -d "$SRC_DIR" ]; then
	if [ ! -f "$TARBALL" ]; then
		echo "Downloading libgit2 $LIBGIT2_VERSION"
		curl -fsSL -o "$TARBALL.tmp" "$URL"
		mv "$TARBALL.tmp" "$TARBALL"
	fi
	tar -xzf "$TARBALL" -C "$WORK_DIR"
fi

# The bundled zlib defines fdopen() as NULL when TARGET_OS_MAC is set, which
# breaks against current macOS SDKs. Apply the upstream zlib 1.2.12 fix.
ZUTIL_H="$SRC_DIR/deps/zlib/zutil.h"
if grep -q '^#if defined(MACOS) || defined(TARGET_OS_MAC)$' "$ZUTIL_H"; then
	sed 's/^#if defined(MACOS) || defined(TARGET_OS_MAC)$/#if defined(MACOS)/' \
		"$ZUTIL_H" > "$ZUTIL_H.tmp"
	mv "$ZUTIL_H.tmp" "$ZUTIL_H"
fi

if command -v nproc >/dev/null 2>&1; then
	JOBS="$(nproc)"
else
	JOBS="$(sysctl -n hw.ncpu 2>/dev/null || echo 2)"
fi

rm -rf "$SRC_DIR/build"
mkdir -p "$SRC_DIR/build"
cd "$SRC_DIR/build"

# CMAKE_POLICY_VERSION_MINIMUM lets CMake >= 4 configure this older project,
# and the -Wno-*/--no-warn-unused-cli flags silence warnings about its outdated
# CMake setup (older CMake versions also don't know the policy variable).
# Network transports are disabled: grv only uses libgit2 to read local
# repositories and shells out to git for anything touching a remote.
cmake \
	-Wno-dev \
	-Wno-deprecated \
	--no-warn-unused-cli \
	-DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
	-DCMAKE_BUILD_TYPE=RelWithDebInfo \
	-DCMAKE_INSTALL_PREFIX="$PREFIX" \
	-DCMAKE_C_FLAGS="-fPIC -w" \
	-DBUILD_SHARED_LIBS=OFF \
	-DBUILD_CLAR=OFF \
	-DBUILD_EXAMPLES=OFF \
	-DTHREADSAFE=ON \
	-DUSE_EXT_HTTP_PARSER=OFF \
	-DUSE_HTTPS=OFF \
	-DUSE_NSEC=OFF \
	-DUSE_SSH=OFF \
	-DCURL=OFF \
	-DUSE_GSSAPI=OFF \
	-DUSE_BUNDLED_ZLIB=ON \
	..

cmake --build . --parallel "$JOBS"
cmake --build . --target install

echo "libgit2 $LIBGIT2_VERSION installed in $PREFIX"
