# Play Search Selection Immediately

## Status

Accepted

## Context

A Search Result can be selected while another track, context, or externally
supplied queue is active. The first increment does not include explicit queue
management.

## Decision

Pressing `Enter` on a Search Result will:

1. replace the current playback context and queue;
2. activate the SpotUI Connect Device;
3. load the selected track;
4. start local playback immediately.

The selected track starts at position zero. Autoplay may supply recommendations
after the resulting context ends when enabled.

## Consequences

- Search selection is deterministic and does not append silently.
- Existing queued items are discarded.
- Explicit add-to-queue behavior is deferred.
- Playback loading and errors must be visible without freezing the TUI.
