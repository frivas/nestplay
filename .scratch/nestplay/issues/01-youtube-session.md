# Establish personalized YouTube access and session lifecycle
<!-- Canonical local Wayfinder ticket. -->
Type: research
Label: wayfinder:research
Status: resolved
Assignee: youtube-research
Parent: [Decide Nestplay&apos;s proof-of-concept architecture](https://github.com/frivas/nestplay/issues/2)
Blocked by: none

## Question

What current supported or unofficial integration can retrieve authenticated Home recommendations with pagination/search and preserve/reconnect the session across server restarts? Establish sign-in handoff limitations, runtime/license, reliable download integration, and what requires live testing. Do not access user cookies or sign in.

## Answer

Resolved as documentation research on September 10, 2026. Use a replaceable YouTube.js cookie-based adapter for Home/search and yt-dlp for downloads. Ordinary OAuth does not supply personalized web Home; YouTube.js OAuth is TV-client-only and yt-dlp OAuth is unsupported. Local manual cookie import is the documented PoC path, subject to the user's UX decision. Automatic browser handoff and long-term session renewal remain unproven. Persist/reload sessions and expose reconnect state; never substitute generic recommendations silently.

Live validation pending: user-assisted account/Home pagination/search, restart, expiry/reconnect and sample downloads/captions. No account access or live validation occurred. See [YouTube session research](../../../docs/research/youtube-session.md) for evidence and acceptance checks.

Context: isolated repository `work/research/youtube`, branch `research/youtube-session`, artifact `research.md`. Commit signing unavailable through the local 1Password agent; research committed without signing.
