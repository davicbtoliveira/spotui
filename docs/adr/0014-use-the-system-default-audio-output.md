# Use the System Default Audio Output

## Status

Accepted

## Context

Selecting and configuring audio hardware would add platform-specific UI and
failure modes to the first go-librespot increment. The target installation flow
must work without configuration.

Linux and macOS already expose a user-selected default output through their
audio systems.

## Decision

The first increment will open the operating system's default audio output:

- Linux through the available go-librespot ALSA or PulseAudio backend;
- macOS through go-librespot AudioToolbox.

On Linux, SpotUI selects go-librespot PulseAudio when `PULSE_SERVER` is set or
the WSLg PulseAudio socket is available. It otherwise selects go-librespot
ALSA. Spotify search, playback, and audio output remain inside go-librespot.

SpotUI will not include an audio-device selector or config file requirement in
the first increment.

An output initialization failure will remain inside the TUI and show an
actionable error rather than crash or silently discard audio.

## Consequences

- Normal OS audio settings determine speakers, headphones, or other output.
- Changing the OS default while SpotUI runs may require restarting playback or
  SpotUI, depending on the backend.
- Explicit device selection and backend overrides are deferred.
- Linux release validation must cover the supported default backend paths.
