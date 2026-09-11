# Nestplay connection PoC

The first incremental Nestplay milestone: a Go server and a native tvOS app that checks its health over the LAN. For this PoC the Apple TV uses the fixed address `http://aquinas.local:8765` and checks it on launch. It also sends debug traces to `POST /debug/log`, which the server prints. Media features are intentionally deferred to later Wayfinder decisions.

## Run the server

Build and run the Go server from the repository root:

```sh
make run
```

It listens on port 8765. The tvOS PoC checks `http://aquinas.local:8765` by itself. Use the Mac's address, not `localhost`, on a physical Apple TV. If `.local` fails, the fallback is the Mac LAN IP on port 8765.

If macOS asks about incoming connections, allow the server for this LAN test. Both devices must be on the same network and the Mac must remain awake. No port forwarding is needed. This milestone exposes health, a read-only local catalog, and YouTube Home/search. Downloads and playback are still deferred.

Optional configuration:

```sh
.build/bin/nestplay-server -config server/config.example.json
.build/bin/nestplay-server -listen 0.0.0.0:8766
```

Stop with Control-C. To rebuild, install Go 1.23 or newer and run `make server`. No third-party Go modules are needed.

## Run the Apple TV app

1. Open `apps/tvos/Nestplay.xcodeproj` in Xcode.
2. Select the Nestplay app target. For a physical Apple TV, select your signing team and an available bundle identifier under Signing & Capabilities. The checked-in project does not select an account or team.
3. Pair/select the Apple TV in Xcode's Devices and Simulators window. Enable Developer Mode if the device requests it, then build and run.
4. Run the app. It checks `http://aquinas.local:8765` automatically. Accept local-network access if requested. Use **Check connection** to retry. UI tests may still pass `--server-url`.
5. Expect **Connected to Nestplay**. Stop the server and check again to see an actionable failure; restart and retry to recover. Watch the Mac terminal for `GET /health` and `tvos …` debug lines.

Target: Apple TV HD A1625 with tvOS 26.6. The app's minimum deployment target is tvOS 17. The simulator uses `http://127.0.0.1:8765` to reach the Mac. Simulator success is not a claim of physical-device acceptance.

The PoC address is fixed. The app displays ready, checking, connected and failure states; validates service identity/API version; and times out instead of waiting indefinitely. It uses a scoped local-network transport exception, not unrestricted HTTP access.

## Health contract

`GET /health` returns HTTP 200 and uncached JSON:

```json
{"service":"nestplay","status":"ok","api_version":1}
```

`POST /debug/log` accepts uncached JSON `{"source":"tvos","level":"info","message":"..."}` and prints it on the Mac.

`GET /library` returns HTTP 200 and uncached JSON `{ "items": [ { "id", "title", "filename", "artwork" } ] }`. `artwork` is a Nestplay path such as `/library/{id}/artwork`, never a filesystem path. `GET /library/{id}/artwork` returns a 16:9 JPEG still from the watched-folder file, or a designed placeholder if a still cannot be taken. An empty or missing watched folder is a valid empty list. Set `library_root` in the config file or pass `-library-root`.

`GET /youtube/status`, `GET /youtube/home`, and `GET /youtube/search?q=` return uncached JSON `{ "state", "message?", "items", "cursor?" }`. `state` is `connected`, `reconnect_required`, or `unavailable`. Items are landscape recommended videos `{ "id", "title", "artwork" }` with Nestplay artwork paths `/youtube/{id}/artwork`. Pass `cursor` for the next page. The Apple TV never receives cookies, YouTube.js records, or filesystem paths. If the YouTube session is missing or rejected, Nestplay says reconnect is required and does not invent generic recommendations; the local library stays available.

YouTube cookies stay on the Mac. Put a YouTube `Cookie` header or a YouTube-only Netscape cookie file at `~/Library/Application Support/nestplay/youtube-cookies.txt` (or pass `-youtube-cookies`) and restart the server. Do not paste cookies into chat, Git, or the tvOS app.

Other paths return 404 and unsupported methods return 405. This is a liveness check plus a read-only filename catalog with posters and a YouTube Home/search adapter, not a claim that downloads or playback are ready. This debug ingest is unauthenticated LAN-only PoC tracing.

## Structure

- `apps/tvos/`: SwiftUI tvOS app, checked-in Xcode project and UI checks.
- `server/`: Go service, configuration example and HTTP contract tests.
- `packages/HealthClient/`: reusable Swift network client and command-line integration probe.
- `docs/architecture.md`: ownership boundaries and dependency rules for future increments.
- `docs/research/`: architecture and component research.

Perch, Jellyfin and yt-dlp are not dependencies of this milestone. YouTube.js runs through a replaceable Deno adapter on the Mac.

## Developer checks

```sh
make check
make server
```

`make check` is the required verification (format, tests, static analysis, and builds). GitHub-hosted Actions do not run on push. See [continuous integration](docs/ci.md).

To regenerate the Xcode project after changing `apps/tvos/project.yml`, run `make project` with XcodeGen installed. Regeneration is not required to open the checked-in project.

With the server running on port 8765, run the Nestplay scheme's tests in a tvOS simulator. The UI checks cover a successful real-server response and an unreachable endpoint with retry available. The portable Swift package tests malformed responses, wrong server/version, address validation and network failure.

See `VALIDATION.md` for observed results and remaining device acceptance.
