# Nestplay
<!-- Canonical project glossary. -->

Nestplay is a private home media library and a way to browse personalized YouTube recommendations and select videos for local downloading.

## Language

**Local library**:
The completed downloaded videos and movies supplied through a watched folder that are available for playback. Nestplay on Apple TV opens here.
_Avoid_: YouTube feed, local feed, Home

**YouTube Home**:
The personalized recommendations associated with the single connected YouTube account, including additional pages of results. On Apple TV this is landscape recommended videos only, refreshable, with search.
_Avoid_: Subscriptions, trending, Shorts, chips, community posts, live feed

**Library playlist**:
A user-created ordered list of local-library items.
_Avoid_: YouTube playlist

**Collection**:
The containing folder of a local-library file, relative to the watched folder. Next, Previous, and auto-advance stay inside a collection and follow filename order.
_Avoid_: library playlist, YouTube playlist, filesystem path

**Download queue**:
The selected videos waiting to download or currently downloading; a queued video is not yet a playable library item. Finished YouTube files are published into the watched folder passed as `-library-root`. Scratch files stay in a hidden `.nestplay-work` folder there, are deleted after each job, and never appear on Apple TV. Existing unplayable files are converted next to the original; the original is deleted only after that succeeds.

**Watched folder**:
A folder where the user places movie files for automatic library discovery. Nested directories (for example Mateo → PJ Masks) are how the Apple TV library is browsed. A file joins the local library only after it has finished arriving, not while a copy is still growing.

**Spanish subtitle match**:
A Spanish subtitle track selected for compatibility and synchronization with a particular video, regardless of Spanish regional variant.

**Original-language metadata**:
Movie titles and descriptions presented in the movie's original language, distinct from the user's preferred subtitle language.

**Storage reserve**:
The 20 GB of free disk space at which new downloads and conversion writes stop. Completed local-library playback continues. The Apple TV sees the refuse on the download queue item, never over video.

**Temporary playback copy**:
Replaced for this PoC: ffmpeg writes a Living Room–playable file, then the unplayable original is deleted. Unplayable files stay hidden until that succeeds.

**Playable profile**:
Living Room (Apple TV HD): 1080p SDR H.264 + AAC in MP4. H.264 in MP4 with Opus may already play and stays visible. New YouTube files must meet this profile after ffmpeg. Below 720p is not downloaded.
