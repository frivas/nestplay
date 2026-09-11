# Nestplay: personalized YouTube access and downloads
<!-- Preserved project research. -->

Research checked September 10, 2026. Documentation decision resolved; authenticated operation has not been tested. No accounts, browser stores or cookies were accessed.

## Decision

Use a replaceable YouTube.js adapter for personalized Home and search, with yt-dlp subprocesses for downloads. Go owns the durable queue, API and library. Start the PoC with explicit, local cookie import on the Mac; do not promise a seamless browser OAuth callback. A live, user-assisted authentication test remains a gate before final architecture approval.

## What current evidence establishes

**Official API:** Google's `activities.list` documentation says the former Home parameter is deprecated and requests using it return an empty list. Ordinary Google OAuth therefore does not provide the personalized Home feed Nestplay needs. [Google activities API](https://developers.google.com/youtube/v3/docs/activities/list).

**Unofficial feed:** YouTube.js calls the internal `FEwhat_to_watch` browse endpoint through `getHomeFeed()`. Its HomeFeed class exposes `getContinuation()`. Search also has continuation support. These establish interface availability, not proof that this user's personalized feed and every renderer parse correctly. [Innertube source](https://github.com/LuanRT/YouTube.js/blob/main/src/Innertube.ts), [HomeFeed API](https://ytjs.dev/api/youtubei.js/namespaces/YT/classes/HomeFeed), [Search API](https://ytjs.dev/api/youtubei.js/namespaces/YT/classes/Search).

**Authentication:** Current YouTube.js documentation recommends cookies for web clients. It describes signing into a private browser window, copying the YouTube request Cookie header, then closing the private session. OAuth now works only with the TV Innertube client; its credential update/cache events must not be mistaken for cookie refresh or proof of web Home support. There is no documented automatic browser-to-Nestplay cookie handoff. [Authentication guide](https://ytjs.dev/guide/authentication).

**Restart and expiry:** Proposed Nestplay behavior is to keep imported cookies in Mac Keychain, reconstruct the adapter after restart, and verify authenticated identity before reporting Connected. Cookie persistence preserves credentials, not their validity. If authentication fails, mark Reconnect required on the Mac and Apple TV, avoid silently substituting generic recommendations, and keep local playback available. The refresh/reconnect policy is Nestplay design, pending live validation; automatic renewal of expired cookies is not established by the documentation.

**Downloads:** yt-dlp documents that YouTube OAuth no longer works, and recommends cookies for account-required content. It also documents frequent browser cookie rotation and a private-session export procedure. Public downloads need not automatically reuse the Home account. Some formats/features require a PO-token provider; successful feed authentication does not establish successful downloading. [YouTube extractor documentation](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#youtube).

Use anonymous downloads first, then an explicit account-required state when appropriate. Keep authentication secrets off tvOS, logs and command-line arguments. Do not implement a broad browser-cookie export: yt-dlp's FAQ says the combined browser-export command writes cookies for all sites. A dedicated local import can accept YouTube cookies and normalize header versus Netscape formats without accessing unrelated browser data. [yt-dlp cookie FAQ](https://github.com/yt-dlp/yt-dlp-wiki/blob/master/FAQ.md#how-do-i-pass-cookies-to-yt-dlp).

## Runtime, licensing and maintenance

YouTube.js is MIT and runs on Node or Deno. Its authentication guide was updated May 2026; its repository and current API documentation are available, but none constitutes an upstream service-level guarantee. Keep the adapter pinned and independently replaceable. [Repository](https://github.com/LuanRT/YouTube.js), [commit history](https://github.com/LuanRT/YouTube.js/commits/main/).

yt-dlp offers a standalone universal macOS executable. Its source is Unlicense, but PyInstaller bundles are GPLv3+ combined works. Its EJS dependency includes MIT/ISC components. It supports playlist downloads, structured metadata/progress and subtitle language selection. Select Spanish subtitle tracks from available metadata, preferring supplied captions; retain language/source information and label automatic captions if used. An absent track never blocks completion. Do not assume every automatic Spanish entry is an original-language track rather than a translated one. [yt-dlp README](https://github.com/yt-dlp/yt-dlp).

Current YouTube challenge solving requires EJS plus a supported JavaScript runtime. Deno is recommended (minimum documented 2.3.0); Node is supported (minimum 22.0.0) with explicit configuration. One shared Deno runtime for YouTube.js and yt-dlp is the initial packaging candidate; verify it rather than bundling Node and Deno by default. Bundle pinned EJS assets to avoid first-run package fetching. [EJS guide](https://github.com/yt-dlp/yt-dlp/wiki/EJS).

## Adapter boundary

Proposed operations: `connectionStatus`, `home(cursor)`, `search(query,cursor)` and `videoDetails(id)`. Return normalized video records and opaque cursors. Go expands selected playlists into persistent per-video jobs and launches yt-dlp only for active jobs. Feed and download failures must remain separate; downloaded files remain usable with no Internet. Store media metadata and artwork locally when a download completes. Treat upstream cursors as short-lived; after session change or restart, refresh the feed instead of promising old cursor validity.

## Live validation gate

With the user present on the Mac, test a deliberately imported session: verify the account and recognizable personalized recommendations; load three Home pages; search and load another page; restart both adapter and server and repeat. Invalidate the imported session to verify reconnect behavior and continued local playback. Re-import and recover without resetting the library. Record whether ordinary browser use rotates the session; persistence for a few minutes cannot prove long-term stability.

Then download one public video, a short playlist and one video with Spanish captions; verify format availability, metadata and completion, and repeat with a current pinned downloader if challenges occur. Record any PO-token requirement as a separate dependency decision. No test result or RAM footprint is claimed here.

**Remaining decision:** whether manual local import is acceptable for the PoC, or whether a browser integration must be built. That is a UX decision for the user, not something research can silently choose. The clean Connect YouTube experience remains unproven.
