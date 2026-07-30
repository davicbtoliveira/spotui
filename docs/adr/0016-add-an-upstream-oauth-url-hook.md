# Add an Upstream OAuth URL Hook

## Status

Accepted

## Context

go-librespot's interactive authentication creates a loopback callback server and
constructs the Spotify authorization URL, but exposes that URL only through a
log message. SpotUI must open the browser after an explicit Login action and
must not depend on parsing human-readable dependency logs.

Reimplementing the complete authentication flow in SpotUI would duplicate
go-librespot internals.

## Decision

We will propose an upstream go-librespot API that exposes the authorization URL
through an explicit callback such as:

`OnAuthURL(url string)`

SpotUI will use a minimal temporary fork containing that change until it is
released upstream. The fork will not contain unrelated behavior changes.

## Consequences

- SpotUI can open the browser and render the fallback URL without parsing logs.
- Release builds must pin an exact fork revision until an upstream release
  includes the hook.
- The change needs focused upstream tests and documentation.
- If upstream rejects the API, we must revisit ownership of the OAuth flow.
- The fork will be removed after migration to an upstream tagged release.
