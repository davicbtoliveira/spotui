# Multi-section Browsing Experience

Status: ready-for-agent

## Problem Statement

After Login, SpotUI presents Track Search as a single full-screen list. A
listener cannot browse their Library, discover personalized recommendations,
search for entities other than Tracks, inspect a Playlist, Album, or Artist, or
keep a stable sense of location while controlling playback. The current player
also lacks contextual playback, Shuffle, seeking, and cover artwork.

The result is a functional playback tracer bullet rather than a complete
terminal music-browsing experience. Listeners must use another Spotify client
to find most content before SpotUI can play it.

## Solution

Replace the authenticated single-list screen with a responsive browsing shell.
On wide terminals, a persistent navigation column on the left exposes
`Library`, `Recommended`, and `Search`, while the right panel shows the active
listing or detail view and retains a global player at the bottom. Compact
terminals use the same navigation hierarchy as a single-pane flow.

`Library` exposes Liked Tracks, User Playlists, and Saved Albums.
`Recommended` exposes clearly labelled Playlist and Album groups derived from
the listener's supported Spotify Top Artists and Top Tracks data; it does not
pretend to reproduce Spotify Home. `Search` returns grouped Tracks, Albums,
Artists, and Playlists. Opening an Album, Playlist, or Artist replaces the
right-panel listing with its details; the listener can return one level without
losing the previous selection or scroll position.

Cover artwork uses a native terminal image protocol when support is positively
detected. Other true-color terminals receive a Unicode half-block
representation, with a textual placeholder as the final fallback.

The global player remains usable across navigation and adds contextual
Playlist or Album playback, Shuffle, ten-second seeking, and an
elapsed/total-time display without weakening the existing standalone local
playback and Spotify Connect transfer rules.

## User Stories

1. As an authenticated listener, I want the application to open on Library, so
   that my saved music is immediately available.
2. As a listener, I want persistent access to Library, Recommended, and Search,
   so that I always understand the main navigation choices.
3. As a keyboard user, I want to see which navigation section is active, so
   that I know which content the right panel represents.
4. As a keyboard user, I want to see which panel has focus, so that movement
   keys have predictable effects.
5. As a keyboard user, I want `Tab` and `Shift+Tab` to switch between the
   navigation and content areas, so that I can move without a mouse.
6. As a keyboard user, I want `j`/`k` and the arrow keys to move within the
   focused area, so that existing SpotUI movement conventions remain useful.
7. As a listener, I want the player to remain visible while I browse, so that
   navigation does not hide playback state.
8. As a listener, I want playback controls to remain global, so that I do not
   need to focus the player before controlling audio.
9. As a listener, I want Library to separate Liked Tracks, User Playlists, and
   Saved Albums, so that each kind of saved content is easy to find.
10. As a listener, I want Library lists to load only when needed, so that Login
    is not blocked by every remote collection.
11. As a listener with a large Library, I want paginated loading, so that the
    terminal stays responsive and memory use remains bounded.
12. As a listener, I want an empty Library category to explain that no items
    are saved, so that an empty screen is not mistaken for a failure.
13. As a listener, I want a Playlist row to show its cover, name, owner or
    curator when available, and item count, so that I can identify it before
    opening it.
14. As a listener, I want Playlist details to show its cover, name, owner or
    curator, description when available, track count, and Tracks, so that I can
    understand its contents before playing.
15. As a listener, I want an Album row to show its cover, name, Artists, and
    release year when available, so that similar releases are distinguishable.
16. As a listener, I want Album details to show its cover, name, Artists,
    release information, and ordered Tracks, so that I can browse the complete
    release.
17. As a listener, I want Track rows to show name, Artists, duration, and
    explicit-content state when available, so that I can choose the intended
    recording.
18. As a listener, I want opening a Playlist or Album to replace the current
    listing in the content panel, so that the interface does not require a
    third column.
19. As a listener, I want `Esc` or `Backspace` to return from details, so that
    drill-down navigation is easy to reverse.
20. As a listener, I want returning from details to restore the previous
    selection and scroll position, so that I do not lose my place.
21. As a listener, I want selecting a Track within a Playlist or Album to start
    at that Track and continue through the same context, so that playback
    follows the source I was browsing.
22. As a listener, I want contextual playback to preserve the displayed order
    when Shuffle is off, so that Album and Playlist sequencing is predictable.
23. As a listener, I want Recommended to show Playlist and Album groups derived
    from my Spotify Top Artists and Top Tracks, so that I can discover relevant
    content without depending on a removed Home-feed API.
