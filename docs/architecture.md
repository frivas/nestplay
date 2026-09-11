# Nestplay architecture boundaries

Nestplay keeps product policy in the Go server and platform behavior in the tvOS app. External services sit behind small interfaces so a provider can change without rewriting the client or catalog.

## Repository layout

- `apps/tvos` owns presentation, focus, remote interaction, playback controls and calls to the Nestplay API.
- `packages/HealthClient` owns the current reusable Swift transport and health response contract.
- `server` owns the LAN API and will grow into application, domain and adapter packages as each Wayfinder milestone requires them.
- `docs` holds durable architecture decisions, research and validation guidance.
- `.scratch` holds the canonical local Wayfinder maps and tickets.

## Dependency rules

The tvOS app depends on Swift client packages. Swift client packages know the Nestplay wire contract and do not import app views.

The Go entry point assembles dependencies. Future domain packages must not import HTTP handlers, SQLite, Jellyfin, YouTube, TMDB, OpenSubtitles, yt-dlp or FFmpeg implementations. Those implementations adapt external systems to interfaces owned by the use case that consumes them.

Nestplay owns stable identifiers and API records. Provider response models remain inside their adapters. The Apple TV app never receives YouTube.js, yt-dlp, TMDB, OpenSubtitles or Jellyfin records directly.

## Planned server modules

Add these only when a scoped milestone needs them:

- `internal/catalog` for local media identity and availability.
- `internal/downloads` for queue policy and completed-file publication.
- `internal/youtube` for personalized Home, search and session lifecycle through a replaceable adapter.
- `internal/metadata` for movie identification, original-language metadata and artwork.
- `internal/subtitles` for embedded, sidecar and provider Spanish subtitle selection.
- `internal/playback` for direct play and a replaceable Jellyfin-backed conversion adapter.
- `internal/storage` for admission control and the 20 GB free-space reserve.

Each package should expose the smallest interface its caller needs. Keep orchestration in application services, validation near the boundary, and provider-specific retry or pagination details in the adapter.

## Delivery rules

A downloaded item becomes available only after its final file is complete and atomically published. Temporary download and conversion files never appear in the library.

Every component that writes media must consult the shared storage policy. New downloads and conversions stop before free space reaches the 20 GB reserve, and the server reports the blocked state to the client.

Playback reads from completed local media. Missing Spanish subtitles remain visible as item status but never make a playable file unavailable.

Configuration and secrets stay on the Mac. YouTube sessions and provider credentials never enter the tvOS app, logs, fixtures or Git.

## Change discipline

Use the canonical Wayfinder maps before starting an increment. Keep each increment usable through the real client-server path. Add tests at stable contracts and policy boundaries. Avoid tests that repeat implementation details.
