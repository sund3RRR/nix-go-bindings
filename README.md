# nix-go-bindings

Experimental Go bindings for the [Nix](https://github.com/NixOS/nix) C API.

The Nix project is implemented in C++, and its public C API packages are thin C
facades over those C++ libraries. This repository turns that C API into Go via
[c-for-go](https://github.com/xlab/c-for-go). A small C shim lives in
`nix_go_store.h` and `nix_go_store.c` to translate awkward C API shapes, such as
callback-returned strings and `const char ***` store parameters, into signatures
that c-for-go can generate more cleanly.

This is currently a low-level binding package, not an idiomatic Go client.

## Quick Start

To build another Go project that imports this package, enter the development
environment from this flake:

```sh
nix develop github:sund3RRR/nix-go-bindings
```

That shell sets up Go, cgo, `pkg-config`, and the Nix C API libraries needed to
build projects using these bindings.

## How It Fits Together

- `flake.nix` provides the development shell, Nix C API libraries, pkg-config
  paths, Go, c-for-go, and the binding generation app.
- `nix-go-bindings.yml` is the c-for-go configuration.
- `nix_go_store.h` and `nix_go_store.c` are the local C shim layer.
- `nix.go`, `types.go`, `const.go`, `cgo_helpers.*`, and `doc.go` are generated.

## Contribution

### Setup environment

- Nix with flakes enabled.
- Go 1.23 or newer.
- The Nix C API development packages visible to `pkg-config`.

The easiest path is to work inside the dev shell:

```sh
git clone https://github.com/sund3RRR/nix-go-bindings.git && cd nix-go-bindings
nix develop
go test ./...
```

Outside the shell, Go builds need `pkg-config` to find at least
`nix-util-c` and `nix-store-c`, plus the same cgo-related environment that the
flake exports.

### Regenerating Bindings

```sh
nix run .#generate-go-bindings
```

The generator writes into a temporary directory and copies the generated package
files back into the repository root.

## Usage Notes

Current bindings are intentionally close to the C layer. Strings returned by the
shim are C-owned `*byte` values and must be released with `StringFree`.
Store handles should be released with `StoreFree`.

Some generated helper methods call `C.free` on opaque Nix objects. Prefer the
API-specific free functions exposed by this package when they exist; missing
free wrappers are tracked below.

## Upstream C API Surface

The upstream C API packages are:

- [`nix-util-c`](https://github.com/NixOS/nix/tree/master/src/libutil-c): common
  initialization, contexts, errors, settings, version, and verbosity.
- [`nix-store-c`](https://github.com/NixOS/nix/tree/master/src/libstore-c):
  stores, store paths, derivations, realization, closure traversal, and copying.
- [`nix-expr-c`](https://github.com/NixOS/nix/tree/master/src/libexpr-c):
  evaluation state, values, primops, external values, and GC hooks.
- [`nix-fetchers-c`](https://github.com/NixOS/nix/tree/master/src/libfetchers-c):
  fetcher settings.
- [`nix-flake-c`](https://github.com/NixOS/nix/tree/master/src/libflake-c):
  flake settings, references, lock flags, locking, and output lookup.
- [`nix-main-c`](https://github.com/NixOS/nix/tree/master/src/libmain-c):
  plugin initialization and log format.

## Implementation Status

- [x] `nix-util-c`
  - [x] Shared `nix_c_context` and `nix_err` types are imported for store calls.
  - [x] Context lifecycle: `CContextCreate`, `CContextFree`.
  - [x] Library initialization: `LibutilInit`.
  - [x] Settings/version/verbosity helpers.
  - [x] Error helpers: message, name, code, clear, set.
  - [x] Generated enum constants for `NIX_OK`, `NIX_ERR_*`, and `NIX_LVL_*`.
- [x] `nix-store-c`
  - [x] Store initialization: `LibstoreInit`, `LibstoreInitNoLoadConfig`.
  - [x] Store lifecycle: `StoreOpen`, `StoreFree`.
  - [x] Store strings: URI, store dir, version, real path.
  - [x] Store path parsing and validity checks.
  - [x] StorePath lifecycle and helpers: clone, free, name, hash, create from parts.
  - [x] Realization callback wrapper.
  - [x] Derivation JSON import and `AddDerivation`.
  - [x] Closure traversal and copy helpers.
  - [x] Query path by hash part.
  - [x] Derivation lifecycle and JSON export: clone, free, to JSON.
- [x] `nix-expr-c`
  - [x] `nix_libexpr_init`.
  - [x] Evaluation state builder and state lifecycle.
  - [x] Expression evaluation and function calls.
  - [x] Value allocation, ref-counting, forcing, getters, and initializers.
  - [x] Lists, attrsets, primops, external values, realized strings, and GC hooks.
- [x] `nix-fetchers-c`
  - [x] Fetchers settings lifecycle.
- [x] `nix-flake-c`
  - [x] Flake settings lifecycle and eval-state integration.
  - [x] Reference parse flags and reference parsing.
  - [x] Lock flags, input overrides, lock operation, locked flake lifecycle.
  - [x] Locked flake output attribute lookup.
- [x] `nix-main-c`
  - [x] Plugin initialization.
  - [x] Log format configuration.
