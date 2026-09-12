#!/usr/bin/env bash
# Import YouTube cookies from a local browser into Nestplay's Mac cookie file.
# Does not print cookie values. Do not paste the output of this script into chat.
set -euo pipefail

DEST="${NESTPLAY_YOUTUBE_COOKIES:-$HOME/Library/Application Support/nestplay/youtube-cookies.txt}"
BROWSER="${1:-chrome}"

if ! command -v yt-dlp >/dev/null 2>&1; then
  echo "yt-dlp is required. Install it with: brew install yt-dlp" >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required to keep only YouTube cookies." >&2
  exit 1
fi

mkdir -p "$(dirname "$DEST")"
chmod 700 "$(dirname "$DEST")" 2>/dev/null || true

workdir="$(mktemp -d "${TMPDIR:-/tmp}/nestplay-yt-cookies.XXXXXX")"
cleanup() { rm -rf "$workdir"; }
trap cleanup EXIT
raw="$workdir/all.txt"

# Load the browser cookie jar and let yt-dlp write it out. Simulate a cheap
# public watch URL so we do not download media. Failures after the jar is
# written are ignored. Redirect yt-dlp so cookie material never hits the terminal.
set +e
yt-dlp \
  --cookies-from-browser "$BROWSER" \
  --cookies "$raw" \
  --simulate \
  --skip-download \
  --no-playlist \
  --quiet \
  --no-warnings \
  --ignore-no-formats-error \
  "https://www.youtube.com/watch?v=jNQXAC9IVRw" \
  >/dev/null 2>"$workdir/ytdlp.err"
ytdlp_status=$?
set -e

if [[ ! -s "$raw" ]]; then
  echo "Could not read cookies from ${BROWSER}." >&2
  echo "Quit Chrome completely and run: $0 ${BROWSER}" >&2
  if [[ -s "$workdir/ytdlp.err" ]]; then
    echo "yt-dlp exited ${ytdlp_status}." >&2
  fi
  exit 1
fi

python3 - "$raw" "$DEST" <<'PY'
import pathlib
import sys

src, dest = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
keep = ("youtube.com", "youtu.be", "youtube-nocookie.com", "google.com", "googlevideo.com")
out = [
    "# Netscape HTTP Cookie File",
    "# Nestplay YouTube session — YouTube/Google hosts only.",
]
chosen = {}
order = []
for line in src.read_text(encoding="utf-8", errors="replace").splitlines():
    stripped = line.strip()
    if not stripped or stripped.startswith("#"):
        continue
    fields = stripped.split("\t")
    if len(fields) < 7:
        continue
    domain = fields[0].lower()
    host = domain.lstrip(".")
    if not any(host == h or host.endswith("." + h) for h in keep):
        continue
    name, value = fields[5].strip(), fields[6].strip()
    if not name or not value:
        continue
    rank = 0 if ("youtube.com" in domain or "youtu.be" in domain or "googlevideo.com" in domain) else 1
    prev = chosen.get(name)
    if prev is None:
        order.append(name)
        chosen[name] = (rank, fields[:7])
    elif rank <= prev[0]:
        chosen[name] = (rank, fields[:7])
out.extend("\t".join(chosen[name][1]) for name in order)
count = len(order)
if count == 0:
    raise SystemExit("No YouTube cookies were in the browser profile. Sign in to YouTube in that browser, then retry.")
dest.write_text("\n".join(out) + "\n", encoding="utf-8")
dest.chmod(0o600)
print(f"Wrote {count} YouTube cookies to {dest}")
print("Restart nestplay-server, then Refresh YouTube Home. Do not paste this file into chat.")
PY
