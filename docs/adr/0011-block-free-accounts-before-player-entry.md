# Block Free Accounts Before Player Entry

## Status

Accepted

## Context

go-librespot requires Spotify Premium for music playback and does not implement
the advertising and limited-control behavior required by Spotify Free.

Letting a Free account enter the normal player would expose controls that cannot
work and produce failures only after user interaction.

## Decision

SpotUI will allow browser authentication to complete, then inspect the account
product level before entering the player.

A Spotify Free account will see an unsupported-account screen stating that
Spotify Premium is required. The screen will offer only:

- Logout, which removes the Local Session and returns to the Logged-Out Screen;
- Quit, which preserves the Local Session and exits.

The Free account will not enter Track Search or playback views.

## Consequences

- Premium capability becomes an explicit post-authentication gate.
- SpotUI must distinguish authentication failure from unsupported account type.
- Upgrading to Premium can be recognized by retrying session initialization or
  logging in again.
- No partial browse-only product is included in the first increment.
