# AGENTS.md

Guidance for agents working on this repository.

## Project Goal

`nix-go-bindings` provides low-level, mostly auto-generated Go bindings for the
Nix C API.

This repository is not intended to become a high-level Go Nix SDK. Keep the API
surface close to upstream Nix C packages, with only the small C shim layer needed
to make binding generation reliable and Go-compatible. Higher-level workflows,
policy, convenience clients, and SDK-style abstractions belong outside this
package or in a separate layer.

The core pieces are:

- Nix as the build system and development environment.
- c-for-go as the binding generator.
- Small C shim headers/sources that adapt the upstream Nix C API for c-for-go.
- Generated Go files in package `nix`.
- Go tests that exercise generated bindings by upstream C package area.

## Binding Workflow

Cover each original upstream Nix C package deliberately. Current upstream C API
packages include:

- `nix-util-c`
- `nix-store-c`
- `nix-expr-c`
- `nix-fetchers-c`
- `nix-flake-c`
- `nix-main-c`

For each package:

1. Inspect the upstream headers and tests in the Nix repository.
2. Add or extend a small local shim for the package.
3. Update `nix-go-bindings.yml` so c-for-go receives clear includes, pkg-config
   dependencies, accept/reject rules, rename rules, pointer tips, memory tips,
   and callback hints.
4. Regenerate bindings with `nix run .#generate-go-bindings`.
5. Add Go tests for the generated API area.
6. Run `nix develop -c go test ./...`.

## C Shim Rules

The shim exists to make the upstream C API easier for c-for-go to parse and
safer for Go to call. Keep it small and transparent.

Use the shim to:

- Hide or rewrite signatures that c-for-go cannot generate cleanly.
- Remove features from newer C standards when the generator expects C89-C99
  compatible headers.
- Avoid extreme pointer shapes and deeply nested C signatures where a simple
  struct or callback adapter is clearer.
- Convert callback-returned strings into explicit owned strings when useful.
- Convert awkward arrays, null-terminated structures, or `const char ***` style
  arguments into generator-friendly structs.
- Preserve upstream ownership rules and expose package-specific free functions.
- Keep function names predictable with a `go_nix_` prefix before generator
  renaming.

Do not use the shim to:

- Add high-level Nix behavior.
- Change semantics from the upstream C API.
- Hide important ownership or error handling requirements.
- Build a separate SDK-style abstraction layer.

## Generator Configuration

`nix-go-bindings.yml` should be clear and intentional. When adding symbols, also
add the hints needed for maintainability:

- `GENERATOR.Includes` should point at local shim headers, not directly at large
  upstream header sets unless they are known to generate cleanly.
- `GENERATOR.PkgConfigOpts` should list the Nix C packages needed by the
  included shims.
- `PARSER.Defines` should pin C compatibility assumptions, such as
  `__STDC_VERSION__: 199901`.
- `TRANSLATOR.Rules` should accept only the intended shim symbols and required
  opaque upstream types.
- `TRANSLATOR.MemTips` and `TRANSLATOR.PtrTips` should document ownership and
  pointer passing expectations for generated wrappers.
- Rename rules should produce stable, exported Go names without losing the link
  to upstream C symbols.

Prefer adding explicit rules over relying on broad accidental generation.

## Test Organization

c-for-go may generate most functions into one `nix.go`, but tests should follow
the original upstream C package structure.

Use filenames such as:

- `nix_util_test.go`
- `nix_store_test.go`
- `nix_expr_test.go`
- `nix_fetchers_test.go`
- `nix_flake_test.go`
- `nix_main_test.go`

Tests should:

- Run inside `nix develop`.
- Prefer dummy stores or isolated temporary stores where possible.
- Check context/error behavior, not only happy paths.
- Verify ownership-sensitive paths, including free functions and returned
  strings.
- Exercise shim adapters directly through the generated Go API.

## Commit Style

Make logical, reviewable commits. Good examples:

- `feat: generate bindings for libutil`
- `feat: generate bindings for libstore`
- `feat: add libstore tests`
- `fix: libstore shim`
- `fix: generator hints for StorePath ownership`

Avoid mixing unrelated package work, generated output, shim fixes, and tests in
one large commit when they can be reviewed separately.

## Boundaries

Before adding code, ask whether it belongs in this low-level binding package. If
the feature looks like a friendly Go client, builder API, flake workflow helper,
or package-management SDK, it probably belongs in another package that consumes
these bindings.

Keep this repository focused: faithful bindings, clear shims, reproducible
generation, and tests that prove the bindings work.
