# Nestplay downloader components, checked 2026-09-10

Primary-source research, not execution benchmarks. Footprints below are architectural assessments unless the project states numbers.

## yt-dlp (preferred core)

Native Mac executable or Python package; no always-running service or database needed. CLI plus embeddable Python API, search extraction, individual videos, channels and playlists, selection/filter/archive controls, JSON metadata, thumbnails, subtitles, FFmpeg merging/remuxing/reencoding. It is a download engine, not a media server or finished YouTube browsing UI. Best fit for a bespoke Nestplay server that starts download processes only when needed. [README](https://github.com/yt-dlp/yt-dlp/blob/master/README.md?plain=1).

Core repository and PyPI distributions are Unlicense. **Do not label all distributed yt-dlp binaries simply Unlicense**: PyInstaller executables include GPLv3+ code and combined binaries are GPLv3+; Unix zipimport/tarball include ISC/MIT components. [Licensing section](https://github.com/yt-dlp/yt-dlp/blob/master/README.md?plain=1).

Latest stable found: **2026.08.19**; release includes ongoing YouTube client/extractor fixes. [Releases](https://github.com/yt-dlp/yt-dlp/releases).

Current full YouTube support requires JavaScript challenge scripts (`yt-dlp-ejs`) plus a supported JS runtime. Deno is recommended/default enabled, minimum 2.3.0; Node minimum 22, explicitly enabled. Official Mac executable bundles EJS scripts, but runtime remains separate. For Python use default dependency group and keep EJS version aligned. Bun now deprecated. Preinstalling scripts avoids fetching EJS at runtime. [EJS guide](https://github.com/yt-dlp/yt-dlp/wiki/EJS).

Some YouTube clients additionally require PO Tokens. Official recommendation is a provider plugin; manual tokens are no longer recommended because of per-video binding. This is distinct from the EJS runtime, and may add dependencies. Do not promise “any video” or zero maintenance; validate against the actual account/IP. [PO Token guide](https://github.com/yt-dlp/yt-dlp/wiki/PO-Token-Guide).

## ytdl-sub (preferred metadata automation)

**GPL-3.0**. YAML subscription CLI wrapping yt-dlp; schedules can launch and exit, keeping idle overhead low. Produces media-center-ready filenames, NFO, episode thumbnails, poster/fanart; presets treat channels as TV shows and accept playlist sources. It has no playback/transcoding server or general YouTube discovery UI. CLI/files are the composition boundary, not a documented public HTTP API. Pip/local installation is listed; native Mac is plausible Python deployment, but README does not claim a dedicated supported macOS binary. Docker is optional; avoid the recommended “Web-GUI” container when optimizing weight because it is code-server with ytdl-sub installed, not a lightweight bespoke downloader frontend. [Repository and install options](https://github.com/jmbannon/ytdl-sub).

Latest stable **2026.08.26.post1**, feature work including optional source-sync pruning, alongside frequent 2026 releases. [Releases](https://github.com/jmbannon/ytdl-sub/releases).

Follow-up install-doc verification: the upstream environment-agnostic page explicitly says **PIP installation is not recommended** because of potential FFmpeg requirement conflicts. It requires Python 3.10+ and FFmpeg; local repo installation is presented as development setup. Thus native Mac remains a viable evaluated deployment route with isolation/testing, not upstream-preferred production packaging. [Install details](https://ytdl-sub.readthedocs.io/en/latest/guides/install/agnostic.html).

## MeTube (preferred ready web download UI)

**AGPL-3.0**, Python backend and Angular UI; official multi-architecture Docker images (amd64/arm64). One service, no Elasticsearch/Redis stack; likely lighter than Tube Archivist but **no trustworthy upstream idle RAM benchmark found**. Docker on Mac adds a Linux VM. Source deployment possible but less turnkey than containers. Individual videos, playlists/channels, captions, thumbnails and — in current version — **automatic channel/playlist subscriptions**. URL submission and browser extensions/bookmarklets are primary intake; not a full YouTube discovery frontend. Metadata JSON/thumbnails can be configured, but NFO/library organization needs an adapter or another tool. Completed files can be served/downloaded; not an adaptive streaming/media-library server. [Repository](https://github.com/alexta69/metube).

Source exposes HTTP `POST /add`, `POST /subscribe`, retry/cancel and other queue endpoints; practical for Nestplay integration, but source-level API dependency should be pinned. [Backend API source](https://github.com/alexta69/metube/blob/master/app/main.py).

Maintenance evidence: feature commit **2026-08-28**; yt-dlp updated to 2026.08.19 on Aug20; active fixes through August. [Commits](https://github.com/alexta69/metube/commits/master).

## Pinchflat (good composition, maintenance caveat)

**AGPL-3.0**; self-contained single Docker container with SQLite, Elixir/Phoenix plus downloader dependencies. Explicitly designed for channel/playlist subscriptions, media-center interoperability, RSS and custom after-download scripts; individual downloads supported but not its strongest workflow. Does not aim to be an in-app playback library. Docker is documented main deployment, bringing VM overhead on Mac; no formal RAM figure found. Compose using common files/NFO and hooks. Did not find a documented stable public REST API; router is mostly browser/feed endpoints. [Repository](https://github.com/kieraneglin/pinchflat), [router](https://github.com/kieraneglin/pinchflat/blob/master/lib/pinchflat_web/router.ex).

**Maintenance is paused, not confidently active.** Maintainer's still-open “Temporary Development Pause” dated **2025-09-26** says feature/minor bug work is paused with breaking fixes watched. [Announcement](https://github.com/kieraneglin/pinchflat/issues/800). Latest release **v2025.9.26** added Deno; latest master commits observed **2025-12-16**. It has a yt-dlp auto-updater, but that does not guarantee current runtime/extractor integration. [Releases](https://github.com/kieraneglin/pinchflat/releases), [commits](https://github.com/kieraneglin/pinchflat/commits/master).

## Tube Archivist (feature-rich, not lightweight)

**GPL-3.0** ([license](https://github.com/tubearchivist/tubearchivist/blob/master/LICENSE)). Django application plus Redis and Elasticsearch, Docker required for documented production deployment. Upstream quotes around **2 GB available RAM for small testing / 4 GB mid-large installation**, dual core/four threads minimum. These are stack requirements, not measured idle RAM. Supports video/channel/playlist intake and subscriptions, searchable metadata, local web playback, watched status, browser companion and Jellyfin plugin. It indexes the archive rather than offering a universal live YouTube browsing replacement. Compatible format selection is needed for Apple browsers; do not assume on-demand adaptive transcodes. [Repository](https://github.com/tubearchivist/tubearchivist).

Strongest documented REST interface of these downloader UIs: token authentication and Swagger at `/api/docs/`. [API intro](https://docs.tubearchivist.com/api/introduction/), [API docs](https://docs.tubearchivist.com/api/docs/).

Latest release **v0.5.12, Aug25** (page omits year in text; current release list and dependency context indicate 2026, verify if requiring precise year). Release updates yt-dlp and PO-provider components: actively maintained. [Release](https://github.com/tubearchivist/tubearchivist/releases/tag/v0.5.12).

## Recommendation to parent

Additional candidate **TubeSync**, AGPL-3.0: container-oriented Python/Django scheduler with SQLite by default, yt-dlp/FFmpeg, retries/backoff, channels/playlists, artwork/metadata and built-in Jellyfin/Plex refresh integration. No published RAM benchmark found. Library playback handled by downstream media server; no dedicated Apple TV client. Useful alternative to paused Pinchflat when recurring channel capture matters more than quick individual-video intake. Latest **v0.18.4, September 4, 2026**; predecessor release notes explicitly date the cycle 2026, active fixes and PO-provider updates. [Repository](https://github.com/meeb/tubesync), [releases](https://github.com/meeb/tubesync/releases).

Native yt-dlp worker + EJS/Deno + FFmpeg + local library metadata is the lightest building block. Use ytdl-sub if channel/playlist-to-TV-library formatting avoids writing metadata adapters. MeTube is now a stronger maintained choice for ready-made download UI plus subscriptions than Pinchflat, whose pause should be explicit. Tube Archivist is a viable archival product but mismatches the lightweight Mac requirement. None of these supplies a beautiful native Apple TV app by itself. Preserve source video IDs, original playlist order and many-to-many playlist membership in Nestplay separately from media-server episode mappings. Local playback and library stay LAN-only; YouTube browsing/downloads necessarily use outbound Internet.
