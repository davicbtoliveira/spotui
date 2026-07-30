# Establish the Technical Capability Surface

Type: research
Status: resolved

## Question

Which capabilities exposed by the pinned go-librespot fork, its first-party
protocol sources, and suitable terminal image specifications can support the
confirmed Library, Recommended, unified Search, detail metadata, contextual
playback, shuffle, seek, and two-tier cover rendering experience; where are the
gaps and safest extension seams?

## Answer

The decision-ready research is recorded on branch
`research/spotui-technical-capabilities` at commit
`c242a5cfa44e0ba30bb0eaf52d2aa101b6c60348` in
`docs/research/multi-section-browsing-capabilities.md`.

Use the pinned fork's authenticated in-memory Web API request behind typed
SpotUI read-only Catalog contracts for Library, grouped Search, and supported
details. Add one narrow fork context-browser request for arbitrary Playlist
items, plus adapter operations for contextual playback, Shuffle, and relative
seek. Use the existing Kitty encoder utilities with an active protocol probe,
then fall back to ANSI half-block rendering.

There is no stable personalized Spotify Home feed of Playlist and Album groups.
`Recommended` must derive transparent groups from supported Top Artists and Top
Tracks data or show an unavailable state. Artist follower count is unavailable;
genres are optional and unreliable; Artist popular Tracks require the private
Artist metadata seam and an empty fallback.