24. As a listener, I want each Recommended group to have a visible title, so
    that I understand which listening signal produced it.
25. As a listener, I want a recommended Playlist or Album to use the same detail
    flow as Library, so that entity behavior is consistent across sections.
26. As a listener, I want Search to accept a text query in the right panel, so
    that the navigation column remains stable while I type.
27. As an existing SpotUI user, I want `/` to activate Search input directly,
    so that the established shortcut remains available.
28. As a listener, I want Search results grouped into Tracks, Albums, Artists,
    and Playlists, so that mixed result types remain understandable.
29. As a listener, I want each Search result group to expose its own empty or
    unavailable state, so that one missing type does not make the whole query
    look broken.
30. As a listener, I want `Enter` on a Track Search Result to replace the
    current queue and play that Track immediately, so that established Search
    playback behavior is preserved.
31. As a listener, I want `Enter` on an Album, Artist, or Playlist Search Result
    to open details instead of starting arbitrary playback, so that the action
    matches the selected entity.
32. As a listener, I want Artist details to show an image, name, optional
    genres, popular Tracks, and principal Albums when available, while omitting
    unavailable facts instead of displaying fabricated values, so that I can
    explore an Artist without misleading metadata.
33. As a listener, I want an Artist's popular Track to play immediately, so
    that Artist discovery leads directly to playback.
34. As a listener, I want an Artist's Album to open the normal Album detail
    view, so that duplicate detail behavior is avoided.
35. As a listener, I want remote loading to show a local progress state in the
    affected panel, so that the rest of SpotUI remains usable.
36. As a listener, I want a remote error to stay within the affected section or
    detail view, so that playback and unrelated cached content remain usable.
37. As a listener, I want a failed load to offer a retry action, so that a
    transient failure does not require restarting or logging in again.
38. As a listener, I want successfully loaded browsing data cached for the
    current session, so that returning to a section is immediate.
39. As a listener, I want Logout to clear all browsing caches and selected
    account content, so that another account cannot see the previous account's
    data.
40. As a listener, I want transient engine reconnection to preserve the current
    browsing location when safe, so that a playback connection failure does not
    reset unrelated navigation.
41. As a listener, I want the player to display Track, Artists, Album, playback
    state, volume, Autoplay, Shuffle, elapsed time, total duration, and progress,
    so that playback is understandable at a glance.
42. As a listener, I want play/pause, previous/next, volume, and Autoplay
    controls to keep their existing shortcuts, so that the expanded interface
    does not invalidate learned controls.
43. As a listener, I want a global shortcut to toggle Shuffle for the active
    Playlist or Album context, so that I can change ordering while browsing
    elsewhere.
44. As a listener, I want global shortcuts to seek backward or forward by ten
    seconds, so that I can revisit or skip a short passage.
45. As a listener, I want seeking to clamp at the beginning and end of the
    Track, so that a shortcut never produces an invalid position.
46. As a listener, I want engine events to update Shuffle, seek position, and
    player state, so that the display reflects actual playback rather than an
    optimistic command.
47. As a listener, I want playback controls disabled after playback transfers
    to another Spotify Connect device, so that SpotUI does not become a remote
    controller.
48. As a listener, I want choosing a Track after transfer to reactivate local
    SpotUI playback, so that established transfer recovery remains intact.
49. As a listener in a compatible terminal, I want the unmodified cover image
    rendered in its allocated cells with a Spotify attribution/link action, so
    that Albums, Playlists, Artists, and the current Track are recognizable
    without losing their source.
50. As a listener in a true-color terminal without native image support, I want
    a colored Unicode rendering of the cover, so that artwork remains useful.
51. As a listener in a limited terminal, I want a stable textual artwork
    placeholder, so that missing graphics do not corrupt the layout.
52. As a listener using SSH or a terminal multiplexer, I want image capability
    detection to fall back safely when graphics commands do not traverse the
    session, so that the TUI remains readable.
53. As a listener resizing a supported terminal, I want existing cover data to
    be repositioned without visible image leaks or repeated downloads, so that
    resizing remains smooth.
54. As a listener leaving a detail view or exiting SpotUI, I want only SpotUI's
    image placements removed, so that graphics from other terminal applications
    are untouched.
55. As a listener with a terminal at least `80x24`, I want a fixed navigation
    column beside the content panel, so that browsing and location stay visible
    together.
56. As a listener with a terminal between the compact minimum and the wide
    threshold, I want a single-pane navigation flow, so that the same features
    remain usable without crushed columns.
