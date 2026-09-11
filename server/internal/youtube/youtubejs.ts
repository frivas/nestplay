import { Innertube } from "npm:youtubei.js@18.0.0";

type Request = {
  op: "status" | "home" | "search";
  cursor?: string;
  query?: string;
};

type Item = { id: string; title: string };

const SHORT_TYPES = new Set(["ReelItem", "ShortsLockupView"]);
const homes = new Map<string, Awaited<ReturnType<Innertube["getHomeFeed"]>>>();
const searches = new Map<
  string,
  Awaited<ReturnType<Innertube["search"]>>
>();

let tube: Innertube | null = null;

function cookieHeader(): string {
  const path = Deno.env.get("NESTPLAY_YOUTUBE_COOKIES_FILE") ?? "";
  if (!path) return "";
  try {
    return Deno.readTextFileSync(path).trim();
  } catch {
    return "";
  }
}

function videoID(value: unknown): string {
  if (typeof value !== "string") return "";
  if (!/^[A-Za-z0-9_-]{11}$/.test(value)) return "";
  return value;
}

function videoTitle(value: unknown): string {
  if (typeof value === "string") return value.trim();
  if (value && typeof value === "object") {
    const record = value as { text?: unknown; toString?: () => string };
    if (typeof record.text === "string") return record.text.trim();
    if (typeof record.toString === "function") {
      const text = record.toString();
      if (text && text !== "[object Object]") return text.trim();
    }
  }
  return "";
}

function landscapeVideos(videos: unknown[]): Item[] {
  const out: Item[] = [];
  const seen = new Set<string>();
  for (const raw of videos) {
    const v = raw as {
      type?: string;
      is_short?: boolean;
      video_id?: unknown;
      id?: unknown;
      title?: unknown;
    };
    if (!v || SHORT_TYPES.has(v.type ?? "") || v.is_short === true) continue;
    const id = videoID(v.video_id) || videoID(v.id);
    const title = videoTitle(v.title);
    if (!id || !title || seen.has(id)) continue;
    seen.add(id);
    out.push({ id, title });
  }
  return out;
}

async function session(): Promise<Innertube> {
  if (tube) return tube;
  const cookie = cookieHeader();
  if (!cookie) {
    throw new Error("reconnect");
  }
  tube = await Innertube.create({ cookie });
  return tube;
}

function cursorToken(): string {
  return crypto.randomUUID().replaceAll("-", "");
}

async function handle(req: Request): Promise<unknown> {
  if (req.op === "status") {
    const cookie = cookieHeader();
    if (!cookie) {
      return {
        ok: true,
        state: "reconnect_required",
        message: "YouTube needs a reconnect on your Mac.",
      };
    }
    try {
      await session();
      return { ok: true, state: "connected" };
    } catch {
      tube = null;
      return {
        ok: true,
        state: "reconnect_required",
        message: "YouTube needs a reconnect on your Mac.",
      };
    }
  }

  if (req.op === "home") {
    const yt = await session();
    const feed = req.cursor ? homes.get(req.cursor) : await yt.getHomeFeed();
    if (!feed) {
      return { ok: false, state: "unavailable", message: "That page expired. Refresh YouTube Home." };
    }
    if (req.cursor) {
      homes.delete(req.cursor);
    }
    const page = req.cursor ? await feed.getContinuation() : feed;
    const next = page.has_continuation ? cursorToken() : "";
    if (next) homes.set(next, page);
    return {
      ok: true,
      state: "connected",
      items: landscapeVideos([...page.videos]),
      cursor: next,
    };
  }

  if (req.op === "search") {
    const query = (req.query ?? "").trim();
    if (!query) {
      return { ok: true, state: "connected", items: [], cursor: "" };
    }
    const yt = await session();
    const feed = req.cursor
      ? searches.get(req.cursor)
      : await yt.search(query, { type: "video" });
    if (!feed) {
      return { ok: false, state: "unavailable", message: "That page expired. Search again." };
    }
    if (req.cursor) {
      searches.delete(req.cursor);
    }
    const page = req.cursor ? await feed.getContinuation() : feed;
    const next = page.has_continuation ? cursorToken() : "";
    if (next) searches.set(next, page);
    return {
      ok: true,
      state: "connected",
      items: landscapeVideos([...page.videos]),
      cursor: next,
    };
  }

  return { ok: false, state: "unavailable", message: "Unknown YouTube operation." };
}

const decoder = new TextDecoder();
let buffer = "";

for await (const chunk of Deno.stdin.readable) {
  buffer += decoder.decode(chunk, { stream: true });
  let newline = buffer.indexOf("\n");
  while (newline >= 0) {
    const line = buffer.slice(0, newline).trim();
    buffer = buffer.slice(newline + 1);
    if (line) {
      try {
        const result = await handle(JSON.parse(line) as Request);
        console.log(JSON.stringify(result));
      } catch (err) {
        const reconnect = String(err).includes("reconnect");
        console.log(JSON.stringify({
          ok: false,
          state: reconnect ? "reconnect_required" : "unavailable",
          message: reconnect
            ? "YouTube needs a reconnect on your Mac."
            : "YouTube Home could not be loaded.",
        }));
      }
    }
    newline = buffer.indexOf("\n");
  }
}
