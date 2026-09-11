# Continuous integration

Run `make check` locally first. GitHub CI is a path-scoped second pass on `develop`, not a pull-request gate. Documentation-only commits do not start either workflow. Use workflow dispatch only when you want an explicit recheck of `develop`.

## Local first pass

From the repository root:

```sh
make check
```

That runs the same server and tvOS checks GitHub runs, using Go 1.26.6, Staticcheck 2026.1, govulncheck 1.8.0, and gosec 2.29.0. `make test` remains a faster unit-test subset.

## Server CI

`Server CI` runs when a push to `develop` changes `server/**`, the root `Makefile`, or the server workflow. It uses Go 1.26.6 for compatibility with the latest stable Staticcheck release and checks:

- `gofmt` with no automatic edits.
- `go mod tidy` with a clean working tree.
- `go vet` and Staticcheck 2026.1.
- Tests with the race detector and an atomic coverage profile.
- A trimmed server build.
- Known reachable vulnerabilities with govulncheck 1.8.0.
- Source security findings with gosec 2.29.0.

## tvOS CI

`tvOS CI` runs when a push to `develop` changes `apps/tvos/**`, `packages/HealthClient/**`, the root `Makefile`, or the tvOS workflow. It uses the GitHub `macos-26` runner and Xcode 26.6 to check:

- Official `swift-format` lint for the app and shared package.
- Shared Swift package tests.
- XcodeGen output matches the committed project.
- The app property list is valid.
- Xcode static analysis with Swift warnings treated as errors.
- An unsigned generic tvOS build.

Changes to the shared `Makefile` intentionally run both workflows because it controls both build paths.

## Security and dependency maintenance

Actions are pinned to immutable commit SHAs. Dependabot checks GitHub Actions and the Go module once a month. The workflows use read-only `contents` access. CodeQL is not used: this private repository has no GitHub Code Security license, and the CodeQL Action cannot run against GitHub's code-scanning APIs without that product.
