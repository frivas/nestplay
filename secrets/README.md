# Nestplay provider secrets (local only)

Copy `providers.env.example` to `providers.env` in this directory, or to
`~/.config/nestplay/providers.env`. Fill in values on the Mac. Do not paste them
into chat or commit them.

Nestplay accepts only unpaid plans:

- TMDB developer/personal API (non-commercial)
- OpenSubtitles.com **Free** consumer. Do not buy Light, Startup, Basic,
  Premium, Pro, Enterprise, or a VIP package.

The PoC consumer has **Allow anonymous downloads** and **Under dev** enabled:
downloads use the API key only (no user JWT), up to 100 per 24 hours. Search
does not spend that quota. Nestplay still skips local Spanish tracks, machine
translations, and weak name-only matches.

Username and password stay optional while those consumer flags remain on.
Override the path with `NESTPLAY_PROVIDERS_FILE` if needed.
