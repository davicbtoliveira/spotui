# Public Playlist Discovery

Label: wayfinder:map

## Destination

SpotUI exposes Public Playlists in grouped Search results, Recommended, and
artist details, while the Library remains the home of User Playlists. Opening
any playlist uses the existing album-like detail and context-playback flow.

## Notes

Use the `Public Playlist` and `User Playlist` terms from `CONTEXT.md`.
Catalog reads must preserve the standalone authenticated-client architecture:
the current implementation intentionally avoids the quota-bound Spotify Web
API. Search already has a Playlist result group and Recommended already has a
playlist section, but neither is populated by the adapter.

## Decisions so far

<!-- Closed decision tickets are indexed here. -->

## Not yet specified

- The available authenticated Spotify-native discovery endpoints and the
  result semantics they provide for search, personalized/editorial feeds, and
  artist-associated playlists.
- The fallback experience when an individual discovery surface is unavailable
  or returns partial data.

## Out of scope

- Saving, following, editing, or otherwise mutating Public Playlists.
