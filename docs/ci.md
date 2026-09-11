# Continuous integration

Run `make check` locally first. GitHub CI is a path-scoped second pass on `develop`, not a pull-request gate. Documentation-only, Makefile-only, and workflow-file-only commits do not start the workflow. Use workflow dispatch when you want an explicit recheck of both jobs on `develop`, including after editing the workflow itself.

## Local first pass

From the repository root:

```sh
make check
```

That runs the same server and tvOS checks GitHub runs, using Go 1.26.6, Staticcheck 2026.1, govulncheck 1.8.0, and gosec 2.29.0. `make test` remains a faster unit-test subset.

## GitHub workflow

One workflow, `CI`, runs when a push to `develop` changes `server/**`, `apps/tvos/**`, or `packages/HealthClient/**`. A cheap selector job compares the push to the previous `develop` tip and starts only the matching verify job:

- Server files start **Verify Go server** on Ubuntu.
- tvOS or HealthClient files start **Verify tvOS app** on macOS.
- An identical-tree force-push (metadata-only rewrite) starts the selector and skips both verify jobs.

The shared `Makefile` and the workflow file are not triggers. Change them locally, run `make check`, and use workflow dispatch if GitHub should recheck `develop`.

### Verify Go server

Uses Go 1.26.6 for compatibility with the latest stable Staticcheck release and checks:

- `gofmt` with no automatic edits.
- `go mod tidy` with a clean working tree.
- `go vet` and Staticcheck 2026.1.
- Tests with the race detector and an atomic coverage profile.
- A trimmed server build.
- Known reachable vulnerabilities with govulncheck 1.8.0.
- Source security findings with gosec 2.29.0.

### Verify tvOS app

Uses the GitHub `macos-26` runner and Xcode 26.6 to check:

- Official `swift-format` lint for the app and shared package.
- Shared Swift package tests.
- XcodeGen output matches the committed project.
- The app property list is valid.
- Xcode static analysis with Swift warnings treated as errors.
- An unsigned generic tvOS build.

## Security and dependency maintenance

Actions are pinned to immutable commit SHAs. Dependabot checks GitHub Actions and the Go module once a month. The workflow uses read-only `contents` access. CodeQL is not used: this private repository has no GitHub Code Security license, and the CodeQL Action cannot run against GitHub's code-scanning APIs without that product.
