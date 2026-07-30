# 11 — Discover recommendations from Top items

**What to build:** Make `Recommended` a transparent, supportable discovery
surface that derives Album and Playlist groups from the listener's Spotify Top
Artists and Top Tracks without claiming to reproduce Spotify Home.

**Blocked by:** 10 — Explore Artist details.

**Status:** ready-for-agent

- [ ] Recommended loads supported Spotify Top Artists and Top Tracks through
      typed authenticated Catalog contracts.
- [ ] Album groups are derived from Top Artists and the Albums of Top Tracks;
      Playlist groups use Playlist Search associated with Top Artists.
- [ ] Every group title identifies the listening signal that produced it, such
      as a Top Artist, rather than presenting the group as Spotify editorial
      content.
- [ ] Items are deduplicated by stable URI and preserve Top-item rank before
      applying any secondary catalog ordering.
- [ ] Recommended never calls deprecated Featured Playlists, New Releases, seed
      Recommendations, or speculative private recommendation metadata.
- [ ] Empty Top-item history produces a clear empty state instead of fabricated
      generic recommendations.
- [ ] Albums and Playlists reuse their established summaries, artwork,
      attribution, details, pagination, back navigation, and contextual
      playback behavior.
- [ ] A failure in one derived group is retryable and does not hide successful
      groups or interrupt playback.
- [ ] Top-item inputs and derived groups are cached for the account session and
      cleared on Logout or credential expiry.
- [ ] Root-model tests cover derivation, transparent titles, stable ordering,
      deduplication, partial errors, empty history, entity drill-down, and cache
      clearing.
- [ ] Adapter contract tests cover supported Top-item reads and derived Search
      requests without live Spotify access.
