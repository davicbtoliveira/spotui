# Third-Party Notices

SpotUI is licensed under the GNU General Public License version 3 only.
Copyright (C) 2026 SpotUI contributors. See `LICENSE` for the complete license
and warranty disclaimer.

## go-librespot

SpotUI incorporates go-librespot in-process.

- Upstream: <https://github.com/devgianlu/go-librespot>
- Upstream base revision: `bfa4a350025b9df1b5da8f568753be4f37e4bef0`
- SpotUI fork: <https://github.com/davicbtoliveira/go-librespot>
- Pinned fork revision: `552a9e503e480c6892dbb9a96237527ef2b96b2d`
- Copyright notice shipped by upstream: Copyright (C) 2023 devgianlu
- License: GNU General Public License version 3

The SpotUI fork modifies go-librespot to expose an interactive Authorization
URL Hook, make its callback lifecycle cancelable, add a runtime-safe Autoplay
request, and expose account product events. Those changes are recorded in fork
commits `26108cb`, `589b659`, `f08561f`, and `93bc254`. The fork also exposes
native, paginated Track Search through Spotify contexts and TRACK_V4 metadata
without the public Web API; see commits `3d42788` and `c32bceb`. Commit
`552a9e5` pins the PulseAudio restart correction used during track transitions.

## pulse

SpotUI uses a patched PulseAudio client through go-librespot.

- Upstream: <https://github.com/jfreymuth/pulse>
- Upstream base revision: `d6270f1`
- SpotUI fork: <https://github.com/davicbtoliveira/pulse>
- Pinned fork revision: `82efa3cf17f7ce14730a6f2629afc304840bc30f`
- License: MIT

The fork prevents playback restarts from waiting forever when a PulseAudio
server, including WSLg, omits its started event after an underflow. See commits
`5c744c2` and `82efa3c`.

The complete corresponding source for SpotUI and the modified go-librespot and
pulse forks is available at the repositories linked above. Binary distributions
must include this notice and `LICENSE`, and must provide equivalent access to
the corresponding source as required by GPL-3.0.

Other Go dependencies remain subject to their respective licenses and
copyright notices in their source distributions.
