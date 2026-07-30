# Reconnect Without Reauthentication

## Status

Accepted

## Context

Network loss and temporary Spotify service failures can terminate or stall the
go-librespot runtime. A valid reusable credential should not require another
browser Login after a transient failure.

Repeated immediate reconnect attempts would create a busy loop and unnecessary
load.

## Decision

SpotUI will enter a visible reconnecting state after a transient engine or
network failure.

It will recreate the go-librespot runtime with the existing Local Session using
bounded exponential backoff with jitter. The delay will start near one second
and cap at thirty seconds.

Only an explicit credential rejection will end reconnection. SpotUI will then
remove the unusable credential and return to the Logged-Out Screen with a
session-expired message.

## Consequences

- Temporary outages do not open the browser.
- Search and playback commands are disabled while reconnecting, but Quit remains
  available.
- A successful reconnect returns to the ready state without automatically
  resuming audio.
- Tests need deterministic backoff injection rather than real timers.
