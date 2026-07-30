# 13 — Keep browsing stable through failures and terminal changes

**What to build:** Harden the complete browsing experience so Catalog failures,
engine reconnection, playback transfer, image failures, terminal resizing, and
account teardown remain recoverable and understandable without corrupting
navigation or terminal output.

**Blocked by:** 06 — Render artwork across terminal capabilities; 11 — Discover
recommendations from Top items; 12 — Control contextual playback with Shuffle
and seek.

**Status:** ready-for-agent

- [ ] Each section, result group, detail subsection, page, and artwork request
      owns its loading/error/retry state; one failure cannot replace the whole
      authenticated shell.
- [ ] Late asynchronous Catalog and image responses are ignored when their
      account session, query, route, resource, or generation is no longer
      current.
- [ ] Transient engine reconnection preserves the active section, route stack,
      cached Catalog data, selections, and scroll positions while keeping
      playback controls disabled until recovery.
- [ ] Credential expiry and confirmed Logout clear account-scoped Catalog
      caches, images, navigation stacks, player context, and SpotUI-owned image
      placements before returning to the logged-out screen.
- [ ] Playback transfer preserves browsing and Catalog access while visibly
      separating transferred playback from pause, stop, and reconnect states.
- [ ] Switching between wide, compact, undersized, and restored layouts keeps a
      valid route and selection, does not redownload unchanged artwork, and
      leaves no native-image placement artifacts.
- [ ] Repeated retry, resize, navigation, reconnect, transfer, Logout, and quit
      sequences remain responsive and race-free.
- [ ] Help, README controls, and the manual validation guide describe the new
      navigation, responsive thresholds, Library categories, grouped Search,
      Recommended derivation, details, artwork tiers, Shuffle, seek, and known
      unsupported metadata.
- [ ] The root-model suite exercises cross-surface failure and recovery
      scenarios at the fake engine/catalog seam rather than private state
      arrangement.
- [ ] The complete automated suite passes with race detection and requires no
      Spotify credentials, browser, network, audio hardware, or native-image
      terminal.
- [ ] Manual Premium-account validation covers real Library data, all Search
      types, owned and non-owned Playlist details, Artist fallbacks,
      Top-item-derived Recommended groups, contextual playback, Shuffle, seek,
      transfer/reclaim, and reconnect.
- [ ] Manual terminal validation covers a positive Kitty-protocol terminal, a
      true-color fallback terminal, compact and undersized layouts, resize,
      navigation cleanup, and process shutdown cleanup.
