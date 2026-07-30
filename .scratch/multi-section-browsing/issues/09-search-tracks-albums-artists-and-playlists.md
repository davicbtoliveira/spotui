# 09 — Search Tracks, Albums, Artists, and Playlists

**What to build:** Replace Track-only Search with one grouped Search experience
where Tracks play immediately and Albums, Artists, and Playlists open the
appropriate browseable entity route.

**Blocked by:** 07 — Browse and play User Playlists; 08 — Browse and play Saved
Albums.

**Status:** ready-for-agent

- [ ] Activating `Search` or pressing `/` focuses a query input in the content
      panel without removing the persistent navigation or global player.
- [ ] While Search input is active, printable characters, Space, Backspace,
      Enter, and Esc edit or submit the query without triggering playback
      letter shortcuts.
- [ ] One authenticated grouped Search request returns independently paginated
      Tracks, Albums, Artists, and Playlists with a maximum of 10 results per
      type and no upstream JSON types in the TUI.
- [ ] The four result groups have clear headings and independent loaded, empty,
      loading-more, and retryable error states.
- [ ] A newer query supersedes older requests; late responses cannot overwrite
      the current query or route.
- [ ] `Enter` on a Track preserves established behavior by replacing the
      current queue and starting that Track immediately.
- [ ] `Enter` on an Album or Playlist opens the existing detail route,
      including support for non-owned public Playlist Tracks.
- [ ] `Enter` on an Artist opens an initial Artist route with stable identity,
      name, image reference, and Spotify URL ready for richer exploration.
- [ ] Returning from a Search-opened detail restores the query, active group,
      page, selection, and scroll position.
- [ ] Search results and pages are cached by normalized query for the account
      session and cleared on Logout or credential expiry.
- [ ] Root-model tests cover typing, global-key suppression, grouped results,
      independent pagination/errors, stale responses, entity routing, state
      restoration, and immediate Track playback.
- [ ] Adapter contract tests cover grouped request limits and normalized
      translation for all four result types without live Spotify access.
