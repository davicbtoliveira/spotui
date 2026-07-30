# 10 — Explore Artist details

**What to build:** Turn an Artist Search Result into a useful exploration route
with honest metadata, playable popular Tracks, and principal Albums that reuse
the normal Album experience.

**Blocked by:** 09 — Search Tracks, Albums, Artists, and Playlists.

**Status:** ready-for-agent

- [ ] Artist details show image, name, Spotify URL, optional genres, popular
      Tracks when supplied, and paginated principal Albums.
- [ ] Follower count is not requested or displayed because no supported source
      provides it; missing genres are omitted rather than rendered as empty or
      unknown facts.
- [ ] Popular Tracks use the private Artist metadata seam behind the fork and
      provide an explicit empty/unavailable state when that metadata cannot be
      decoded or served.
- [ ] Principal Albums use the supported authenticated Catalog read and
      deduplicate equivalent releases by stable identity where possible.
- [ ] `Enter` on a popular Track starts that Track immediately without creating
      an invented Artist context.
- [ ] `Enter` on an Album opens the established Album detail and contextual
      playback route.
- [ ] Back navigation restores the Artist route and then the originating Search
      group, including selections and scroll positions at both levels.
- [ ] Artist metadata, popular Tracks, and Album pages load independently so one
      unavailable subsection does not fail the whole view.
- [ ] Loaded Artist data is cached for the account session and cleared on
      Logout or credential expiry.
- [ ] Root-model tests cover optional metadata, unavailable popular Tracks,
      independent errors, Track activation, nested Album drill-down, and
      multi-level state restoration.
- [ ] Adapter and fork contract tests cover Artist translation, private popular
      Track metadata, and Album paging without live Spotify access.