57. As a listener with a terminal smaller than `50x16`, I want a clear
    terminal-too-small message, so that corrupted output is avoided.
58. As a listener restarting SpotUI, I want the application to return to
    Library, so that stale navigation state is not persisted.
59. As a listener, I want all new interface text in English, so that the
    expanded screen remains consistent with the existing application.
60. As a Spotify Free account holder, I want the existing unsupported-account
    screen to remain ahead of all browsing views, so that unavailable playback
    features are not presented as usable.

## Implementation Decisions

- The feature extends the authenticated Bubble Tea root model rather than
  creating a second application lifecycle. Logged-out, authenticating,
  reconnecting, unsupported-account, Logout, and quit behavior remain
  authoritative outside the browsing shell.
- The root state distinguishes the active navigation section, focused pane,
  current content route, a detail-navigation stack, per-route selection and
  scroll position, per-resource load state, session cache, and current player
  state.
- `Library`, `Recommended`, and `Search` are canonical navigation section
  names. They are not independent application tabs or separate player
  instances.
- The wide layout is active only when width is at least 80 cells and height is
  at least 24 rows. Its navigation column occupies a fixed 24 cells, including
  its divider; the content panel receives the remaining width.
- The compact layout is active when the terminal is at least `50x16` but does
  not meet both wide thresholds. It shows either navigation or content as one
  pane and uses the same route stack and back action as the wide layout.
- Any dimension below `50x16` renders only the existing
  terminal-too-small-style state. Loading and playback continue without
  accepting layout-specific navigation actions.
- The player is global to the authenticated shell and anchored to the bottom of
  the content region. It is not a focusable pane; its shortcuts are handled
  before focused-list navigation.
- `Tab` and `Shift+Tab` alternate navigation/content focus in the wide layout.
  In compact mode, `Enter` moves from navigation to content and `Esc` or
  `Backspace` returns through detail routes and then to navigation.
- `j`/`k` and Up/Down move selection. `Enter` activates the selected item.
  Existing `/`, Space, `n`, `p`, `-`, `+`, `a`, `L`, `?`, `q`, and `Ctrl+C`
  behavior is preserved. `s` toggles Shuffle; `h` and `l` seek backward and
  forward by ten seconds. Existing previous/next Search page shortcuts remain
  available where pagination is explicit.
- Search input owns printable runes, Space, Backspace, Enter, and Esc while
  active. Global quit remains available; playback letter shortcuts do not fire
  while text input is active.
- The detail-navigation stack records entity identity, source section, source
  query where applicable, selection, scroll offset, and pagination cursor.
  Back navigation restores this state rather than refetching or resetting the
  parent listing.
- Every browsable entity has a stable Spotify URI and a display record owned by
  SpotUI. Track records include name, Artists, Album identity, duration, image
  reference, and optional explicit state. Playlist, Album, and Artist records
  contain only the summary fields needed by listings; detail records add their
  richer metadata and child collections.
- `Library` is defined as Liked Tracks, User Playlists, and Saved Albums. A User
  Playlist is any Playlist present in the listener's Library, whether owned or
  followed. All three surfaces are read-only.
- `Recommended` contains two supported, transparent personalization families:
  Albums associated with the listener's Spotify Top Artists or the Albums of
  Top Tracks, and Playlist Search results associated with those Top Artists.
  Groups state the source signal in their title, deduplicate by stable URI, and
  preserve Top-item rank before any secondary catalog ordering. This is a
  SpotUI-derived view of supported Spotify personalization, not a clone of
  Spotify Home.
- No implementation may depend on deprecated Featured Playlists, New Releases,
  seed Recommendations, or undocumented `RECOMMENDED_PLAYLISTS` and
  `POPULAR_RELEASES` metadata identifiers. If Top-item data is empty or
  unavailable, `Recommended` shows an honest section-local empty/error state.
- Read-only Catalog data uses typed methods over the pinned fork's authenticated
  in-memory Web API request. Credentials and generic JSON stay inside the
  adapter. This does not introduce a second OAuth flow, developer-dashboard
  setup, distributed client secret, or direct Web API dependency in the TUI.
- Unified Search accepts one query and returns separately typed, paginated
  result groups for Tracks, Albums, Artists, and Playlists through one
  authenticated grouped Search request. Each type has an independent page and
  a maximum page size of 10 under the supported API contract.
- Playlist metadata uses the authenticated read-only Catalog seam. Playlist
  items for non-owned, non-collaborative Playlists use one narrow fork
  extension that pages an arbitrary context through the existing context
  resolver and metadata loader. The TUI receives normalized Track records and
  never protocol objects.
