# Nestplay connection PoC

The first incremental Nestplay milestone: a Go server and a native tvOS app that checks its health over the LAN. For this PoC the Apple TV uses the fixed address `http://aquinas.local:8765` and checks it on launch. It also sends debug traces to `POST /debug/log`, which the server prints. Media features are intentionally deferred to later Wayfinder decisions.

## Run the server

Build and run the Go server from the repository root:

```sh
make run
```

It listens on port 8765. The tvOS PoC checks `http://aquinas.local:8765` by itself. Use the Mac's address, not `localhost`, on a physical Apple TV. If `.local` fails, the fallback is the Mac LAN IP on port 8765.

If macOS asks about incoming connections, allow the server for this LAN test. Both devices must be on the same network and the Mac must remain awake. No port forwarding is needed. This milestone exposes health, a live local catalog, YouTube Home/search, direct play of completed local-library files, and a YouTube download queue that publishes only Living Room–playable files.

Optional configuration:

```sh
.build/bin/nestplay-server -config server/config.example.json
.build/bin/nestplay-server -listen 0.0.0.0:8766
.build/bin/nestplay-server -yt-dlp /opt/homebrew/Caskroom/miniconda/base/envs/ytdlp/bin/yt-dlp -ffmpeg /opt/homebrew/bin
```

yt-dlp defaults to that conda env binary; ffmpeg and ffprobe default to `/opt/homebrew/bin`. Pass `-yt-dlp` / `-ffmpeg` or set `yt_dlp` and `ffmpeg` in the config file to override.

Stop with Control-C. To rebuild, install Go 1.25 or newer and run `make server`. The catalog watcher uses `github.com/fsnotify/fsnotify` and persists the snapshot with `modernc.org/sqlite`.

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

`GET /library` returns HTTP 200 and uncached JSON `{ "items": [ { "id", "title", "filename", "collection", "artwork" } ] }`. `artwork` is a Nestplay path such as `/library/{id}/artwork`, never a filesystem path. `collection` is the slash-normalized parent folder relative to the watched folder (empty at the root). Apple TV browses that tree (Mateo → PJ Masks) and keeps Next/Previous inside a folder; it is not a Mac path. `GET /library/{id}/artwork` returns a 16:9 JPEG still from the watched-folder file, or a designed placeholder if a still cannot be taken. `GET /library/{id}/media` streams that file with HTTP Range via `ServeContent` so Apple TV can seek. The handler clears the write deadline so JSON `WriteTimeout` does not abort a long body. An empty or missing watched folder is a valid empty list. Set `library_root` in the config file or pass `-library-root`.

`GET /library/events` is a Server-Sent Events stream. Nestplay writes `data: {"changed":true}` when the in-memory catalog snapshot changes; the Apple TV then reloads `GET /library` without reconnecting. Watched-folder paths never appear on this stream. Files already on disk are admitted at startup; new or growing files wait until their size is stable. Hidden names and `*.part` files stay out of the snapshot.

`GET /youtube/status`, `GET /youtube/home`, and `GET /youtube/search?q=` return uncached JSON `{ "state", "message?", "items", "cursor?" }`. `state` is `connected`, `reconnect_required`, or `unavailable`. Items are landscape recommended videos `{ "id", "title", "artwork" }` with Nestplay artwork paths `/youtube/{id}/artwork`. Pass `cursor` for the next page. The Apple TV never receives cookies, YouTube.js records, or filesystem paths. If the YouTube session is missing or rejected, Nestplay says reconnect is required and does not invent generic recommendations; the local library stays available.

YouTube cookies stay on the Mac. From the repo root, with **Chrome signed in to youtube.com** (the browser whose Home shows recommendations):

```sh
scripts/import-youtube-cookies.sh
```

That writes a YouTube-only Netscape file to `~/Library/Application Support/nestplay/youtube-cookies.txt`. Pass `safari`, `brave`, or `chrome:Profile 1` as the first argument if needed. Quit the browser if the import cannot open the cookie database. Then restart the Nestplay server. On the Apple TV, **Refresh** loads YouTube Home from the Mac; that shelf stays until you Refresh again. Do not paste cookies into chat, Git, or the tvOS app. You can still place a Cookie header in that same path by hand, or pass `-youtube-cookies`.

`GET /queue` and `POST /queue` are the download queue. `POST /queue` accepts uncached JSON `{ "items": [ { "id", "title" } ] }` of YouTube video ids. Invalid ids are ignored. `GET /queue` returns `{ "items": [ { "id", "title", "state", "message?" } ] }`. `state` is `queued`, `downloading`, `converting`, `waiting_space`, `too_low`, `failed`, or `complete`. Watched-folder paths never appear. yt-dlp fetches at least 720p (preferring 1080p H.264 when YouTube has it); ffmpeg writes 1080p-or-source H.264+AAC MP4 into the watched folder (`-library-root`). Scratch files stay in `.nestplay-work` under that folder (skipped by the catalog) and are deleted after each job; leftover `.mp4.part` files and the work directory are removed when the server starts. Below 720p is not downloaded. Unplayable intermediates stay out of `GET /library`. Local library listings omit files Living Room cannot draw until ffmpeg replaces them in place.

Other paths return 404 and unsupported methods return 405. This is a liveness check plus a live local catalog with posters, direct play of completed files, a YouTube Home/search adapter, and a download queue that only admits Apple TV–playable files. This debug ingest is unauthenticated LAN-only PoC tracing.

## Structure

- `apps/tvos/`: SwiftUI tvOS app, checked-in Xcode project and UI checks.
- `server/`: Go service, configuration example and HTTP contract tests.
- `packages/HealthClient/`: reusable Swift network client and command-line integration probe.
- `docs/architecture.md`: ownership boundaries and dependency rules for future increments.
- `docs/research/`: architecture and component research.

Perch, Jellyfin are not dependencies of this milestone. yt-dlp and ffmpeg run as off-process writers on the Mac. YouTube.js runs through a replaceable Deno adapter.

## Developer checks

```sh
make check
make server
```

`make check` is the required verification (format, tests, static analysis, and builds). GitHub-hosted Actions do not run on push. See [continuous integration](docs/ci.md).

To regenerate the Xcode project after changing `apps/tvos/project.yml`, run `make project` with XcodeGen installed. Regeneration is not required to open the checked-in project.

With the server running on port 8765, run the Nestplay scheme's tests in a tvOS simulator. The UI checks cover a successful real-server response and an unreachable endpoint with retry available. The portable Swift package tests malformed responses, wrong server/version, address validation and network failure.

See `VALIDATION.md` for observed results and remaining device acceptance.
