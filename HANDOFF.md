# Nestplay engineering handoff

Updated 11 September 2026. This is the starting point for another assistant continuing Nestplay. Read this file, `AGENTS.md`, both canonical Wayfinder maps, and the relevant next ticket before changing code or resolving a decision.

## Current outcome

The first executable milestone exists and passes its automated checks. A small Go server exposes `GET /health`; a native SwiftUI tvOS app accepts a Mac address, checks the endpoint, validates that it is Nestplay API version 1, and shows ready, checking, connected and actionable failure states. The address persists on Apple TV. The client times out, lets the user retry and does not treat arbitrary HTTP 200 responses as success.

The full PoC has not been built. YouTube, media indexing, playback, Jellyfin, metadata, subtitles and storage supervision are deliberate later increments.

Source: this directory. Canonical architecture map: https://github.com/frivas/nestplay/issues/2. Canonical health milestone map: https://github.com/frivas/nestplay/issues/1. Research is in `docs/research/`.

## Product destination and accepted requirements

Nestplay is a private, single-profile home media platform. A Mac server will provide a local movie/downloaded-video library and authenticated personalized YouTube Home/search to a custom Apple TV client. The client can select several YouTube videos and continue browsing while the server downloads them. Only complete files become playable.

Accepted behavior:

- One local profile and one YouTube account for the PoC.
- YouTube authentication occurs on the Mac. The user accepted local session import for the PoC; session secrets never go to Apple TV or chat.
- Personalized YouTube Home recommendations, continuation pages and search are required. Channel subscriptions are deferred.
- Highest quality that the target Apple TV can direct-play. Convert an incompatible movie during playback through a bounded temporary cache; preserve the source and do not retain a duplicate compatibility copy.
- Movies placed in a watched folder receive real identification, covers, and title/description in each movie's original language. Wrong matches need a correction path.
- Inspect embedded and sidecar Spanish subtitles first, then automatically seek the best-synchronized Spanish match regardless of regional variant. Enable Spanish by default. Missing subtitles never block playback and appear quietly on details/player menus, never as a persistent overlay.
- Download available Spanish YouTube captions when practical and enable them by default. Missing captions do not fail the video. Automatic translation or speech generation is outside the PoC.
- Playback controls disappear after remote inactivity and return with interaction.
- Stop download/transcode writes when free space reaches 20 GB and tell the Apple TV client. Use proactive margin because buffered writes make literal zero overshoot impossible on a shared disk. Direct reads can continue.
- Retain completed downloads until explicit deletion. Deleting through Nestplay requires confirmation and removes the source file.
- Prefer a Go coordinator and a single user-facing installation/configuration. Internal helper processes are acceptable.
- Perch was inspected as optional reference code and is not a Nestplay dependency.

Target hardware observed: server candidate is an M1 Pro MacBook Pro with 16 GB RAM and macOS 26.6.2. Playback target is Apple TV HD A1625 on tvOS 26.6. Apple documents 1080p H.264 up to 60 fps and HEVC SDR up to 1080p30; do not use a generic Apple TV 4K playback profile.

## Wayfinder state

Always use the named Wayfinder skill and the GitHub Issues tracker. Refer to tickets by linked name in user-facing text, not bare numbers. Claim a ticket before work. Resolve it with a resolution comment, close the issue, and append only a one-line linked pointer to the map. Never resolve more than one non-research ticket in a session.

