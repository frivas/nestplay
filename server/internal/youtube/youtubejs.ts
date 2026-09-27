import { Innertube } from "npm:youtubei.js@18.0.0";
import { emptyHomeMessage, homeItemsFromTree, landscapeVideos } from "./youtube_items.ts";

type Request = {
  op: "status" | "home" | "search";
  cursor?: string;
  query?: string;
};

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

function memoList(feed: { memo?: Map<string, unknown> }, type: string): unknown[] {
  const items = feed.memo?.get(type);
  return Array.isArray(items) ? items : [];
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
    let items = landscapeVideos([...page.videos]);
    if (items.length === 0 && !req.cursor) {
      const raw = await yt.actions.execute("/browse", {
        browseId: "FEwhat_to_watch",
        parse: false,
      }) as { data?: unknown };
      items = homeItemsFromTree(raw.data);
    }
    return {
      ok: true,
      state: "connected",
      items,
      cursor: next,
      message: items.length === 0 ? emptyHomeMessage(memoList(page, "FeedNudge")) : "",
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
