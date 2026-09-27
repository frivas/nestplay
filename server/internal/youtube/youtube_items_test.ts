import {
  emptyHomeMessage,
  homeItemsFromTree,
  landscapeVideos,
  videoID,
} from "./youtube_items.ts";
import { assertEquals } from "jsr:@std/assert@1";

Deno.test("classic Video renderer keeps landscape ids", () => {
  const items = landscapeVideos([
    { type: "Video", video_id: "abcdefghijk", title: { text: "Heat" } },
    { type: "ReelItem", video_id: "shortshort1", title: { text: "Short" } },
  ]);
  assertEquals(items, [{ id: "abcdefghijk", title: "Heat" }]);
});

Deno.test("LockupView uses content_id and metadata title", () => {
  const items = landscapeVideos([
    {
      type: "LockupView",
      content_type: "VIDEO",
      content_id: "lmnopqrstuv",
      metadata: { title: { text: "Lockup" } },
    },
    {
      type: "LockupView",
      content_type: "SHORT",
      content_id: "shortshort2",
      metadata: { title: { text: "Skip me" } },
    },
  ]);
  assertEquals(items, [{ id: "lmnopqrstuv", title: "Lockup" }]);
});

Deno.test("WEB lockupViewModel tree extracts landscape Home cards", () => {
  const items = homeItemsFromTree({
    contents: {
      twoColumnBrowseResultsRenderer: {
        tabs: [{
          tabRenderer: {
            content: {
              richGridRenderer: {
                contents: [
                  {
                    richItemRenderer: {
                      content: {
                        lockupViewModel: {
                          contentId: "abcdefghijk",
                          contentType: "LOCKUP_CONTENT_TYPE_VIDEO",
                          metadata: {
                            lockupMetadataViewModel: {
                              title: { content: "Heat" },
                            },
                          },
                        },
                      },
                    },
                  },
                  {
                    richItemRenderer: {
                      content: {
                        lockupViewModel: {
                          contentId: "shortshort3",
                          contentType: "LOCKUP_CONTENT_TYPE_SHORT",
                          metadata: {
                            lockupMetadataViewModel: {
                              title: { content: "Skip" },
                            },
                          },
                        },
                      },
                    },
                  },
                ],
              },
            },
          },
        }],
      },
    },
  });
  assertEquals(items, [{ id: "abcdefghijk", title: "Heat" }]);
});

Deno.test("empty Home surfaces the YouTube history nudge", () => {
  assertEquals(
    emptyHomeMessage([{ title: { text: "Your YouTube history is off" } }]),
    "Your YouTube history is off",
  );
  assertEquals(emptyHomeMessage([]), "");
  assertEquals(videoID("abcdefghijk"), "abcdefghijk");
});