The [health milestone map](https://github.com/frivas/nestplay/issues/1) has one resolved ticket:

- **Build and validate the first LAN health check** is resolved at simulator level. Its answer records exact evidence.

Its frontier is:

- **Accept the health check on the Apple TV HD**. Install a signed build on the physical A1625 and run the five checks in `VALIDATION.md`. This is the recommended immediate next task. Do not mark it resolved from simulator evidence.

The [parent architecture map](https://github.com/frivas/nestplay/issues/2) has three resolved research decisions:

- **Establish personalized YouTube access and session lifecycle** selects a replaceable YouTube.js adapter plus yt-dlp, with live account proof remaining.
- **Select original-language metadata and Spanish subtitle sources** selects TMDB, OpenSubtitles.com and existing YouTube captions, with provider/account and file validation remaining.
- **Choose modular playback and realistic Mac packaging** selects a Go coordinator with interchangeable native Jellyfin playback and AVPlayer on tvOS.

The parent map's open frontier is:

- **Accept the PoC YouTube connection experience**. The user already accepted local YouTube session import in conversation. A future session should claim this ticket, record that accepted answer, resolve only this ticket, and update the map. Do not request the decision again.
- **Prepare metadata and subtitle provider access**. This remains open and requires user-mediated TMDB/OpenSubtitles setup with secrets stored locally. Never ask the user to paste credentials into chat.

Blocked behind those are **Validate account access and playback on the target devices**, then **Confirm the Nestplay PoC stack and packaging tradeoffs**. The validation ticket also depends on playback research, which is resolved, and physical-device fixtures. `docs/research/validation-plan.md` describes bounded proofs.

Do not convert the entire destination into implementation tickets now. Wayfinder keeps later work in fog until the next decision makes it precise. For execution milestones, add an explicit execution exception to the map Notes, as done for the health milestone.

## Candidate architecture

The research-supported candidate is:

1. Go owns the public LAN API, SQLite catalog, download queue, stable media IDs, provider orchestration and disk policy.
2. A replaceable YouTube.js adapter, initially using Deno, obtains authenticated Home/search and opaque continuation cursors.
3. yt-dlp handles downloads and available captions; Deno/EJS handles current YouTube extraction challenges; FFmpeg/ffprobe merges, probes and converts.
4. TMDB supplies movie identity, original-language metadata and artwork. Query language per movie; `original_title` alone does not provide an original-language overview.
5. OpenSubtitles.com supplies Spanish subtitles using embedded/sidecar-first, then hash-first/release-aware matching. The provider's free quota documentation conflicts, so record the live allowance.
6. Native Jellyfin is initially behind a Nestplay playback adapter for direct play/remux/transcode. Nestplay's public model must not expose Jellyfin item schemas so another backend can be added later.
7. The Apple TV app uses SwiftUI and AVPlayerViewController for native controls and subtitle/audio menus.

Detailed evidence is in `docs/research/architecture-findings.md`, `youtube-session.md`, `metadata-subtitles.md` and `playback-packaging.md`.

## Implemented source

`server/main.go` uses only Go's standard library. Default listen address is `0.0.0.0:8765`; `-listen` overrides an optional JSON `-config`. It supplies read-header/write/idle timeouts, a small maximum header size, graceful signal shutdown, exact method/path routing and an uncached JSON health response.

`packages/HealthClient/Sources/HealthClient/HealthClient.swift` is the reusable network boundary. It accepts only HTTP(S) base URLs with no credentials, query, fragment or non-root path, appends `/health`, uses an ephemeral no-cache session and validates the response contract. Its command-line `health-probe` makes end-to-end checks possible without the UI.

`apps/tvos/Nestplay/ConnectionView.swift` is the only screen. It persists the submitted address with `AppStorage`, displays connection state and takes `--server-url` for automated UI checks. `Info.plist` explains local-network access and permits local networking without a broad arbitrary-load exception.

`apps/tvos/project.yml` is the source for the generated checked-in Xcode project. Use XcodeGen after changing project structure. Signing team and user-specific identifiers are intentionally absent.

## Verification already completed

See `VALIDATION.md` for the durable record. On 11 September 2026:

- Go tests and arm64 build passed.
- Five Swift test functions passed, including parameterized invalid inputs.
- The same Swift client reached the real Go process and validated the response.
- The unsigned generic tvOS build succeeded.
- Two tvOS 26.5 simulator UI tests passed: real-server success and unreachable-server failure with retry enabled.
- The app plist passed validation.

The tvOS 26.5 simulator runtime was installed into Xcode during this work. The simulator name used was `Apple TV`, device ID `5A226AE4-16EB-4215-86C6-CEC9279D6BDD`; device IDs can change and must not become configuration.

Routine commands from this directory:

```sh
make test
make server
.build/bin/nestplay-server
cd apps/tvos && xcodegen generate
```

The exact simulator test requires the server running on `127.0.0.1:8765`, then:

```sh
xcodebuild -project apps/tvos/Nestplay.xcodeproj -scheme Nestplay \
  -destination 'platform=tvOS Simulator,name=Apple TV' \
  -derivedDataPath /private/tmp/nestplay-sim-derived \
  CODE_SIGNING_ALLOWED=NO test
```

The user's global instruction requires prefixing shell commands with `rtk`; use it in agent sessions. The plain commands above are end-user README-style commands.

## Physical-device continuation

The Mac local hostname observed was `aquinas`; try `http://aquinas.local:8765`. Start the server without the localhost-only test override so it binds to the LAN. At handoff, Xcode listed only simulators and generic tvOS destinations; the physical A1625 was not paired or visible. Pair the Apple TV in Xcode, choose the user's development team, change the bundle identifier only if Xcode requires it, install, and accept local-network permission. Never commit team IDs, provisioning profiles or credentials.

Run the five acceptance checks in `VALIDATION.md`: connect; stop server and see failure; restart and recover; restart app and retain address; record firewall/focus/timing issues. If `.local` fails, find the active LAN address locally. Do not add automatic discovery until this manual path has been observed.

## Known limitations and risks

- `/health` is unauthenticated plain HTTP. It reveals only fixed service/version status. This is acceptable for this narrow LAN proof, not for future account or library endpoints.
- Binding `0.0.0.0` reaches every Mac interface. The later security/pairing decision must define LAN identity and access; “no port forwarding” is not access control.
- The server binary is generated under `.build/bin/` and is never committed. Build it from source with `make server`.
- No project license has been chosen. Do not distribute beyond the private PoC until licensing is decided.
- Simulator success does not prove physical Apple TV networking, signing, firewall behavior or remote focus.
- Local session-cookie import for YouTube is accepted but not implemented. YouTube's official API does not provide personalized Home. Cookie validity and reconnection need live proof.
- A successful feed session does not guarantee yt-dlp downloads. Account-required videos and PO-token handling are separate checks.
- TMDB/OpenSubtitles are external services with accounts, terms and quotas. “Spanish subtitles always” means automatically try and show a quiet missing state; availability and synchronization cannot be guaranteed.
- Jellyfin 12 API behavior must be pinned and verified. Its segment cleanup is duration-based and does not enforce the 20 GB reserve.
- Polling available disk cannot guarantee zero overshoot because of buffered writers and other Mac applications. Use a measured safety margin and supervise every writable volume and child process.
- Packaging will be one user-facing installation with multiple internal executables, not literally one process. Exact yt-dlp/FFmpeg/Jellyfin redistribution obligations depend on the chosen artifacts.

## Recommended route after physical acceptance

1. Resolve **Accept the health check on the Apple TV HD** with physical evidence and any address-screen changes.
2. In a separate Wayfinder session, resolve **Accept the PoC YouTube connection experience** from the user's already-given acceptance.
3. Work **Prepare metadata and subtitle provider access** without exposing secrets. This can proceed alongside planning the next execution milestone only if Wayfinder dependencies permit it.
4. Create the next bounded execution map for local library discovery and read-only catalog display before deletion or transcoding. This exercises the server API and TV browsing model without YouTube authentication risk.
5. Then validate YouTube Home/search/session persistence as a separate adapter proof. Keep feed failure independent from the local library.
6. Add queued downloads only after feed records have stable IDs and continuation behavior is observed.
7. Add playback through the adapter, beginning with direct-play fixtures on the A1625 and then one incompatible conversion case through Jellyfin.
8. Implement the storage supervisor before enabling simultaneous downloads and conversion.
9. Run the full account, metadata, subtitle, playback, disk and resource plan. Confirm the final stack only after recording those results.

My recommendation is to keep each milestone vertically testable through the TV: health, local catalog, YouTube browse, selection/queue, completed download, direct playback, then conversion and subtitle refinement. That exposes integration mistakes early while Wayfinder keeps larger decisions reviewable.

## Repository and handoff integrity

This repository lives at `/Users/francisco/Downloads/Projects/nestplay` on branch `develop`. Its `origin` remote is the private GitHub repository `frivas/nestplay`. The initial commit should contain source, generated Xcode project, tests, validation, Wayfinder-aware instructions and this handoff. If the tree is not clean, inspect every change before continuing; never discard user work.

Generated binaries, Swift/Xcode build products and Xcode user data remain ignored.
