# 08 — Browse and play Saved Albums

**What to build:** Make Saved Albums a complete Library route where a listener
can browse saved releases, inspect ordered Tracks, return without losing
position, and start contextual Album playback at any selected Track.

**Blocked by:** 05 — Browse Liked Tracks in the new shell.

**Status:** ready-for-agent

- [ ] Saved Albums load on demand through a typed, authenticated, paginated
      Catalog contract.
- [ ] Album summaries show stable identity, cover reference, name, Artists,
      release year when available, and Spotify URL.
- [ ] Opening an Album replaces the content listing with its cover, name,
      Artists, release information, disc/Track ordering, and paginated Tracks.
- [ ] Multi-disc releases preserve disc and Track sequence and distinguish
      duplicate Track names by stable URI.
- [ ] `Esc` or `Backspace` returns to Saved Albums with the previous page,
      selection, and scroll position restored.
- [ ] Selecting an Album Track starts that Album context at the chosen Track and
      Next/Previous continue in release order while Shuffle is off.
- [ ] Loading, empty, unavailable-Track, and retryable error states remain local
      to the Album list or detail route.
- [ ] Album summaries, details, and Track pages are cached for the account
      session and cleared on Logout or credential expiry.
- [ ] Root-model tests cover Album pagination, multi-disc ordering, drill-down
      restoration, local errors, and contextual playback.
- [ ] Adapter contract tests cover saved-Album, Album-detail, Album-Track, and
      context-plus-selected-Track translation without live Spotify access.
