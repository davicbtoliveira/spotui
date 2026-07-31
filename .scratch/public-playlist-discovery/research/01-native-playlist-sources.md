# Native Public Playlist Discovery Sources

Date: 2026-07-31
Status: researched — no implementation-safe discovery source exists in the pinned native surface

## Answer

The authenticated native protocols currently callable from the pinned go-librespot fork can **open and paginate the tracks of a known playlist URI** and can list the listener's existing User Playlists. They do **not** provide an implemented, paginated source of Public Playlist summaries for native search, personalized/editorial Recommended, or artist-associated playlists.

Consequently, this feature cannot be implemented faithfully under the ticket's constraint by wiring the existing native calls differently. It needs a separate protocol-discovery spike with captured authenticated responses and contract tests, or a deliberate product/architecture decision to permit a Spotify Web API source. The latter conflicts with the accepted no-Web-API catalog policy.

## Evidence from the pinned implementation

| Surface | Callable native protocol today | What it returns | Result |
| --- | --- | --- | --- |
| Playlist detail and playback | `GET /context-resolve/v1/<spotify:playlist:…>` plus its native context pages | Context metadata and paginated **tracks** for a known URI. The fork translates title, description, image URL, owner name and track count when metadata is present. | Suitable *after* discovery; not a discovery source. [Pinned implementation](../../../third_party/go-librespot/daemon/native_catalog.go) |
| Search | `spotify:search:<query>` resolved through `/context-resolve/v1/…` | Only `ContextTrack` entries, each hydrated as `TRACK_V4`; the request and native catalog response explicitly construct only a track page. | No playlist summaries. [Search implementation](../../../third_party/go-librespot/daemon/search.go) and [native response translation](../../../third_party/go-librespot/daemon/native_catalog.go) |
| Recommended | Native collection context (`spotify:user:<user>:collection`) for top tracks/artists | Local recommendation page is derived from collection tracks, then albums; it does not call a Spotify recommendation/feed source. | No personalized or editorial playlist source. [Adapter](../../../internal/spotengine/catalog.go) and [accepted native-catalog ADR](../../../docs/adr/0024-integrate-go-librespot-through-an-in-memory-adapter.md) |
| Artist | `ARTIST_V4` extended metadata | Artist metadata, popular tracks and album groups only. | No playlists containing the artist's tracks. [Pinned implementation](../../../third_party/go-librespot/daemon/native_catalog.go) |
| User Playlists | `GET /user-profile-view/v3/profile/<username>?playlist_limit=…` | Playlist-shaped entries from the listener profile, recursively collected and de-duplicated by URI. | Existing Library source, but it supplies no reliable ownership/follow/public eligibility field. [Pinned implementation](../../../third_party/go-librespot/daemon/native_catalog.go) |

`requestNativeCatalog` enumerates all presently exposed kinds (`liked`, `playlist`, `album`, `artist`, `playlists`, `saved_albums`, `top_tracks`, `top_artists`, and `search`); none represents public discovery, a feed, or artist-playlist relationships. [Pinned request boundary](../../../third_party/go-librespot/daemon/native_catalog.go)

The client is authenticated: every `Spclient.Request` attaches both the client token and the current bearer access token. Thus this is a capability gap in the implemented endpoints and decoders, not an authentication or client-secret gap. [Pinned `Spclient` request implementation](../../../third_party/go-librespot/spclient/spclient.go)

## Native candidates that are *not* implementation-ready

