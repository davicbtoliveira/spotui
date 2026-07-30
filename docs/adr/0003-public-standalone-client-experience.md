# Public Standalone Client Experience

## Status

Accepted as a product requirement; enabled through ADR 0004

## Context

SpotUI is intended to behave like a standalone Spotify client. Any Spotify user
should be able to install it, authenticate in the browser, and use it without
creating a developer application or being manually added to an application
allowlist.

Spotify applications start in Development Mode. Development Mode requires each
user to be allowlisted and currently supports at most five authenticated users.
Only Extended Quota Mode removes that restriction. Spotify currently limits new
Extended Quota applications to qualifying organizations.

## Decision

The target user experience will not include:

- Spotify Developer Dashboard access;
- client ID configuration;
- redirect URI configuration;
- custom builds;
- manual user allowlisting.

SpotUI will present itself as one installed client with browser-based login.

## Consequences

- Development Mode cannot satisfy the public release requirement.
- An official public release would require Spotify approval for Extended Quota
  Mode.
- ADR 0004 accepts the unofficial go-librespot path instead of depending on
  Extended Quota Mode.
- Distribution must disclose the unofficial integration and its compatibility
  risk.

## References

- https://developer.spotify.com/documentation/web-api/concepts/quota-modes
- https://developer.spotify.com/policy
