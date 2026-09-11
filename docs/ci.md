# Continuous integration

Nestplay runs separate GitHub Actions workflows for the server and Apple TV code. Documentation-only commits do not start either workflow.

## Server CI

`Server CI` runs when a pull request or push to `develop` changes `server/**`, the root `Makefile`, or the server workflow. It uses Go 1.26.x for compatibility with the latest stable Staticcheck release and checks:

- `gofmt` with no automatic edits.
- `go mod tidy` with a clean working tree.
- `go vet` and Staticcheck 2026.1.
- Tests with the race detector and an atomic coverage profile.
- A trimmed server build.
- Known reachable vulnerabilities with govulncheck 1.8.0.
- Source security findings with gosec 2.29.0.
- Go security analysis with CodeQL.

## tvOS CI

`tvOS CI` runs when a pull request or push to `develop` changes `apps/tvos/**`, `packages/HealthClient/**`, the root `Makefile`, or the tvOS workflow. It uses the GitHub `macos-26` runner and Xcode 26.6 to check:

- Official `swift-format` lint for the app and shared package.
- Shared Swift package tests.
- XcodeGen output matches the committed project.
- The app property list is valid.
- Xcode static analysis with Swift warnings treated as errors.
- An unsigned generic tvOS build.
- Swift security analysis with CodeQL.

Changes to the shared `Makefile` intentionally run both workflows because it controls both build paths. Workflow dispatch remains available for an explicit full recheck.

## Security and dependency maintenance

Actions are pinned to immutable commit SHAs. Dependabot checks GitHub Actions and the Go module once a month. The workflows grant `contents: read` and `actions: read`. CodeQL still calls the workflow-run API during analyze, so `actions: read` is required on this private repository even when SARIF is not uploaded. CodeQL runs in both jobs and fails the check if the SARIF contains findings. Results are not uploaded to GitHub code scanning: this private repository does not have GitHub Code Security. Enable that product later if alerts should appear on the Security tab.

## Local checks

From the repository root:

```sh
make test
make server
xcrun swift-format lint --recursive apps/tvos packages/HealthClient
xcodebuild analyze -project apps/tvos/Nestplay.xcodeproj -scheme Nestplay \
  -destination 'generic/platform=tvOS' CODE_SIGNING_ALLOWED=NO \
  SWIFT_TREAT_WARNINGS_AS_ERRORS=YES
```

Staticcheck, govulncheck, gosec, CodeQL, and the pinned GitHub runner provide the authoritative CI results.
