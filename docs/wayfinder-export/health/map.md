# Prove Apple TV can reach the Nestplay Go server
<!-- Exported Wayfinder snapshot. -->
Label: wayfinder:map

## Destination

Build a basic Go server with GET /health and a tvOS app that checks it over the local network, with observable connected, checking and unreachable states.

## Notes

User explicitly requested this incremental build before YouTube integration. Execution is authorized for this milestone. Use Wayfinder for every subsequent increment. Target Apple TV HD A1625, tvOS 26.6. Go and SwiftUI; no Perch dependency. Manual server URL is the simplest reversible setup choice; automatic discovery can follow. Source deliverable: this repository.

## Decisions so far

- [Build and validate the first LAN health check](issues/01-health-proof.md) — The Go-to-tvOS contract, device build and simulator success/failure flows work; physical Apple TV acceptance remains open.

## Not yet specified

Subsequent media and account work remains in the parent architecture map.

## Out of scope

YouTube authentication, media playback, downloads, Jellyfin installation, metadata, subtitles and automatic discovery during this milestone.
