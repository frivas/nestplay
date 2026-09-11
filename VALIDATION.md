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

- Installation, signing and LAN traffic on the physical Apple TV HD A1625 with tvOS 26.6.
- macOS firewall prompts, network isolation or guest Wi-Fi behavior.
- Recovery across Mac sleep/wake or hostname/IP changes.
- Long-running server lifecycle, launch at login, authentication or encryption.
- Media discovery, playback, YouTube, metadata, subtitles, downloads, disk reserve and Jellyfin. They are outside this milestone.

## Physical-device acceptance

Start the server and install the app by following README.md. Use `http://aquinas.local:8765` as the first hostname candidate; `aquinas` was the Mac's local hostname when this record was written. Then:

1. Confirm the app reaches `Connected to Nestplay`.
2. Stop the server and retry; confirm the failure is readable and does not strand focus.
3. Restart the server and retry; confirm recovery without reinstalling or re-entering the address.
4. Restart the Apple TV app; confirm it remembers the address and can check again.
5. Record any firewall prompt, timing or remote-control usability issue in the open Wayfinder ticket.

Passing those steps resolves physical-device acceptance. Simulator evidence must not be presented as a substitute.

At handoff time, Xcode listed only generic tvOS destinations and simulators. The physical A1625 was not paired or visible to this Mac's Xcode session.
