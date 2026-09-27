# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## Overview

Video server component of the **ikascrew** VJ (video jockey) system. It renders video to an OpenCV window and is controlled remotely over gRPC (effect switching, transitions, volume/light/wait parameters). Project/content metadata is prepared once from a separate `ikasbox` HTTP service into a local work file; at runtime the server has no external dependencies.

## Build & Run

Requires OpenCV installed locally (gocv dependency). Tests cover the package root (`handler_test.go`, `stream_test.go`, `opening_test.go`, `video_gen_test.go`) and `config/`; run them with `go test ./...`.

```
go build ./...                          # build everything
go run ./cmd/ika-server create <id>     # fetch project from ikasbox, write work file .server/config.json
go run ./cmd/ika-server start           # run the server from the work file (no ikasbox needed)
go run ./cmd/ika-server -ikasbox [-db <path>] start   # co-hosted mode: also runs ikasbox HTTP :5555 in-process
go run ./cmd/check                      # manual smoke test: cycles through all contents with transitions
```

**Co-hosted mode (`-ikasbox`)**: `ikasbox.go` starts ikasbox (HTTP :5555) inside the same process (`ikasbox.Start` in a goroutine) and registers two extra API endpoints via `api.AddEndpoint`: `v1/server/status` (capability probe — the React UI shows a per-project "Server" button only when this answers) and `v1/server/create` (writes the work file via the same loopback HTTP path as CLI create, then hot-reloads `config` and `output.Set`). Creating is refused while fullscreen (= performing). In this mode the server starts even without a work file (empty mapping until created from the UI). Do not run a standalone ikasbox against the same DB simultaneously (SQLite locking). `config.Reload` swaps the config under an RWMutex; in-flight handlers keep their snapshot.

`create` requires a reachable ikasbox server (default `localhost:5555`, endpoint `http://<DBIP>:<DBPort>/project/content/list/<project-id>`); `start` and `check` only need the previously created `.server/config.json` (gitignored). gRPC listens on port 55555 by default (configurable via `config.Option` functions in `config/option.go`).

## Architecture

Startup flow (`server.go: Start`): load config from the work file → start a UDP multicast announcer (`ikascrew/core/multicast`) → create the initial "terminal" video (system-info splash from `opening.go`) → create a `Window` → start the gRPC server in a goroutine → `Window.Play` blocks on the main thread until shutdown.

Key pieces and how they interact:

- **`Window` (window.go)** — owns the gocv window and the render loop. `Play` must run on a locked OS thread (OpenCV requirement). Each iteration it either applies a pushed video (`Push`, called from gRPC handlers, via the `wait` channel), handles Ctrl+C (captured with `signal.Notify`), or renders a frame. Shutdown policy: fullscreen means "performing live" — while fullscreen, ESC / window close / Ctrl+C are all ignored; when windowed, any of them exits cleanly (release stream → close window → main returns).
- **`Stream` (stream.go)** — the transition engine. Holds three video slots (`now_video`, `old_video`, `release_video`) that rotate on each `Switch`; frames are blended with `gocv.AddWeighted` using `now_value`/`old_value` (0–200, `SWITCH_VALUE`) as crossfade alpha. Intermediate blend states are a feature (e.g. holding 50:50, or 25:25:50 across three videos), so all active slots are decoded every frame — their `Next()` calls run in parallel goroutines and are joined before blending. Blends are skipped only when alpha is exactly 0/1 and when `light == 0`. Pushing a video whose source is already in a slot is allowed (logged, not an error); each push creates a fresh video instance so releases never collide. Also applies a global `light` (darken) effect and computes the frame `wait` (FPS throttle).
- **`handler.go`** — gRPC service implementation (`pb.IkascrewServer`): `Effect` (load & push a video by content ID — the type and JSON params are resolved from the **work file** (ikasbox 由来の Type/Params), so the client's `Type` field only matters as a fallback for old work files without Type; empty Params falls back to Path), `Switch` ("next"/"prev" triggers auto crossfade), `PutVolume` (index selects SWITCH=0 / LIGHT=1 / WAIT=2 mode and writes the value directly onto the Stream), `Sync` (toggles fullscreen ⇔ windowed).
- **`video_gen.go`** — thin wrapper over the `github.com/ikascrew/plugin/video` registry (canonical types "file", "img", "cd", "terminal"; `video.Normalize` absorbs legacy names like "image"/"countdown"). Params are opaque JSON strings interpreted only by each plugin.
- **`config/`** — global singleton config. `config.Create` fetches project dimensions and the content-ID→{Path,Type,Params} map from ikasbox over HTTP and writes `.server/config.json`; `config.Set` reads that work file at server startup (ikasbox is a preparation-time dependency only). Content IDs must stay consistent with the client — both sides must see the same ID mapping (re-run `create` on both after changing project contents).

Concurrency note: gRPC handlers run on separate goroutines and mutate `Stream` fields (`now_value`, `light`, `wait`, `mode`) without locking; the render loop reads them on the main thread. This data race is known and currently accepted. Video switches are serialized through the `Window.wait` channel.

## Conventions

- Errors are wrapped with `golang.org/x/xerrors` (`xerrors.Errorf` with `%w`).
- Comments and commit messages are often in Japanese; commit subjects use `fix:` / `feat:` / `perf:` / `chore:` prefixes.
- The related ikascrew repos (`core`, `plugin`, `pb`, `ikasbox`) live under the same GitHub org and are versioned as pseudo-version dependencies; protocol changes require touching `github.com/ikascrew/pb`.

## Roadmap context (from README)

Effects are planned to be redesigned: per-video `effect` (currently only the global light/wait on `Stream`) separated from `transition` (currently only the AddWeighted crossfade), with new OpenCV-based transitions (masks, bitwise ops, etc.) added after that split.
