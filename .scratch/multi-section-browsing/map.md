# Multi-section Browsing Experience

Label: `wayfinder:map`

## Destination

An implementation-ready specification for SpotUI's authenticated main screen,
with enough product, interaction, data, and compatibility decisions for
development to begin without another design phase.

## Notes

- This map produces a specification; it does not implement the feature.
- Consult `domain-modeling` whenever vocabulary changes, `research` for facts
  outside this repository, `prototype` for visual or behavioral uncertainty,
  and `codebase-design` for state and module boundaries.
- Preserve the existing Go, Bubble Tea, Lip Gloss, and in-memory go-librespot
  adapter architecture unless a ticket records a reason to change it.
- The authenticated screen has persistent `Library`, `Recommended`, and
  `Search` navigation sections. The right panel owns the active content and a
  global player.
- `Library` includes Liked Tracks, User Playlists, and Saved Albums.
- Opening a playlist, album, or artist replaces the right-panel listing;
  `Esc` or `Backspace` returns one level.
- `Recommended` shows transparent Playlist and Album groups derived from
  supported Spotify Top Artists and Top Tracks. A Spotify Home clone, Track
  radio, and infinite recommendation feeds are outside this increment.
- `Search` returns grouped Tracks, Albums, Artists, and Playlists. Enter plays
  a Track and opens details for the other result types.
- Artist details include image, name, optional genres, popular tracks when
  private metadata supplies them, and principal albums. Followers and biography
  are excluded because no supported source supplies them.
- Tracks started from a playlist or album retain that source as playback
  context and continue in its displayed order.
- The global player includes play/pause, previous/next, volume, autoplay,
  shuffle, ten-second seek, elapsed time, total duration, and progress.
- `Tab` and `Shift+Tab` change focus between navigation and content; `j`/`k`
  and arrow keys move within the focused area; playback controls remain global.
- At `80x24` and above, use a fixed-width navigation column and a content
  panel. From `50x16` through the compact range, use a single-pane flow. Below
  `50x16`, show the terminal-too-small state.
- Prefer full image rendering in compatible terminals such as Kitty, with a
  Unicode ANSI true-color rendering fallback and a textual placeholder when
  neither is available.
- Library content is read-only. Load remote sections on demand with pagination,
  session-memory caching, local loading/error states, and no offline cache.
- The UI remains in English. Startup always opens `Library`; navigation state
  is not persisted across process restarts.

## Decisions so far

- [Establish the Technical Capability Surface](issues/01-establish-technical-capability-surface.md) — use typed authenticated Catalog reads, one context-browser fork seam, contextual player extensions, and probed Kitty/ANSI artwork; derive Recommended from Top items and omit unsupported Artist fields.

## Not yet specified

- Prototype feedback may expose accessibility or terminal-layout questions not
  yet precise enough to ticket.

## Out of scope

- Liking or unliking tracks, saving or removing albums, and creating, editing,
  or deleting playlists.
- Saved or followed artists, podcasts, and audiobooks.
- Artist biographies.
- Artist follower counts, a Spotify Home clone, deprecated editorial or seed
  recommendation feeds, Track radio, infinite recommendation feeds, repeat
  modes, and an editable playback queue.
- Offline caching, navigation persistence across restarts, localization, and
  Windows support.
