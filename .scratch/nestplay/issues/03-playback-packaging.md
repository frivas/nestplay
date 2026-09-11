# Choose modular playback and realistic Mac packaging
<!-- Canonical local Wayfinder ticket. -->
Type: research
Label: wayfinder:research
Status: resolved
Assignee: playback-research
Parent: ../map.md
Blocked by: none

## Question

Can a Go server compose native Jellyfin playback/transcoding while preserving replaceability and a 20 GB free-space stop across downloads and temporary files? Compare standalone direct HTTP with Jellyfin for Apple TV seeking/audio/subtitles/HDR. Determine install/update/runtime/licensing implications of one executable versus bundled helpers and a separate Jellyfin app. Define bounded-cache guarantees, live-test gaps and resource measurement plan.

## Answer

Resolved the documented architecture choice: Go coordinator plus interchangeable native Jellyfin playback adapter, with AVPlayerViewController on tvOS. Direct HTTP remains the lightest compatible-file option. A folder/app bundle or executable with configured helper paths is realistic; a dependency-free single process is not.

Jellyfin provides playback negotiation and session stopping, but duration-based segment cleanup does not enforce a disk quota. A Nestplay supervisor, admission control and an early-stop margin are needed; polling cannot guarantee zero overshoot of the 20 GB reserve on a shared disk.

[Playback and packaging research](../../../docs/research/playback-packaging.md) records primary sources, licenses, proposed boundaries and packaging. Research branch: `research/playback-packaging` in `work/research/playback`, file `research.md`. Live Apple TV compatibility, cache cancellation/seek behavior and resource measurements remain outstanding; this research resolution does not claim implementation validation.
