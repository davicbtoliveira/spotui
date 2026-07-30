# Premium Local Playback Test

Use this check before releasing changes to the Playback Engine. It requires a
Spotify Premium account and speakers or headphones connected to the system
Default Audio Output.

## Procedure

1. Stop every official Spotify player.
2. Set the operating-system default output to the device being tested.
3. Run `go run .` from a local terminal.
4. Log in and confirm that the Track Search screen appears.
5. Search for a known track and verify name, artist, duration, and URI-backed
   selection remain stable while moving the cursor and changing pages.
6. Press `Enter` on the Search Result.
7. Confirm audio starts from the beginning through the default output.
8. Confirm metadata, buffering, playing state, and progress change in the TUI.
9. Disconnect or disable the default output and select another result.
10. Confirm the audio failure appears in the TUI and the terminal remains
    responsive.
11. Quit and confirm local audio stops.

## Evidence

Record platform, SpotUI commit, selected track URI, output device, result, and
any error text in the release notes. Never record credentials or session-file
contents.
