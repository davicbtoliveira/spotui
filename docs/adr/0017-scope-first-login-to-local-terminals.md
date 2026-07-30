# Scope First Login to Local Terminals

## Status

Accepted

## Context

Interactive go-librespot OAuth returns through a loopback HTTP listener on the
machine running SpotUI. In a local desktop terminal, the browser and SpotUI
share that loopback interface.

In an SSH or otherwise headless session, the browser often runs on another
machine. The redirect cannot reach SpotUI without port forwarding or a manual
callback relay, which conflicts with the zero-configuration login goal.

## Decision

The first increment supports interactive Login only when SpotUI and the browser
run on the same machine.

SSH, remote containers, and headless authentication are deferred and documented
as unsupported for the first release.

## Consequences

- The loopback callback can use a dynamically allocated local port.
- No SSH tunnel or callback-copy instructions are part of the MVP.
- Restoring an existing Local Session can still work without launching a
  browser.
- Remote authentication requires a later design.
