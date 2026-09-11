# Nestplay playback and Mac packaging decision
<!-- Preserved project research. -->

Research checked 10 September 2026. This resolves the documented architecture choice, not live compatibility or resource validation. The host is an M1 Pro MacBook Pro with 16 GB RAM and macOS 26.6.2, as observed by the coordinating session. Apple TV model remains to be confirmed.

## Recommendation

Use **Go as the Nestplay coordinator, native Jellyfin as the initial interchangeable playback service, and a custom SwiftUI/tvOS client using AVPlayerViewController**. Distribute a self-contained folder or app bundle with a configuration file; initially allow a separately installed native Jellyfin. Do not promise a literal single process or dependency-free executable.

Keep a direct-HTTP adapter as the lightest alternative for compatible files. It cannot fulfill arbitrary movie conversion by itself. The user's requirement for arbitrary folder imports, Spanish subtitles and on-demand conversion makes Jellyfin worth its extra resident process. This is architectural inference, not a measured memory ranking.

## Why Jellyfin fits

Jellyfin negotiates playback using client capabilities. Its modes progress from unchanged direct play to container remux, audio conversion, and full video conversion. A mismatched subtitle format can also cause expensive video burn-in. Consequently, selecting by MP4 extension alone is insufficient: probe container, video profile, audio and subtitle tracks, then negotiate with the actual player. [Playback modes](https://jellyfin.org/docs/general/post-install/transcoding/), [codec support](https://jellyfin.org/docs/general/clients/codec-support/).

Native macOS supports VideoToolbox hardware acceleration and Metal/native tone mapping. The documented Apple Silicon native installation currently requires macOS 14; the observed Mac exceeds it. Use Jellyfin's supported FFmpeg build rather than assuming an arbitrary FFmpeg binary behaves identically. [Mac installation](https://jellyfin.org/docs/general/installation/macos/), [Apple acceleration](https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/apple/), [FFmpeg guidance](https://jellyfin.org/docs/general/post-install/transcoding/hardware-acceleration/).

The release endpoint currently resolves to Jellyfin 12.0. Source examined below is the moving master branch, so verify these contracts against the pinned release before implementation; this report does not certify all master features as shipped in 12.0. [Release](https://github.com/jellyfin/jellyfin/releases/tag/v12.0).

## Keep the integration replaceable

Nestplay owns stable media IDs, download state, storage policy and client-facing responses. A playback adapter returns a session ID, URL, playback mode, track choices and supported seeking behavior. Separate operations prepare playback, report progress and stop the session. Store the Nestplay-to-Jellyfin item mapping internally; do not make Jellyfin item schemas the public app model. This is the proposed boundary, not an existing Nestplay implementation.

Jellyfin exposes POST `Items/{itemId}/PlaybackInfo` with a device profile, bitrate and audio/subtitle selections. Current source has POST `Sessions/Playing`, `/Progress` and `/Stopped`; the stopped handler kills transcoding jobs associated with the supplied play-session/device before recording the stop. Stopping the player UI alone is therefore insufficient: Go must explicitly end the backend session and reject subsequent requests when storage is blocked. [Playback API source](https://raw.githubusercontent.com/jellyfin/jellyfin/master/Jellyfin.Api/Controllers/MediaInfoController.cs), [session API source](https://raw.githubusercontent.com/jellyfin/jellyfin/master/Jellyfin.Api/Controllers/PlaystateController.cs).

For simple delivery, Go's `ServeContent` handles byte ranges, supporting HTTP seeking for compatible files. It supplies neither codec conversion nor subtitle discovery. [Go HTTP documentation](https://pkg.go.dev/net/http#ServeContent).

Keep Jellyfin bound locally where practical, with Nestplay providing authenticated LAN access to playback resources. HLS playlists and segment URLs require deliberate routing; proxy behavior is part of the playback proof. Disable remote exposure and automatic port mapping. One household profile still needs a device-pairing credential.

## Apple TV player choice

AVPlayerViewController provides the standard tvOS playback interface, subtitle/audio menus and remote interactions. It is the best initial match for disappearing controls. Supply supported selectable subtitles through the playback presentation rather than assuming an arbitrary sidecar SRT URL becomes a native menu item. [Apple player](https://developer.apple.com/documentation/avkit/avplayerviewcontroller), [tvOS customization](https://developer.apple.com/documentation/avkit/customizing-the-tvos-playback-experience).

VLCKit is an alternative tvOS playback engine and is LGPL-2.1-or-later. Evaluate it only if measured format coverage materially reduces transcoding; it adds another player integration and licensing surface. Do not assume it preserves the same native UI or every HDR/audio capability. [VideoLAN project](https://github.com/videolan/vlckit).

Test HDR separately: Jellyfin documents HDR-to-SDR when video transcoding, not preservation of HDR through that conversion. A compatible direct-play path is preferable for original HDR quality. [HDR limitation](https://jellyfin.org/docs/general/post-install/transcoding/).

## The 20 GB reserve needs an explicit supervisor

Jellyfin exposes a transcode directory, throttling and segment retention/deletion settings. These are duration-based controls, not a documented hard byte quota or a 20 GB free-space guarantee. Jellyfin warns that transcoding storage can approach the source size. [Encoding settings](https://raw.githubusercontent.com/jellyfin/jellyfin/master/MediaBrowser.Model/Configuration/EncodingOptions.cs), [storage guidance](https://jellyfin.org/docs/general/administration/storage/).

Proposed policy: define 20 GB as 20,000,000,000 bytes available to the process on each affected volume. Monitor downloads, merge outputs, transcode segments, artwork, subtitles and logs. Serialize new work admissions; refuse new writes at the threshold; stop active child writers and playback conversion; notify the client; retain manual resume. Direct reads continue. The supervisor must cover Jellyfin requests too, so Apple TV cannot restart conversion behind it.

Stop proactively above 20 GB using a safety margin sized from measured write rate and stop latency. Buffered writes, polling intervals and unrelated Mac applications mean **a literal guarantee that free space never crosses 20 GB is not achievable through polling alone**. State this honestly. A filesystem quota can bound a dedicated cache, but does not reserve free space against other applications on the shared disk. Test cancellation acknowledgement and orphan-process cleanup.

Enable segment deletion and throttling; monitor cache bytes independently; clean abandoned sessions after restart without deleting active segments. Seek backward after segment removal must regenerate data correctly. Also budget temporary download merging: separate audio/video plus merged output can temporarily coexist even though no duplicate is retained permanently.

## Packaging and licenses

A Go binary can supervise sibling executables. Bundle versioned yt-dlp, FFmpeg/ffprobe and Deno plus the YouTube.js helper; start helpers only when needed. YouTube.js supports Deno and is MIT; Deno is MIT. yt-dlp offers standalone macOS builds, recommends FFmpeg for merging, and currently requires a supported JavaScript runtime for full YouTube support. [YouTube.js](https://github.com/LuanRT/YouTube.js), [Deno](https://github.com/denoland/deno), [yt-dlp](https://github.com/yt-dlp/yt-dlp).

Do not describe the whole bundle as Unlicense: yt-dlp source is Unlicense but bundled third-party components have their own terms. Jellyfin uses GPLv2; FFmpeg is LGPL by default with optional GPL components changing the resulting build's license. Preserve notices and source obligations for the exact redistributed artifacts. Separate-process composition is an engineering boundary, not a blanket licensing conclusion. [Jellyfin license](https://github.com/jellyfin/jellyfin/blob/master/LICENSE), [FFmpeg licensing](https://ffmpeg.org/legal.html).

For the PoC, a Go executable plus config and explicit helper paths is sufficient. Later, a signed app bundle can present one install while retaining multiple internal executables. Keep updates atomic and versions pinned; never replace a running transcoder. Packaging must be tested on a clean Mac without developer tools.

## Remaining live proof

Measure combined process memory, idle CPU, startup time, playback start/seek latency and cache growth for direct play, audio-only conversion, subtitle burn-in and HDR conversion. Repeat with a simultaneous download; exercise sleep/wake, restart and reserve-trigger cancellation on an isolated test volume. Use representative files including MP4/H.264, HEVC/HDR, MKV, multichannel audio, SRT and image subtitles. No benchmark, installation or playback test was performed in this research ticket.
