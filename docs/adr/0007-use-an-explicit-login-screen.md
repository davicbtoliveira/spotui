# Use an Explicit Login Screen

## Status

Accepted

## Context

SpotUI currently starts authentication immediately and opens the user's browser
during application initialization. This is surprising, leaves no useful
logged-out state, and makes retry and logout flows awkward.

The desired first-run experience gives the user an option to log in.

## Decision

When no valid Local Session exists, SpotUI will show a logged-out screen instead
of opening the browser automatically.

The screen will:

- identify SpotUI;
- state that Spotify Premium is required for music playback;
- offer `Enter` to log in with Spotify;
- offer the normal quit action.

Pressing `Enter` will start go-librespot interactive OAuth, open the browser,
show pending progress in the TUI, and transition to library loading after the
loopback callback succeeds.

Authentication errors will return to the logged-out screen with a retryable
message instead of quitting SpotUI.

## Consequences

- Authentication becomes an explicit user action.
- The application state model needs logged-out, authenticating, loading, ready,
  and authentication-error behavior.
- A failed or cancelled browser flow remains recoverable inside the TUI.
