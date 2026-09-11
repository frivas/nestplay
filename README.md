# Nestplay connection PoC

The first incremental Nestplay milestone: a Go server and a native tvOS app that checks its health over the LAN. Media features are intentionally deferred to later Wayfinder decisions.

## Run the server

The included `bin/nestplay-server` is a development build for Apple Silicon Macs. From this directory:

```sh
./bin/nestplay-server
```

It listens on port 8765. On Apple TV, enter `http://<Mac local hostname>.local:8765` or `http://<Mac LAN IP>:8765`. Find the local hostname in macOS System Settings → General → Sharing. Use the Mac's address, not `localhost`, on a physical Apple TV.

If macOS asks about incoming connections, allow the server for this LAN test. Both devices must be on the same network and the Mac must remain awake. No port forwarding is needed. This milestone exposes only a health response, with no media or account data.

Optional configuration:

```sh
./bin/nestplay-server -config server/config.example.json
./bin/nestplay-server -listen 0.0.0.0:8766
```

Stop with Control-C. To rebuild, install Go 1.23 or newer and run `make server`. No third-party Go modules are needed.

## Run the Apple TV app

1. Open `tvos/Nestplay.xcodeproj` in Xcode.
2. Select the Nestplay app target. For a physical Apple TV, select your signing team and an available bundle identifier under Signing & Capabilities. The checked-in project does not select an account or team.
3. Pair/select the Apple TV in Xcode's Devices and Simulators window. Enable Developer Mode if the device requests it, then build and run.
4. Enter the Mac address and select **Check connection**. Accept local-network access if requested.
5. Expect **Connected to Nestplay**. Stop the server and check again to see an actionable failure; restart and retry to recover.

Target: Apple TV HD A1625 with tvOS 26.6. The app's minimum deployment target is tvOS 17. The simulator uses `http://127.0.0.1:8765` to reach the Mac. Simulator success is not a claim of physical-device acceptance.

The app remembers the entered address. It displays ready, checking, connected and failure states; validates service identity/API version; and times out instead of waiting indefinitely. It uses a scoped local-network transport exception, not unrestricted HTTP access.

## Health contract

`GET /health` returns HTTP 200 and uncached JSON:

```json
{"service":"nestplay","status":"ok","api_version":1}
```

Other paths return 404 and unsupported methods return 405. This is a liveness check for the basic server, not a claim that future YouTube, storage or playback dependencies are ready.

## Structure

- `server/`: Go executable, configuration example and HTTP contract tests.
- `HealthClient/`: small Swift package used by the app and a command-line integration probe.
- `tvos/`: SwiftUI app, checked-in Xcode project and two UI checks.
- `.scratch/nestplay/` and `.scratch/nestplay-health/`: canonical local Wayfinder maps and tickets.
- `docs/research/`: architecture and component research.
- `docs/wayfinder-export/`: preserved review snapshot from before repository consolidation.

Perch, Jellyfin and YouTube packages are not dependencies of this first milestone.

## Developer checks

```sh
make test
make server
```

To regenerate the Xcode project after changing `tvos/project.yml`, run `make project` with XcodeGen installed. Regeneration is not required to open the checked-in project.

With the server running on port 8765, run the Nestplay scheme's tests in a tvOS simulator. The UI checks cover a successful real-server response and an unreachable endpoint with retry available. The portable Swift package tests malformed responses, wrong server/version, address validation and network failure.

See `VALIDATION.md` for observed results and remaining device acceptance.
