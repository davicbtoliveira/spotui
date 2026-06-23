# Reuse User OAuth Client for Track Search

Track search will use SpotUI's existing Spotify OAuth client instead of adding a separate client-credentials flow. SpotUI is already a user-authorized playback client, and reusing that token keeps search in the same market and playback context while avoiding client secret handling in a distributed terminal app.
