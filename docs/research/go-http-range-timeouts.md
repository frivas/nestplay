# Go HTTP clocks against Apple TV Range streaming
<!-- Preserved project research. -->

Research checked 11 September 2026. This records `net/http` facts for [Confirm Go HTTP clocks against Apple TV Range streaming](https://github.com/frivas/nestplay/issues/26). It does not choose Nestplay module shape — that belongs to [Decide how JSON clocks stay off the media byte pump](https://github.com/frivas/nestplay/issues/28). No Apple TV Range session was measured.

Go module `nestplay.local/server` declares Go 1.23. Citations below use the Go 1.24.0 `net/http` sources; the `Server` clock fields and `ResponseController` methods are the same surface.

## Nestplay today

The process uses one `http.Server` with `ReadHeaderTimeout: 5s`, `WriteTimeout: 10s`, `IdleTimeout: 30s`. JSON routes (`GET /health`, `GET /library`) share those clocks. There is no media Range route yet.

## What each clock covers

**`WriteTimeout`** is the maximum duration for writing the response. It is reset when a new request's header is read. It does not let handlers choose a per-request budget. Zero or negative means no timeout. [Server.WriteTimeout](https://pkg.go.dev/net/http#Server).

In `readRequest`, when `WriteTimeout > 0`, the server defers `SetWriteDeadline(now+WriteTimeout)` so the deadline is armed as soon as headers are finished and the handler runs:

```
if d := c.server.WriteTimeout; d > 0 {
    defer func() {
        c.rwc.SetWriteDeadline(time.Now().Add(d))
    }()
}
```

[go1.24.0 server.go `readRequest`](https://github.com/golang/go/blob/go1.24.0/src/net/http/server.go#L1036-L1040).

A 10s `WriteTimeout` therefore aborts a multi-gigabyte local-library Range body long before a film ends. It is compatible with today's small JSON replies.

**`IdleTimeout`** is the maximum time to wait for the *next* request when keep-alives are enabled. [Server.IdleTimeout](https://pkg.go.dev/net/http#Server). ConnState documents `StateIdle` as a connection that has finished handling a request and is waiting for a new one. [ConnState](https://pkg.go.dev/net/http#ConnState).

The server applies that clock only after `ServeHTTP` returns, as a **read** deadline while peeking for the next request:

```
w.finishRequest()
c.rwc.SetWriteDeadline(time.Time{})
// ...
c.setState(c.rwc, StateIdle, runHooks)
if d := c.server.idleTimeout(); d > 0 {
    c.rwc.SetReadDeadline(time.Now().Add(d))
}
```

[go1.24.0 server.go serve loop](https://github.com/golang/go/blob/go1.24.0/src/net/http/server.go#L2108-L2130). An in-flight Range body is not idle. `IdleTimeout: 30s` does not kill an active stream.

**`ReadHeaderTimeout`** (already 5s) bounds header read before the handler. **`ReadTimeout`** is unset (zero): no whole-request read deadline. GET Range requests have no body, so a missing `ReadTimeout` is not the film-length problem.

## ServeContent

`ServeContent` handles HTTP Range (and If-Range / If-Match / If-None-Match / If-Modified-Since). `*os.File` is a valid `io.ReadSeeker`; the name argument is never sent in the response. It does not transcode. [ServeContent](https://pkg.go.dev/net/http#ServeContent).

That matches the already-closed packaging choice: direct HTTP for compatible local-library files; conversion stays in a sibling process, not the Nestplay request goroutine. [Choose modular playback and realistic Mac packaging](https://github.com/frivas/nestplay/issues/4).

If `WriteTimeout` fires during `ServeContent` / `ServeFile`, historical Go issues record malformed or confusing responses (sendfile vs header flush). Do not rely on a global write deadline as a clean stop for media. [golang/go#43822](https://github.com/golang/go/issues/43822).

## Documented ways to isolate JSON clocks from a byte pump

Handlers cannot change `Server.WriteTimeout` itself. They can change the connection write deadline:

**`ResponseController.SetWriteDeadline`** (Go 1.20+) overrides the server write deadline for that request. A zero `time.Time` means no deadline. Setting a deadline after it has already expired does not extend it — clear or replace it *before* the 10s budget elapses. [ResponseController.SetWriteDeadline](https://pkg.go.dev/net/http#ResponseController.SetWriteDeadline). Proposal record: `SetWriteDeadline` overrides `WriteTimeout`. [golang/go#54136](https://github.com/golang/go/issues/54136).

The server's `response.SetWriteDeadline` is `c.conn.rwc.SetWriteDeadline`. [go1.24.0 server.go](https://github.com/golang/go/blob/go1.24.0/src/net/http/server.go#L500-L502).

Two `http.Server` values (different listeners or a shared listener pattern) can give JSON a tight `WriteTimeout` and media `WriteTimeout <= 0`. That is a process-shape choice for the grilling ticket, not a requirement of `net/http`.

`Hijack` takes the connection away from `net/http`. It drops `ServeContent` Range handling. It is not needed to stream a file.

`TimeoutHandler` wraps a handler with a duration and does **not** support `SetWriteDeadline` / `Hijack` / `Flush`. Do not wrap a media pump in it. [golang/go#54136](https://github.com/golang/go/issues/54136).

## Facts for the next ticket

1. Today's 10s `WriteTimeout` will abort Apple TV Range once a media route exists.
2. `IdleTimeout` will not.
3. `ServeContent` is the stdlib Range mechanism; it copies bytes from a seekable file and does not run FFmpeg.
4. Isolation options that `net/http` actually documents: per-request `SetWriteDeadline(time.Time{})` (or a sliding deadline) inside a media handler, and/or a second `http.Server` with no write timeout. Module shape is still open.
