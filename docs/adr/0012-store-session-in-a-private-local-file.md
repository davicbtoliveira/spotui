# Store Session in a Private Local File

## Status

Accepted

## Context

go-librespot needs reusable account credentials to restore a Local Session
without browser login on every start. OS credential stores are not consistently
available in headless Linux environments and would add platform-specific
integration to the first increment.

OAuth and reusable Spotify credentials are sensitive and must not inherit
permissive default file modes.

## Decision

The first increment will persist one active account in:

`~/.config/spotui/session.json`

or the operating system's equivalent user configuration directory.

The containing SpotUI directory will use mode `0700` and the session file mode
`0600` on Unix systems. Writes will be atomic. Logs and errors will never
contain credential values.

OS keychain and libsecret integration are deferred.

## Consequences

- Headless Linux and macOS use the same persistence model.
- Other local users cannot read the credential under normal filesystem
  permission enforcement.
- Anyone with access to the user's account or configuration directory can still
  extract the credential.
- Logout must delete the session file and any temporary replacement file.
- Corrupt or unreadable session files must produce a recoverable login path.
