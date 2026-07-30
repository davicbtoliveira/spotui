# 12 — Control contextual playback with Shuffle and seek

**What to build:** Complete the global player for Album and Playlist contexts
with event-confirmed Shuffle and ten-second seeking while preserving existing
local-player and Spotify Connect transfer rules.

**Blocked by:** 07 — Browse and play User Playlists; 08 — Browse and play Saved
Albums.

**Status:** ready-for-agent

- [ ] The global player shows current Track, Artists, Album, playback state,
      volume, Autoplay, Shuffle, elapsed time, total duration, and progress
      throughout navigation.
- [ ] `s` toggles Shuffle only when a shuffle-capable local Album or Playlist
      context is active.
- [ ] Shuffle is performed by the engine context rather than by reordering
      visible DTOs, preserves the current Track, and changes visually only after
      the engine reports its resulting state.
- [ ] `h` and `l` issue relative seek operations of `-10000` and `+10000`
      milliseconds while no text input is active.
- [ ] The player clamps relative seek at Track boundaries, reports the resulting
      position through events, and remains paused when seeking from pause.
- [ ] Buffering, Track transition, pause/resume, and stale progress ticks do not
      overwrite a newer engine-reported seek position.
- [ ] Existing Space, `n`, `p`, `-`, `+`, and `a` shortcuts and their status
      feedback remain unchanged.
- [ ] After playback transfers away, play/pause, previous/next, volume,
      Autoplay, Shuffle, and seek do not control the remote device; choosing a
      Track can still reclaim local playback.
- [ ] Help content describes Shuffle, seek, global controls, contextual
      behavior, and transfer limitations.
- [ ] Root-model tests cover Shuffle availability and round-trip state, relative
      seek while playing/paused, boundary clamping, progress reconciliation,
      text-input suppression, and transferred playback.
- [ ] Adapter and fork contract tests cover contextual play, Shuffle commands
      and events, relative seek, and seek events without live Spotify access.
