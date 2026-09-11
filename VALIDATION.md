# Validation record

Checked 11 September 2026 on an Apple M1 Pro MacBook Pro with 16 GB RAM and macOS 26.6.2. Toolchain: Go 1.27.1, Xcode 26.6 build 17F113, XcodeGen 2.46.0, tvOS 26.5 SDK and simulator runtime.

## Passed

| Check | Result |
|---|---|
| Go `go test ./...` | Passed |
| Go arm64 server build | Passed |
| Swift package tests | Five test functions passed, including five invalid-address inputs |
| Swift probe to real Go server | Returned `Connected: nestplay, API 1, ok` |
| Unsigned generic tvOS device build | Passed |
| tvOS simulator, real server available | Passed; displayed `Connected to Nestplay` |
| tvOS simulator, server unavailable | Passed; displayed `Connection needs attention` and retry remained enabled |
| App Info.plist validation | Passed |

The UI test command used the Apple TV simulator supplied by Xcode's tvOS 26.5 runtime and the actual server bound to `127.0.0.1:8765`. It executed two XCTest UI tests with zero failures. The result bundle was created under `/private/tmp`; it is intentionally not a deliverable.

The server's contract tests verify HTTP 200, JSON type, `no-store`, response fields, 404 for unrelated paths and 405 for an unsupported health method. The Swift tests verify the exact `/health` endpoint, service identity, API version, address restrictions, invalid HTTP/body handling and network failure.

## Not yet validated

- Recovery across Mac sleep/wake or hostname/IP changes.
- Long-running server lifecycle, launch at login, authentication or encryption.
- Media discovery, playback, YouTube, metadata, subtitles, downloads, disk reserve and Jellyfin. They are outside this milestone.

## Physical-device acceptance

Checked 11 September 2026 on Living Room (Apple TV HD A1625, 4th generation, tvOS 26.6), signed with a local Personal Team that is not committed. The Mac was `aquinas` at `192.168.1.134` on the same LAN. The PoC uses the fixed address `http://aquinas.local:8765` and prints tvOS traces from `POST /debug/log`.

| Check | Result |
|---|---|
| Signed install reaches Connected | Passed; no URL entry |
| Stop server and retry | Passed; **Connection needs attention**, retry remained enabled |
| Restart server and retry | Passed; `tvos info check connected` |
| Relaunch app | Passed; automatic recheck connected without typing an address |
| Firewall / focus / traces | Server logged Apple TV `GET /health` and `tvos info` lines; no focus stranding reported |

Simulator evidence is not a substitute for this record.
