# Remain Local After Playback Transfer

## Status

Accepted

## Context

As a Spotify Connect device, SpotUI can become inactive when the user transfers
playback to another device. go-librespot then stops local playback.

SpotUI's product scope is a standalone local player, not a general remote
controller for whichever device is active.

## Decision

When playback transfers away from the SpotUI Connect Device, SpotUI will:

- stop local audio;
- retain the authenticated session;
- show that playback was transferred;
- keep Track Search and navigation available;
- avoid sending controls to the new remote device.

Selecting a Search Result with `Enter` will reactivate SpotUI and start local
playback according to ADR 0021.

## Consequences

- Player controls are scoped to SpotUI-owned playback.
- The inactive-device state is distinct from pause and stop.
- External transfer events must update the TUI immediately.
- SpotUI never requires the other device to close before reclaiming playback.