The upstream librespot implementation has a native rootlist request: `/playlist/v2/user/<user>/rootlist?decorate=revision,attributes,length,owner,capabilities,status_code&from=<offset>&length=<length>`. That is useful evidence of a paginated **Library** protocol and is a better future candidate than profile scraping for a complete owned-or-followed User Playlist list. It does not discover Public Playlists, and the pinned Go fork does not currently expose or decode the rootlist response. [Upstream implementation at the inspected revision](https://github.com/librespot-org/librespot/blob/9c7d75615fc093bdcbdb29adbce3fed38c531852/core/src/spclient.rs#L927-L935)

The vendored protocol also names `RECOMMENDED_PLAYLISTS` and defines `PlaylistUriResolverResponse` / `ResolvedPersonalizedPlaylist`. Neither has a corresponding callable endpoint, response decoder, or consumer in the pinned fork. An enum/message name alone does not establish request shape, pagination, or semantics, so neither is a safe source for this ticket. [Extension-kind definition](../../../third_party/go-librespot/proto/spotify/extendedmetadata/extension_kind.proto) and [playlist URI resolver types](../../../third_party/go-librespot/proto/spotify/playlist4/playlist4_external.proto)

## User Playlist eligibility and summary fields

SpotUI defines a **User Playlist** as one in the listener's Library, whether owned or followed, and a **Public Playlist** as discoverable outside that Library. [Ubiquitous language](../../../CONTEXT.md) The current profile collector does not distinguish those states: it accepts an object with `type: playlist`, a playlist URI, or an ID; it neither reads a visibility flag nor distinguishes owner from follower. Therefore it must remain a Library-only source and must not be re-used as a purported public-search or recommendation result.

The existing `PlaylistSummary` needs the following response values:

| SpotUI field | Required native response value |
| --- | --- |
| `URI` | canonical `spotify:playlist:<id>` |
| `Name` | title/name |
| `Owner` | owner display name (and, for eligibility decisions, a stable owner ID/URI) |
| `Description` | description, if supplied |
| `TrackCount` | playlist length/track total |
| `ImageURL` | selected image URL |
| `ExternalURL` | derivable from the canonical URI |

The current translation already consumes this display subset. It lacks a stable owner identifier, visibility, collaboration and “recommended by Spotify” provenance. If a discovery protocol can rank a listener-owned or followed playlist, preserve at least `owner URI/ID` and a source/provenance field alongside the summary; comparing a display name to the listener is not reliable. [Summary type and translator](../../../internal/spotengine/types.go) and [native playlist-shape normalization](../../../third_party/go-librespot/daemon/native_catalog.go)

## Official Spotify API cross-check (not an allowed implementation path)

Spotify's official Web API documents that `/search` accepts `playlist` as a search type and returns playlist pages, so it would technically solve text search. It is excluded here because the ticket and accepted ADR prohibit adding a separate quota-bound Web API client. [Search for Item — Spotify for Developers](https://developer.spotify.com/documentation/web-api/reference/search)

The official featured-playlists and categories-playlists endpoints are marked deprecated; the Recommendations endpoint returns tracks, not playlist recommendations, and is also deprecated. The official artist surface likewise does not provide “playlists containing this artist.” These docs therefore do not supply a viable fallback for the requested Recommended or Artist behavior even if the Web API policy changed. [Featured playlists](https://developer.spotify.com/documentation/web-api/reference/get-featured-playlists), [Recommendations](https://developer.spotify.com/documentation/web-api/reference/get-recommendations), and [artist related artists](https://developer.spotify.com/documentation/web-api/reference/get-an-artists-related-artists)

For comparison only, Spotify documents `/me/playlists` as the list of playlists owned **or followed** by the current user. This agrees with SpotUI's User Playlist language but is not a justification to add Web API access. [Current user playlists](https://developer.spotify.com/documentation/web-api/reference/get-a-list-of-current-users-playlists)

## Implementation implications

1. Keep `nativePlaylist` / context resolution as the shared detail-and-playback path for a discovered `spotify:playlist:` URI; no album-like playback work needs a new protocol.
2. Do not populate Search, Recommended, or Artist playlist sections from the current collection/profile endpoints. Doing so would mislabel library items as public recommendations and cannot establish artist relevance.
3. Before product code, identify and authenticate a new native discovery endpoint for each requested surface (or one endpoint with explicit source metadata), capture a redacted response, and add decoder/pagination contract tests. The response must provide the summary fields above and source-level ranking/provenance.
4. If the rootlist is adopted, use it only to strengthen the existing User Playlist Library implementation; retain it as a separate source from Public Playlist discovery.
