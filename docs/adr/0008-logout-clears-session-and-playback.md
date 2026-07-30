# Logout Clears Session and Playback

## Status

Accepted

## Context

go-librespot persists credentials so a user can resume a session without
logging in on every start. Logout must have different semantics from quitting
SpotUI and must support safe account switching.

Because logout removes reusable credentials and interrupts audio, an accidental
key press would be disruptive.

## Decision

While logged in, `L` will open a `y/N` logout confirmation.

Confirming logout will:

1. stop local playback;
2. close the go-librespot session;
3. remove locally persisted Spotify credentials and session state;
4. clear account-specific data from memory;
5. return to the Logged-Out Screen.

Cancelling will restore the previous screen without changing playback or
session state.

## Consequences

- Quitting SpotUI preserves the Local Session; logout removes it.
- The session store must support explicit deletion and report deletion errors.
- SpotUI must not show logout as successful if credential deletion fails.
- Tests must prove that account data does not survive a successful logout.
