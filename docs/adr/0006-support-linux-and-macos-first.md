# Support Linux and macOS First

## Status

Accepted

## Context

SpotUI currently advertises Linux, macOS, and Windows release binaries.
go-librespot provides local output through ALSA and PulseAudio on Linux and
AudioToolbox on macOS. It does not currently provide a native Windows audio
backend.

Adding Windows playback would require a new output implementation such as
WASAPI and additional release validation.

## Decision

The first go-librespot-based SpotUI release will support:

- Linux;
- macOS on Intel;
- macOS on Apple Silicon.

Windows support is deferred.

## Consequences

- Linux and macOS must both receive playback integration and release tests.
- Existing Windows release documentation and automation must clearly mark
  Windows as unsupported for this release line.
- Windows support requires a separate audio-backend decision and implementation.
