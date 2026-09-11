# Select original-language metadata and Spanish subtitle sources
<!-- Canonical local Wayfinder ticket. -->
Type: research
Label: wayfinder:research
Status: resolved
Assignee: metadata-research
Parent: ../map.md
Blocked by: none

## Question

Which providers and matching workflow can identify folder movies, fetch original-language titles/descriptions and real covers, and acquire synchronized Spanish subtitles? Verify API access, quotas/licensing/cost, embedded subtitles, YouTube caption options, mismatch handling and absent translations. Identify proof needed, without claiming universal availability.

## Answer

Choose TMDB with per-item original-language selection, OpenSubtitles.com REST with hash-first Spanish matching across regions, and yt-dlp existing Spanish captions. [Research and remaining proof](../../../docs/research/metadata-subtitles.md). Evidence branch: `work/research/metadata`, `research/metadata-subtitles`.

Practical proof remains: authenticated subtitle quota, representative identification and synchronization, original-language gaps, YouTube captions, and Apple TV track selection. Provider selection is resolved; universal coverage and playback have not been proven.
