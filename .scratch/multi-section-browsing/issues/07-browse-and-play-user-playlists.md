# 07 — Browse and play User Playlists

**What to build:** Make User Playlists a complete Library route where a listener
can browse owned and followed Playlists, inspect their Tracks, return without
losing position, and start contextual playback at any selected Track.

**Blocked by:** 05 — Browse Liked Tracks in the new shell.

**Status:** ready-for-agent

- [ ] User Playlists load on demand through a typed, authenticated, paginated
      Catalog contract and include both owned and followed Playlists.
- [ ] Playlist summaries show stable identity, cover reference, name, owner or
      curator when available, item count, and Spotify URL.
- [ ] Opening a Playlist replaces the content listing with its cover, name,
      owner or curator, optional description, Track count, and paginated Tracks.
- [ ] Owned and collaborative Playlist Tracks use the supported Catalog read;
      an arbitrary non-owned public Playlist uses one typed fork
      context-browser operation instead of leaking resolver types into SpotUI.
- [ ] Null, unavailable, or unplayable Playlist items are skipped or rendered
      unavailable without shifting the identity of playable rows.
- [ ] `Esc` or `Backspace` returns to User Playlists with the previous page,
      selection, and scroll position restored.
- [ ] Selecting a Playlist Track starts that Playlist context at the chosen
      Track and Next/Previous continue inside the context in displayed order
      while Shuffle is off.
- [ ] Loading, empty, partial-page, and retryable errors stay inside the
      Playlist list or detail route and do not interrupt existing playback.
- [ ] Playlist metadata and Track pages are cached for the account session and
      cleared on Logout or credential expiry.
- [ ] Root-model tests cover owned, followed, and non-owned Playlist flows,
      drill-down restoration, pagination, local errors, and contextual play.
- [ ] Adapter and fork contract tests cover Playlist translation, arbitrary
      context paging, and context-plus-selected-Track playback without live
      Spotify access.
