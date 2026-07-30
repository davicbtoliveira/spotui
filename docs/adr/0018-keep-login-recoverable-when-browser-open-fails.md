# Keep Login Recoverable When Browser Open Fails

## Status

Accepted

## Context

Opening a browser through `xdg-open` on Linux or `open` on macOS can fail even
in a Local Terminal. Authentication can still continue if the user opens the
authorization URL manually.

An external process failure must not terminate SpotUI or discard a valid pending
OAuth callback.

## Decision

While authentication is pending, SpotUI will display the authorization URL in a
copyable form.

If automatic browser opening fails:

- `r` retries opening the same URL;
- `esc` cancels Login, closes the callback listener, and returns to the
  Logged-Out Screen;
- manually opening the displayed URL remains valid while the screen is active.

## Consequences

- Browser launcher failure is recoverable inside the TUI.
- The pending authorization URL and callback server share one cancellable
  lifecycle.
- Retry must not start a second OAuth transaction or callback listener.
- URLs must not be written to persistent logs at normal log levels.
