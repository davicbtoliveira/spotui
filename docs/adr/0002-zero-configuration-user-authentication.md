# Zero-Configuration User Authentication

## Status

Superseded by ADR 0004

## Context

SpotUI currently requires a Spotify client ID before the TUI starts. A user
installing SpotUI must not create a Spotify developer application, register a
redirect URI, set an environment variable, or build a custom binary.

Spotify's Authorization Code with PKCE flow is suitable for installed
applications because it does not require a client secret.

## Decision

SpotUI releases will include the SpotUI Spotify application's public client ID.
The maintainer will register SpotUI's loopback redirect URI in the Spotify
Developer Dashboard.

The user authentication flow will be:

1. Install and start SpotUI.
2. Choose login when no local session exists.
3. Authorize SpotUI in the browser.
4. Return automatically to SpotUI through the loopback callback.
5. Reuse the locally stored session on later starts.

Development builds may override the bundled client ID, but this is not part of
the end-user flow.

## Consequences

- End users do not need a Spotify developer account or custom build.
- No client secret is distributed.
- SpotUI owns the Spotify application registration and redirect URI.
- Development Mode allowlists and quotas apply until Spotify grants SpotUI
  Extended Quota Mode.
- Logout must remove the local session and return to the logged-out state.
