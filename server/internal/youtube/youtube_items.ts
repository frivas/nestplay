export type Item = { id: string; title: string };

const SHORT_TYPES = new Set(["ReelItem", "ShortsLockupView"]);
const SHORT_CONTENT = new Set(["SHORT", "CLIP"]);

export function videoID(value: unknown): string {
  if (typeof value !== "string") return "";
  if (!/^[A-Za-z0-9_-]{11}$/.test(value)) return "";
  return value;
}

export function videoTitle(value: unknown): string {
  if (typeof value === "string") return value.trim();
  if (value && typeof value === "object") {
    const record = value as { text?: unknown; content?: unknown; toString?: () => string };
    if (typeof record.text === "string") return record.text.trim();
    if (typeof record.content === "string") return record.content.trim();
    if (typeof record.toString === "function") {
      const text = record.toString();
      if (text && text !== "[object Object]") return text.trim();
    }
  }
  return "";
}

function itemID(video: {
  video_id?: unknown;
  id?: unknown;
  content_id?: unknown;
}): string {
  return videoID(video.video_id) || videoID(video.id) || videoID(video.content_id);
}

function normalizedContentType(value: unknown): string {
  if (typeof value !== "string") return "";
  return value.replace(/^LOCKUP_CONTENT_TYPE_/, "");
}

function itemTitle(video: {
  title?: unknown;
  metadata?: { title?: unknown; lockupMetadataViewModel?: { title?: unknown } };
}): string {
  return videoTitle(video.title) ||
    videoTitle(video.metadata?.title) ||
    videoTitle(video.metadata?.lockupMetadataViewModel?.title);
}

function isShort(video: {
  type?: string;
  is_short?: boolean;
  content_type?: string;
}): boolean {
  if (SHORT_TYPES.has(video.type ?? "") || video.is_short === true) return true;
  return SHORT_CONTENT.has(normalizedContentType(video.content_type));
}

export function landscapeVideos(videos: unknown[]): Item[] {
  const out: Item[] = [];
  const seen = new Set<string>();
  for (const raw of videos) {
    const v = raw as {
      type?: string;
      is_short?: boolean;
      content_type?: string;
      video_id?: unknown;
      id?: unknown;
      content_id?: unknown;
      title?: unknown;
      metadata?: { title?: unknown };
    };
    if (!v || isShort(v)) continue;
    const id = itemID(v);
    const title = itemTitle(v);
    if (!id || !title || seen.has(id)) continue;
    seen.add(id);
    out.push({ id, title });
  }
  return out;
}

export function emptyHomeMessage(nudges: unknown[]): string {
  if (nudges.length === 0) return "";
  const first = nudges[0] as { title?: unknown };
  return videoTitle(first.title);
}

export function homeItemsFromTree(node: unknown): Item[] {
  const videos: unknown[] = [];
  collectLockups(node, videos, 0);
  return landscapeVideos(videos);
}

function collectLockups(node: unknown, videos: unknown[], depth: number): void {
  if (depth > 30 || node == null) return;
  if (Array.isArray(node)) {
    for (const item of node) collectLockups(item, videos, depth + 1);
    return;
  }
  if (typeof node !== "object") return;
  const record = node as Record<string, unknown>;
  const lockup = record.lockupViewModel;
  if (lockup && typeof lockup === "object") {
    const vm = lockup as {
      contentId?: unknown;
      contentType?: unknown;
      metadata?: { lockupMetadataViewModel?: { title?: unknown } };
    };
    videos.push({
      type: "LockupView",
      content_id: vm.contentId,
      content_type: normalizedContentType(vm.contentType),
      metadata: { title: vm.metadata?.lockupMetadataViewModel?.title },
    });
  }
  for (const value of Object.values(record)) {
    collectLockups(value, videos, depth + 1);
  }
}
