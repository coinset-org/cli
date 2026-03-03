.PHONY: all build rustlib test clean run

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ifeq ($(GOOS),darwin)
  MACOSX_DEPLOYMENT_TARGET ?= 12.0
  export MACOSX_DEPLOYMENT_TARGET

  # Ensure cgo compilation uses the same min version as linking.
  CGO_CFLAGS ?= -mmacosx-version-min=$(MACOSX_DEPLOYMENT_TARGET)
  CGO_LDFLAGS ?= -mmacosx-version-min=$(MACOSX_DEPLOYMENT_TARGET)
  export CGO_CFLAGS
  export CGO_LDFLAGS
endif

# Map GOOS/GOARCH to a Rust target triple we build the staticlib for.
ifeq ($(GOOS),darwin)
  ifeq ($(GOARCH),arm64)
    RUST_TARGET ?= aarch64-apple-darwin
  else
    RUST_TARGET ?= x86_64-apple-darwin
  endif
else ifeq ($(GOOS),linux)
  ifeq ($(GOARCH),amd64)
    RUST_TARGET ?= x86_64-unknown-linux-gnu
  else
    RUST_TARGET ?= aarch64-unknown-linux-gnu
  endif
else ifeq ($(GOOS),windows)
  RUST_TARGET ?= x86_64-pc-windows-gnu
else
  RUST_TARGET ?= $(shell rustc -vV | awk '/host:/ {print $$2}')
endif

LIBDIR := cgo-lib/$(RUST_TARGET)
LIB := $(LIBDIR)/libcoinset.a

all: build

$(LIB):
	mkdir -p "$(LIBDIR)"
	cargo build --release -p coinset-ffi --target "$(RUST_TARGET)"
	cp "target/$(RUST_TARGET)/release/libcoinset.a" "$(LIBDIR)/"

rustlib: $(LIB)

build: rustlib
	CGO_ENABLED=1 go build -o bin/coinset ./cmd/coinset

run: rustlib
	CGO_ENABLED=1 go run ./cmd/coinset

test: rustlib
	cargo test -p coinset-inspect-core
	CGO_ENABLED=1 go test ./...

clean:
	rm -rf bin cgo-lib
