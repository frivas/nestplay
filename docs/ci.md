# Continuous integration

Run `make check` locally. That is the required gate. GitHub-hosted Actions do not run on push: macOS tvOS jobs bill at 10× and exhausted this private account's included minutes. The `CI` workflow is disabled and only accepts manual `workflow_dispatch` until the allowance resets on 1 October 2026.

## Local verification

From the repository root:

```sh
make check
```

That runs format, tests, static analysis, and builds for the Go server and tvOS app, using Go 1.26.6, Staticcheck 2026.1, govulncheck 1.8.0, and gosec 2.29.0. `make test` remains a faster unit-test subset.

Simulator UI tests (needs the server on port 8765):

```sh
xcodebuild test -project apps/tvos/Nestplay.xcodeproj -scheme Nestplay -destination 'platform=tvOS Simulator,name=Apple TV'
```

## GitHub workflow

`.github/workflows/ci.yml` is `workflow_dispatch` only. Do not re-enable it for `push` while included minutes are scarce. A macOS job of about 20 minutes consumes about 200 of the 2,000 included minutes.

If you explicitly re-enable the workflow and run it by hand, it still has two jobs:

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