- Artist follower count is not part of the domain contract because the
  supported Spotify response removed it and the private Artist metadata does
  not supply it. Genres are optional and omitted when empty. Popular Tracks use
  the private Artist metadata extension with an explicit empty-state fallback;
  principal Albums use the supported Artist Albums catalog read.
- Bubble Tea commands invoke only stable SpotUI engine/catalog interfaces.
  Unofficial request types, protobufs, context resolvers, metadata loaders, and
  fork-only hooks remain confined to the in-memory adapter boundary established
  by the accepted architecture.
- Catalog reads are asynchronous and identify the requested resource and page.
  Completion messages carry the same request identity so stale responses from a
  previous route or query can be ignored.
- Every remote resource uses explicit `idle`, `loading`, `loaded`, `empty`, and
  `error` states. An error retains already loaded pages and exposes retry for
  the failed request rather than clearing the full section.
- Successful pages and detail records are cached in memory by account session
  and stable URI. The cache survives ordinary navigation and transient engine
  reconnection, but Logout and credential expiry clear it. No catalog cache is
  written to disk.
- Image references are fetched and decoded independently from catalog records.
  A failed image load does not fail the entity or its Track listing.
- Artwork rendering is represented by one capability-selected renderer with
  native-image, ANSI half-block, and text-placeholder implementations. Layout
  code allocates a cell rectangle and does not emit terminal-specific escape
  sequences itself.
- Native graphics support is established by the normative 1x1 `a=q` protocol
  query performed before Bubble Tea begins competing for terminal input.
  Environment variables and terminal names are hints only; without a positive
  response before a 150–300 ms timeout, SpotUI selects the fallback.
- The native renderer implements the Kitty graphics protocol directly rather
  than launching `kitten icat`. It reuses the resolved Kitty encoder and tmux
  passthrough utilities already present in the dependency graph. PNG data is
  sent inline in compliant chunks, image data and placements use SpotUI-owned
  IDs, redraw reuses existing image data, and cleanup targets only SpotUI
  placements.
- A native protocol response through SSH or a multiplexer permits native
  rendering. A negative response, timeout, or blocked passthrough selects the
  ANSI renderer; GNU Screen receives no special-case assumption of support.
- The ANSI renderer samples the decoded cover into the allocated cell grid and
  uses upper/lower half-block glyphs with independent 24-bit foreground and
  background colors. When true-color output is unavailable, the renderer uses a
  fixed-size textual placeholder that preserves geometry.
- Image download results may be cached in memory by image identity and target
  rendition. Resizing may regenerate placement or ANSI output but must not
  redownload unchanged artwork.
- Catalog records retain their external Spotify URL. Views that display Spotify
  artwork also display compact attribution and allow the selected entity to be
  opened through the existing browser-opening boundary. Artwork preserves
  aspect ratio and is never cropped, branded, or overlaid with controls.
- Track activation from Search preserves ADR 0021: it replaces the current
  queue and plays the isolated Track immediately. Track activation from a
  Playlist or Album instead supplies both the context URI and selected Track
  URI so playback begins at that Track and continues through the context.
- Shuffle is a runtime player operation and an engine-event-sourced state. It
  applies to the active Playlist or Album context, preserves the current Track
  when toggled, and is unavailable when no shuffle-capable local context is
  active.
- Seeking uses the fork's relative seek operation with `-10000` or `+10000`
  milliseconds. The player clamps the result, emits the resulting position,
  and seeking while paused remains paused.
- Player events remain the source of truth for Track, context, progress,
  playing/paused/buffering state, volume, Autoplay, Shuffle, active-device
  state, and transfer state. Command handling does not permanently mutate these
  values before confirmation.
- When playback is transferred away, navigation and catalog reads remain
  available, while play/pause, previous/next, volume, Autoplay, Shuffle, and
  seek do not control the remote device. Selecting a Track reclaims local
  playback under the existing transfer decision.
- Remote Catalog and image failures are scoped to their owning view. Playback
  engine failures continue through the existing reconnect state and must not
  expose partially interactive playback controls.
- Help text and the README control table are updated with focus, back, Shuffle,
  seek, section navigation, and responsive-layout behavior.

## Testing Decisions

- Tests assert externally visible behavior: rendered content, focus and route
  transitions, emitted engine/catalog operations, event-driven player state,
  fallback selection, and error recovery. They do not assert private field
  arrangement, renderer helper calls, or upstream protobuf shapes outside the
  adapter contract.
