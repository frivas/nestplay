# Decide Nestplay's proof-of-concept architecture
<!-- Exported Wayfinder snapshot. -->
Label: wayfinder:map

## Destination

Choose an implementation-ready architecture for a single-profile Nestplay PoC: Go-led Mac server, custom Apple TV client, local movies and downloaded videos, and personalized YouTube browsing. Identify any practical validation that must precede a final stack decision.

## Notes

Use wayfinder, research, grilling and domain-modeling. Local Markdown is the tracker because no Nestplay repository/tracker was specified. Research branches live in isolated repositories under work/research; findings are linked from resolution comments. Perch is an optional reference, not a dependency. Execution exception authorized by the user: first build an incremental tvOS-to-Go health-check PoC, tracked by its own child milestone map, before YouTube/account experiments. This does not authorize implementing the entire media platform at once. Every later increment remains Wayfinder-tracked.

Health milestone status: simulator-complete in this repository; physical Apple TV acceptance remains open in `../../.scratch/nestplay-health/02-physical-device-acceptance.md`.

User requirements: one profile and one YouTube account; server-side authentication, personalized Home pagination and search; select multiple downloads without leaving browsing; completed files only become playable. Preserve highest direct-play quality; incompatible movies use transient on-demand conversion, not permanent duplicates. Playback backend must allow Jellyfin replacement/addition. Original-language movie titles/descriptions and real covers. Automatically obtain best-matching Spanish subtitles regardless of region and enable by default; also obtain available Spanish YouTube captions when practical. Missing subtitles never block playback and notices must not persist over video; controls hide with inactivity. Downloads and transcoding stop writing at 20 GB free; client is notified. No automatic media deletion; confirmed deletion removes source file. Prefer a self-contained/configurable executable, with Go preferred.

## Decisions so far

- [Establish personalized YouTube access and session lifecycle](issues/01-youtube-session.md) — YouTube.js plus yt-dlp is the documented candidate; cookie handoff and live-account reliability remain separate gates.
- [Select original-language metadata and Spanish subtitle sources](issues/02-metadata-subtitles.md) — TMDB, OpenSubtitles.com and existing YouTube captions cover the provider roles, with per-item language policy and quotas requiring validation.
- [Choose modular playback and realistic Mac packaging](issues/03-playback-packaging.md) — Go plus an interchangeable native Jellyfin backend is the candidate; bundle/config packaging and proactive disk-write supervision are required.

## Not yet specified

The user approved local Mac session import for the YouTube PoC. Record this prior conversational acceptance when next resolving its decision ticket; it is not a blocker for the independent health-check milestone.

Refinements to client/server contracts and recovery behavior that emerge from provider and playback investigations. Exact validation fixtures and acceptance thresholds depend on the user's devices and available media.

## Out of scope

Production implementation, multiple profiles/accounts, channel subscriptions, incomplete-download playback, permanent compatibility copies, public access, and automatic subtitle generation/translation.
