# Nestplay: YouTube Data API v3 versus InnerTube
<!-- Preserved project research. -->

Research checked 12 September 2026 against Google's current Data API documentation and YouTube's Terms of Service. No Google Cloud project was created. No Data API credentials were issued. Cookie values were not copied into this note.

This note answers whether Nestplay should replace the YouTube.js InnerTube adapter with YouTube Data API v3 for [Browse YouTube Home from Apple TV](https://github.com/frivas/nestplay/issues/34). It does not change the earlier adapter-boundary decision in `youtube-session.md` until the user accepts a new one.

## Verdict

Data API v3 is the supported, lower-policy-risk way to talk to YouTube. It cannot supply the personalized Home shelf the YouTube app shows. InnerTube can, in principle, and is what the current adapter uses. Switching Home to v3 would not fix the empty shelf; it would replace it with a different empty or generic list.

## What Data API v3 actually returns

Google documents `activities.list` as the historical Home stand-in. The `home` filter is deprecated. Authorized requests that still set it receive items similar to a logged-out YouTube home page, not the signed-in recommended feed. Errors include `homeParameterDeprecated` and `homeChannelNotFound`. Quota cost is 1 unit. [activities.list](https://developers.google.com/youtube/v3/docs/activities/list)

`activities.list` with `mine=true` is the authenticated user's own channel activity (uploads and similar events), not recommendations. [activities.list](https://developers.google.com/youtube/v3/docs/activities/list)

`search.list` is keyword search. It is not Home. The old `relatedToVideoId` parameter was removed from the documentation after 7 August 2023; there is no official replacement for "related" or "recommended because you watched." Default quota for `search.list` is a separate bucket of 100 calls per day, 1 unit per call including each continuation page. [search.list](https://developers.google.com/youtube/v3/docs/search/list), [revision history](https://developers.google.com/youtube/v3/revision_history), [quota](https://developers.google.com/youtube/v3/determine_quota_cost)

`videos.list` with `chart=mostPopular` is a regional chart. As of 21 July 2025 it is Trending Music, Movies, and Gaming, not the former Trending Now list, and not personalized Home. Cost 1 unit. [videos.list](https://developers.google.com/youtube/v3/docs/videos/list), [revision history](https://developers.google.com/youtube/v3/revision_history)

`subscriptions.list` plus `playlistItems.list` can list channels the account follows and their uploads. That is a subscriptions inbox, not the algorithm Home feed. Cost 1 unit each. [quota](https://developers.google.com/youtube/v3/determine_quota_cost)

Default new-project quota, last updated 4 September 2026: 100 `search.list` calls, 100 `videos.insert` calls, and 10,000 units per day for other methods. [quota](https://developers.google.com/youtube/v3/determine_quota_cost)

Enabling the API requires a Google Cloud project, an API key and/or OAuth 2.0 client, and user consent for any private channel data. [overview](https://developers.google.com/youtube/v3/getting-started)

API clients must also follow YouTube API Services Terms: access only through documented methods, plus developer policies that restrict cloning YouTube's user experience. [API TOS](https://developers.google.com/youtube/terms/api-services-terms-of-service), [developer policies guide](https://developers.google.com/youtube/terms/developer-policies-guide)

## InnerTube risks that were not scored when the adapter shipped

YouTube's consumer Terms prohibit accessing the Service by automated means (robots, botnets, scrapers) except search-engine robots.txt, prior written permission, or applicable law. InnerTube is not a documented YouTube API Service. Using it is a ToS risk even for a private Mac-and-Apple-TV PoC. [YouTube Terms](https://www.youtube.com/static?template=terms)

Operational risks observed on this Mac, 12 September 2026, without printing cookie values:

- The same Chrome cookie jar that loads a normal youtube.com Home (22 `lockupViewModel` cards, no history-off banner) yields an InnerTube `FEwhat_to_watch` browse with `feedNudgeRenderer` only and zero video/lockup renderers.
- youtubei.js 18.0.0 ships WEB client version `2.20260623.01.00`. The live youtube.com page used `2.20260911.01.00`.
- Google My Activity showed YouTube History on. InnerTube still returned YouTube's "Your YouTube history is off" nudge. That message is InnerTube's degraded Home, not a Nestplay history dump and not proof that History is off in Google settings.
- Search through the same adapter returned landscape videos. Home is the broken call, not the whole session.

Cookie import remains a high-sensitivity credential: a Netscape/header jar is equivalent to a logged-in browser session on the Mac. Keep it off git, chat, tvOS, and logs. That risk exists for InnerTube and would be replaced, not removed, by storing an OAuth refresh token for Data API v3.

youtubei.js is MIT, pinned, and replaceable. It has no SLA. Parser classes lag YouTube's WEB and TV renderers (`LockupView` on WEB Home; missing `TvBrowse` on TVHTML5).

## What v3 is useful for in Nestplay

Useful later, not as a Home replacement:

- Official keyword search (`search.list`), with quota.
- Duration, captions list, and channel metadata for IDs Nestplay already has (`videos.list`, `captions.list`).
- Optional subscriptions inbox if a future ticket wants "from channels I follow" instead of algorithm Home.

Not useful for #34:

- Personalized Home / What to watch.
- Related-to-this-video rails.
- Passing InnerTube's empty Home off as "fixed" by showing `mostPopular` or `activities.list&home=true` logged-out items. Nestplay's existing contract is not to invent generic recommendations when the personalized feed is missing.

## Decision still needed

Keep InnerTube for Home and repair the WEB browse so it matches youtube.com (client version / request shape), accepting ToS and breakage risk for this private PoC.

Or drop personalized Home from the PoC and use Data API v3 for search plus an explicit non-Home surface (search, subscriptions, or charts), with Google Cloud OAuth on the Mac.

Do not silently substitute v3 charts or channel activity for the YouTube app Home feed.
