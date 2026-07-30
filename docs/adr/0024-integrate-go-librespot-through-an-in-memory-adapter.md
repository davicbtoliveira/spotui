# Integrate go-librespot Through an In-Memory Adapter

## Status

Accepted

## Context

go-librespot's `daemon.App` already coordinates authentication, reusable
credentials, Spotify Connect state, context loading, local audio, controls, and
events. It accepts an `ApiServer` interface whose request and event channels do
not require an HTTP listener.

Coupling Bubble Tea models directly to daemon request types would spread an
unstable unofficial dependency across the application.

## Decision

SpotUI will run `daemon.App` in-process behind `internal/spotengine`.

`internal/spotengine` will:

- implement an in-memory `daemon.ApiServer`;
- translate typed SpotUI commands into go-librespot API requests;
- translate go-librespot events into stable SpotUI player events;
- expose context-aware methods for Login lifecycle, Track Search, play, pause,
  next, previous, volume, Autoplay, status, and shutdown;
- own the go-librespot configuration and StateStore;
- isolate fork-only OAuth URL and runtime Autoplay additions.

No localhost REST or WebSocket control server will run.

Track Search will initially use go-librespot's authenticated Web API passthrough
inside this adapter. Playback will use go-librespot context and player requests,
not Spotify Web API playback endpoints.

## Amendment: Native Track Search

The initial Web API passthrough returned HTTP 429 because its application quota
is shared. Track Search now resolves a `spotify:search:` context, follows its
native pages, and loads TRACK_V4 metadata through go-librespot. This keeps
search on the same authenticated protocol as playback and removes the public
Web API dependency from the SpotUI search path.

Bubble Tea commands will call only SpotUI interfaces. Player events will enter
the TUI as messages and remain the source of truth.

## Consequences

- The application remains one process and one installed executable.
- Unofficial protocol details have one replacement boundary.
- Unit tests can use a fake engine without Spotify or audio hardware.
- Search response translation and go-librespot request shapes need contract
  tests.
- The existing `zmb3/spotify` client, client-ID resolver, official OAuth code,
  and polling playback path can be removed after feature replacement.
- The dependency will be pinned to an exact revision during the fork period.
