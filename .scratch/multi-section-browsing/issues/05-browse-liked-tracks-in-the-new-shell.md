# 05 — Browse Liked Tracks in the new shell

**What to build:** Replace the authenticated single-list screen with the
responsive browsing shell and make Liked Tracks the first complete Library
path. A listener can navigate the shell, load saved Tracks, and start playback
without losing the global player.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] At `80x24` or larger, the authenticated screen shows a fixed 24-cell
      navigation column, a content panel, and the global player anchored at the
      bottom of the content region.
- [ ] From `50x16` through the compact range, the same navigation hierarchy is
      usable as a single-pane flow; smaller terminals show only the
      terminal-too-small state.
- [ ] `Library`, `Recommended`, and `Search` are visible navigation sections,
      with `Library` active after Login and after every process restart.
- [ ] Wide mode uses `Tab`/`Shift+Tab` for pane focus, `j`/`k` and Up/Down for
      movement, and `Enter` for activation; focus and selection are visibly
      distinct.
- [ ] Library exposes Liked Tracks, User Playlists, and Saved Albums, with
      Liked Tracks providing the first working content route.
- [ ] Liked Tracks load through a typed, authenticated, paginated Catalog
      contract behind the in-memory adapter; upstream response types do not
      escape into the TUI.
- [ ] Track rows show name, Artists, duration, and explicit state when
      available, with stable selection and scrolling across loaded pages.
- [ ] Loading, empty, retryable error, and loaded states remain local to Liked
      Tracks and do not hide or interrupt the player.
- [ ] Successfully loaded pages are reused during the account session, and
      Logout or credential expiry clears them.
- [ ] `Enter` on a Liked Track replaces the current queue and begins local
      playback without regressing the established transfer-reclaim behavior.
- [ ] Root-model tests cover wide, compact, undersized, focus, pagination,
      empty/error/retry, cache clearing, and Track activation behavior.
- [ ] Adapter contract tests cover authenticated saved-Track request and
      response translation without requiring network, credentials, or audio.
