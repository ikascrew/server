# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Video server component of the **ikascrew** VJ (video jockey) system. It renders video to an OpenCV window and is controlled remotely over gRPC (effect switching, transitions, volume/light/wait parameters). Content metadata comes from a separate `ikasbox` HTTP service.

## Build & Run

Requires OpenCV installed locally (gocv dependency). There are no tests in this repo.

```
go build ./...                 # build everything
go run ./cmd/ika-server <project-id>   # run the server (needs ikasbox running)
go run ./cmd/check <project-id>        # manual smoke test: cycles through all contents with transitions
```

Prerequisite: an ikasbox server must be reachable (default `localhost:5555`); the server fetches project/content lists from `http://<DBIP>:<DBPort>/project/content/list/<project-id>` at startup. gRPC listens on port 55555 by default (both configurable via `config.Option` functions in `config/option.go`).

## Architecture

Startup flow (`server.go: Start`): load config from ikasbox → start a UDP multicast announcer (`ikascrew/core/multicast`) → create the initial "terminal" video (system-info splash from `opening.go`) → create a `Window` → start the gRPC server in a goroutine → `Window.Play` blocks forever on the main thread.

Key pieces and how they interact:

- **`Window` (window.go)** — owns the gocv window and the render loop. `Play` must run on a locked OS thread (OpenCV requirement). It selects between a channel of pushed videos (`Push`, called from gRPC handlers) and rendering the current frame each tick.
- **`Stream` (stream.go)** — the transition engine. Holds three video slots (`now_video`, `old_video`, `release_video`) that rotate on each `Switch`; frames are blended with `gocv.AddWeighted` using `now_value`/`old_value` (0–200, `SWITCH_VALUE`) as crossfade alpha. Also applies a global `light` (darken) effect and computes the frame `wait` (FPS throttle). The `used` map prevents pushing a video that is still in one of the three slots.
- **`handler.go`** — gRPC service implementation (`pb.IkascrewServer`): `Effect` (load & push a video by content ID; forces type to "img" for image extensions), `Switch` ("next"/"prev" triggers auto crossfade), `PutVolume` (index selects SWITCH=0 / LIGHT=1 / WAIT=2 mode and writes the value directly onto the Stream), `Sync` (fullscreens the window).
- **`video_gen.go`** — factory mapping type strings ("file", "img", "cd", "terminal") to `core.Video` implementations from `github.com/ikascrew/plugin`.
- **`config/`** — global singleton config (`config.Set` / `config.Get`); `load` pulls project dimensions and the content-ID→path map from ikasbox over HTTP.

Concurrency note: gRPC handlers run on separate goroutines and mutate `Stream` fields without locking; the render loop reads them on the main thread. Video switches are serialized through the `Window.wait` channel.

## Conventions

- Errors are wrapped with `golang.org/x/xerrors` (`xerrors.Errorf` with `%w`).
- Comments and commit messages are often in Japanese.
- The related ikascrew repos (`core`, `plugin`, `pb`, `ikasbox`) live under the same GitHub org and are versioned as pseudo-version dependencies; protocol changes require touching `github.com/ikascrew/pb`.