- The primary automated seam is the root Bubble Tea model with a fake
  engine/catalog implementation. Tests send window-size messages, key messages,
  load-result messages, image-capability results, and player events, then assert
  rendered output and recorded operations.
- Root-model scenarios cover wide and compact navigation, minimum-size
  behavior, focus changes, Search typing, grouped results, detail drill-down and
  restoration, pagination, independent loading/error states, cached returns,
  Logout clearing, contextual playback, Shuffle, seeking, transfer disabling,
  and transient reconnection.
- The fake engine/catalog records typed requests and can independently return
  pages, details, images, and errors. This extends the same prior art currently
  used for Track Search, playback commands, Autoplay, Logout, and reconnection.
- Adapter contract tests cover only the unstable boundary: authenticated typed
  Catalog reads, grouped Search with independent pages, arbitrary-context Track
  browsing, response-to-domain translation, pagination cursors, metadata
  hydration, image identity, context-plus-Track playback, Shuffle, relative
  seek, and event translation. Tests use in-memory request/event channels and
  require no Spotify credentials, audio device, browser, or network.
- Renderer tests treat the ANSI implementation as a deterministic transform
  from decoded pixels and cell dimensions to styled Unicode output. They cover
  odd heights, tiny allocations, transparent or missing images, true-color
  absence, and resize regeneration.
- Native-image renderer tests assert protocol-level output: positive/negative
  capability responses, timeout fallback, chunk boundaries, placement reuse,
  resize without retransmission, and targeted cleanup. They must not depend on a
  locally installed Kitty executable.
- View-level tests may be used for stable help text and small pure formatting
  behavior, following the existing help and player renderer tests. They do not
  replace root-model interaction coverage.
- The full automated suite continues to run without Spotify credentials,
  browser, network, audio hardware, or native graphics support.
- Manual Premium-account validation covers real Liked Tracks, User Playlists,
  Saved Albums, Top-item-derived Recommended groups, all four Search result
  types, owned and non-owned Playlist details, entity metadata, pagination,
  contextual playback, Shuffle, seek while playing and paused, and Spotify
  Connect transfer/reclaim.
- Manual graphics validation covers at least one Kitty-protocol terminal, one
  true-color terminal without native images, terminal resizing, SSH or tmux
  fallback behavior where available, and cleanup on navigation and shutdown.
- Existing race-enabled tests remain required because catalog requests, engine
  events, image loading, progress ticks, and terminal input run concurrently.

## Out of Scope

- Liking or unliking Tracks.
- Saving or removing Albums.
- Creating, renaming, reordering, editing, following, unfollowing, or deleting
  Playlists.
- Saved or followed Artists as a Library category.
- Podcasts, podcast episodes, audiobooks, local files, and video.
- Artist biographies or third-party editorial content.
- Track radio and an infinite recommendation feed. Existing Autoplay
  recommendations remain supported as playback behavior.
- A Spotify Home clone, deprecated Featured Playlists, deprecated New Releases,
  deprecated seed Recommendations, or speculative private recommendation
  metadata.
- Repeat modes and an editable queue.
- Remote control of another Spotify Connect device.
- Offline Catalog or image caching and offline browsing.
- Persisting the active section, route stack, selection, or scroll position
  across SpotUI restarts.
- Localization; all new strings are English.
- Windows support.
- A localhost HTTP or WebSocket control server.
- An official Spotify Web API client, developer-dashboard setup, distributed
  client secret, or separate authentication flow.

## Further Notes

- Spotify Premium remains required. The authenticated browsing shell is never
  shown to a Spotify Free account.
- SpotUI uses an unofficial, revision-pinned go-librespot fork. Spotify protocol
  changes may break Catalog surfaces even when local application contracts stay
  stable.
- The existing in-memory adapter is the architectural containment boundary for
  every new native Catalog or playback extension.
- Full-image support is capability based, not Kitty-name based: any terminal
  that positively implements the protocol can use it, and Kitty-like
  environment variables alone are insufficient.
- The technical capability research is recorded on branch
  `research/spotui-technical-capabilities` at commit
  `c242a5cfa44e0ba30bb0eaf52d2aa101b6c60348`. It confirms the read-only Catalog,
  grouped Search, contextual playback, Shuffle, relative seek, and two-tier
  artwork seams and documents the unavailable personalization and Artist
  fields reflected above.
- The accepted Login, session, Logout, local audio, reconnect, Autoplay,
  immediate Search playback, and playback-transfer decisions remain in force
  unless this specification explicitly refines their authenticated-screen
  presentation.
