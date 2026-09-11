# Nestplay architecture validation plan
<!-- Preserved project research. -->

These are proposed bounded proofs, not completed tests or a production implementation plan.

## Known scope

One profile, one YouTube account, Mac-side YouTube access, custom Apple TV browsing and playback. Only completed videos play. Existing movie files remain the source; incompatible playback uses transient conversion. Original-language movie metadata and best-matching Spanish subtitles are separate requirements. Missing subtitles do not block playback. No automatic media deletion.

The host inspected during planning is an Apple M1 Pro MacBook Pro with 16 GB RAM and macOS 26.6.2. Confirm that this is the intended server. The user confirmed Apple TV A1625 with tvOS 26.6. This is Apple TV HD; baseline direct-play proof should target supported 1080p SDR formats, not 4K/HDR output. Include a higher-resolution or HDR source only to verify conversion to a supported output. [Apple specifications](https://support.apple.com/en-us/111928).

## YouTube proof

Use a user-mediated sign-in, without asking the user to paste cookies or credentials into chat. Fetch authenticated Home, load a continuation page and perform a search. Restart the test service and repeat. Exercise an expired/revoked session and show reconnection without losing local library access. Select several results and verify stable IDs, duplicate handling and independence from browsing. Download one authorized test item completely and import available Spanish captions. Document account-specific failures rather than inferring reliability from a successful anonymous search.

## Metadata and subtitle proof

Use user-selected fixtures including an original-language title different from its English translation, an ambiguous remake, an obscure movie with missing metadata, and alternate cuts/releases. Demonstrate identification, original-language overview, artwork and a correction route. Find existing embedded and sidecar Spanish tracks before downloading another. Check matching and synchronization, not merely language tags. Confirm missing subtitles remain a quiet details/player-menu state. Include one YouTube video with Spanish captions and one without.

## Playback proof

On the actual Apple TV, test a completed YouTube download, a typical movie and an incompatible source. Include seeking forward/backward, resume, audio track selection, Spanish subtitles, absent subtitles and controls hiding while idle. Include HDR/multichannel audio only where the user's fixtures and equipment support those cases. Record whether each test direct-plays, remuxes or transcodes. Ensure no permanent compatibility duplicate is retained.

## Storage proof

Do not fill the user's real disk to run this test. Use an isolated quota-limited test volume or simulated free-space inputs for the controller, followed by a safe bounded integration test. Test threshold crossing, concurrent downloads/transcoding, restarting while blocked, cancellation of child writers and cleanup. Distinguish decimal 20 GB from binary 20 GiB in implementation configuration. Choose a conservative early-stop allowance for buffered writes. Verify the client receives a storage-blocked event and direct playback still works. Check the actual volume backing every writable path, including temporary files.

## Resource proof

Measure ten minutes idle, direct playback with seeking, download alongside playback, metadata scanning and one live transcode. Include all helper processes and the media backend. Record memory, CPU, power/energy where available, disk growth, start latency and cleanup time. No resource budget has yet been accepted and no performance result is claimed.

## Packaging proof

Validate on a clean user environment without Homebrew assumptions: start from the proposed bundle/config, resolve every helper, restart, locate logs without secrets, update safely and recover from missing helper or port collision. Document every executable and license. A single entry command is distinct from literally one operating-system process or one physical binary.

## Decision gate

Research can select candidate integrations and identify constraints. Final stack confirmation waits for account and device evidence and explicit acceptance of any packaging/provider tradeoffs. Unproven behavior remains open on the Wayfinder map.
