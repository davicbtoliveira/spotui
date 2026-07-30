# 06 — Render artwork across terminal capabilities

**What to build:** Give the browsing shell a capability-selected artwork path
that renders full covers in compatible terminals, useful ANSI/Unicode covers
elsewhere, and a stable placeholder in limited output environments.

**Blocked by:** 05 — Browse Liked Tracks in the new shell.

**Status:** ready-for-agent

- [ ] Artwork is accessed through one SpotUI renderer contract whose native,
      ANSI, and text implementations consume the same decoded image and
      allocated cell rectangle.
- [ ] Native graphics support is selected only after the normative Kitty
      protocol query succeeds before Bubble Tea competes for terminal input;
      terminal names and environment variables are not treated as proof.
- [ ] A negative response or a 150–300 ms capability-probe timeout selects the
      fallback without delaying application startup further.
- [ ] The native renderer reuses the resolved Kitty encoder and passthrough
      utilities, sends PNG data inline, preserves aspect ratio, and never
      launches `kitten icat`.
- [ ] Image and placement IDs belong to SpotUI; ordinary redraws and resizes
      reuse unchanged image data, and cleanup removes only SpotUI placements.
- [ ] A true-color fallback renders two vertical pixels per cell with Unicode
      half blocks and independent 24-bit foreground/background colors.
- [ ] Non-true-color and non-TTY output degrades through the detected color
      profile or to a fixed-size textual placeholder without corrupting layout.
- [ ] Artwork is not cropped, branded, or overlaid; the selected entity retains
      visible Spotify attribution and an action that opens its Spotify URL
      through the existing browser boundary.
- [ ] Image download or decoding failure affects only the artwork and leaves
      metadata, navigation, and playback usable.
- [ ] Image data and renditions are cached only for the process session;
      resizing does not redownload unchanged artwork.
- [ ] Protocol-level tests cover capability success/failure/timeout, compliant
      transfer chunks, placement reuse, resize, tmux fallback, and targeted
      cleanup without requiring a Kitty executable.
- [ ] Deterministic renderer tests cover true-color half blocks, odd dimensions,
      aspect-ratio preservation, limited color profiles, and missing images.
