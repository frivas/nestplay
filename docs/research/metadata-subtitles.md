# Nestplay: original-language metadata and Spanish subtitles
<!-- Preserved project research. -->

Research decision, 10 September 2026. Documentation and source review only; no accounts, credentials, downloads or playback tests were performed.

## Decision

Use TMDB for movie identity, original-language descriptions and artwork; OpenSubtitles.com REST for external movie subtitles; yt-dlp for existing YouTube captions. Keep matching policy in Nestplay behind provider interfaces, even if Jellyfin performs retrieval. No separate VLC, Bazarr or translation service is needed for the PoC. These are network data providers, not an entirely open-source/offline metadata supply chain.

This closes provider selection at the research level. Coverage, synchronization and account quotas remain acceptance tests, not guarantees.

## Movie identity and original language

TMDB search results expose `original_title` and `original_language`; use the selected TMDB identifier as permanent identity. The normal `title` and `overview` are localized data, so requesting one library-wide language will not satisfy a multilingual original-language library. After identification, select metadata in the item's original language. [Search and details](https://developer.themoviedb.org/docs/search-and-query-for-details), [languages](https://developer.themoviedb.org/docs/languages).

Use the movie translations endpoint to select the original-language overview explicitly. Do not treat an English fallback as the original. If the matching translation is absent, retain the original title and omit the overview with a quiet metadata-incomplete state. Multilingual films still require a canonical language choice; TMDB's `original_language` is a workable PoC default and should be manually correctable. This is Nestplay policy derived from the API, not a guarantee of catalog completeness. [Translations endpoint](https://developer.themoviedb.org/reference/movie-translations).

Proposed identification order: explicit provider ID in a local sidecar or filename, then parsed title/year search, then user correction when candidates are ambiguous. Never confidently attach a remake's artwork merely because its title matches. Persist the chosen ID and a match lock so rescans cannot undo a correction. Automatic identity matching needs a “Fix match” action; universal correctness is impossible from arbitrary filenames alone.

For covers, use the identified movie's images and prefer its original language or artwork without text. TMDB's ordinary poster field can fall back to the highest-rated image in another language; explicitly filter image records if strict original-language artwork matters. Backdrops usually have no language. [Image-language behavior](https://developer.themoviedb.org/docs/image-languages).

TMDB requires an account/API key. Its developer API is free for noncommercial use with attribution; commercial use requires a separate agreement. Include its prescribed attribution and approved logo in Credits. The API is closed source; open-source Nestplay code does not grant rights to redistribute the provider's catalog. [TMDB FAQ](https://developer.themoviedb.org/docs/faq). Cache fetched records/art locally, serialize modest background refreshes, and honor HTTP 429; the published approximate upper bound is around 40 requests/second and can change. [Rate limits](https://developer.themoviedb.org/docs/rate-limiting).

## Spanish movie subtitles

The VLC analogy is valid: OpenSubtitles' VLC extension searches by file hash first, then name, and exposes synchronization, trusted uploader and machine-translation indicators. It supports both Spanish and Latin Spanish. Use its workflow as a reference, not a runtime dependency; its code is GPLv3. A hash result is strong evidence, not independent proof that every uploaded subtitle is correct. [Provider-maintained VLC extension](https://github.com/opensubtitles/vlsub-opensubtitles-com).

Proposed Nestplay pipeline:

1. Wait for the movie file to finish copying. Inspect embedded tracks and accompanying subtitle files before spending API quota. Normalize Spanish language tags; keep full-dialogue and forced-only tracks distinct.
2. Search all supported Spanish variants. Rank exact file-hash matches first, then the confirmed movie ID plus matching release/cut, duration or frame-rate evidence where returned. Never rank a regional variant above synchronization.
3. Download the best credible full subtitle, preserve its provider ID and matching evidence, and save a small language-tagged sidecar. Avoid rewriting or duplicating the movie. Validate parsing and basic timestamp bounds; these checks cannot prove dialogue synchronization.
4. If no credible result exists, leave the movie playable and record `missing`, `needs review`, or `quota deferred` accurately. Retry later and allow manual import/correction. Enable an available Spanish track by default, without an overlay during playback.

OpenSubtitles.com supports title, release, IMDb ID and hash searches. Its help pages disagree on free authenticated quotas: one says 20 daily; another starts regular users at 10. Both describe five anonymous daily downloads. Therefore read actual account/response allowances and reset information; do not hardcode an advertised quota. Large folder imports may take multiple days on a free account. [API overview](https://opensubtitles.tawk.help/article/about-the-api), [getting started](https://opensubtitles.tawk.help/article/getting-started).

A consumer API key and application User-Agent are required; personal account authentication is the practical PoC approach. Paid VIP/Pro plans exist. Pro packages can remove end-user login and increase quotas, but the reviewed public help page does not provide a reliable numeric current price; no purchase is needed to choose this architecture. Confirm the account's offered limits before implementation testing. The service's access terms and subtitle rights are separate from the extension's GPL license. [Pro packages](https://opensubtitles.tawk.help/article/pro-packages).

## YouTube captions

Existing Spanish captions are straightforward in principle: yt-dlp documents subtitle listing, `--write-subs`, language selection via `--sub-langs`, and separate automatic captions. Start with supplied Spanish tracks. If supporting YouTube's existing automatic Spanish captions, distinguish them visibly in the subtitle menu and exclude automatic translations with `youtube:skip=translated_subs`. This downloads available tracks; it does not run speech recognition or create translations. Keep caption failure independent from video completion. [yt-dlp subtitle and extractor options](https://github.com/yt-dlp/yt-dlp#subtitle-options).

A further original-language caveat: yt-dlp documents preferring primary-language video metadata but falling back to English. Strict original-language YouTube descriptions therefore also need provenance handling; do not claim the default output always meets that policy. [YouTube extractor arguments](https://github.com/yt-dlp/yt-dlp#extractor-arguments).

## Jellyfin composition and remaining proof

Jellyfin includes TMDB, OMDb and local NFO providers; OMDb is English-only and should not be the original-language fallback. Its official GPLv3 OpenSubtitles plugin can download subtitles. Keep only one owner of automatic subtitle retrieval to prevent duplicate requests and quota consumption. For strict language policy, Nestplay may supply corrected local metadata and let Jellyfin focus on playback. [Metadata providers](https://jellyfin.org/docs/general/server/metadata/), [OpenSubtitles plugin](https://github.com/jellyfin/jellyfin-plugin-opensubtitles).

Before claiming acceptance: test original-language overviews/posters for English, Spanish and another-language movies; an ambiguous remake and missing overview; embedded full/forced Spanish tracks; exact-hash and name-only subtitles; wrong-cut rejection; no-result and quota exhaustion; and YouTube supplied/automatic/missing Spanish captions. Verify TV default selection, accented characters, seeking synchronization and turning subtitles off. Subtitle format affects whether playback requires conversion or burn-in, so this belongs in the Apple TV media matrix. [Jellyfin subtitle compatibility](https://jellyfin.org/docs/general/clients/codec-support/).
