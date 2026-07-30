# Third-Party Notices

SpotUI is licensed under the GNU General Public License version 3 only.
Copyright (C) 2026 SpotUI contributors. See `LICENSE` for the complete license
and warranty disclaimer.

## go-librespot

SpotUI incorporates go-librespot in-process.

- Upstream: <https://github.com/devgianlu/go-librespot>
- Upstream base revision: `bfa4a350025b9df1b5da8f568753be4f37e4bef0`
- SpotUI fork: <https://github.com/davicbtoliveira/go-librespot>
- Pinned fork revision: `c32bcebc27b56ab1ce0e29fb9d94699fe5be2ee8`
- Copyright notice shipped by upstream: Copyright (C) 2023 devgianlu
- License: GNU General Public License version 3

The SpotUI fork modifies go-librespot to expose an interactive Authorization
URL Hook, make its callback lifecycle cancelable, add a runtime-safe Autoplay
request, and expose account product events. Those changes are recorded in fork
commits `26108cb`, `589b659`, `f08561f`, and `93bc254`. The fork also exposes
native, paginated Track Search through Spotify contexts and TRACK_V4 metadata
without the public Web API; see commits `3d42788` and `c32bceb`.

The complete corresponding source for SpotUI and the modified go-librespot fork
is available at the repositories linked above. Binary distributions must
include this notice and `LICENSE`, and must provide equivalent access to the
corresponding source as required by GPL-3.0.

Other Go dependencies remain subject to their respective licenses and
copyright notices in their source distributions.
