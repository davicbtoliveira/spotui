# Contributing to SpotUI

Thank you for helping improve SpotUI. Keep changes focused, preserve the
standalone client experience, and document user-visible behavior changes.

## Set up a development environment

Install the Go version declared in `go.mod` and the native audio dependencies
listed in the [build instructions](README.md#build-from-source). Then run:

```bash
go mod verify
go test -race ./...
make build
```

Use `make run` for an iterative local run. A Premium account and a connected
audio device are only needed for manual playback validation; the automated
test suite does not use credentials, a browser, network access, or audio.

## Project layout

| Path | Responsibility |
| --- | --- |
| `main.go` | Process entry point and graceful engine shutdown |
| `internal/tui` | Bubble Tea model, browse shell, commands, and views |
| `internal/catalog` | Typed browse routes and catalog command payloads |
| `internal/spotengine` | `Engine` boundary and the go-librespot adapter |
| `internal/browser` | Cross-platform authorization URL opening |
| `internal/artwork` | Terminal artwork decoding and rendering |
| `third_party/go-librespot` | Pinned fork used for authentication, catalog, Connect, and playback |
| `docs/adr` | Decisions that constrain architecture or product behavior |
| `.github/workflows/release.yml` | Test, build, package, and publish release archives |

The runtime flow and package boundaries are described in
[docs/architecture.md](docs/architecture.md).

## Design constraints

These constraints are part of the product, not incidental implementation
details:

- Keep the application standalone: one process, no daemon or sidecar.
- Do not add a Spotify developer application, client credentials, or an
  official Web API dependency for core behavior.
- Keep the TUI dependent on `spotengine.Engine`, not directly on
  `go-librespot`; an architecture test enforces this boundary.
- Preserve Linux and macOS release targets and the system Default Audio Output
  behavior.
- Treat the Local Session as sensitive data and never log or commit it.

If a change intentionally alters one of these constraints, record the decision
in a new numbered ADR under `docs/adr` and update the README.

## Testing and checks

Run the focused test while iterating, then the full suite before submitting a
change:

```bash
go test ./internal/tui/...
go test ./internal/spotengine/...
go test -race ./...
go mod verify
```

For playback, Connect transfer, audio-output, login, and Free-account behavior,
follow the [manual release validation matrix](docs/manual-playback-test.md).
Manual playback evidence is required for a release because automated tests do
not verify audible output on real hardware.

## Documentation checklist

Update the relevant documentation when a change affects:

- supported platforms, installation, native dependencies, or release archives;
- keyboard controls, login, session storage, playback, Connect, or recovery;
- public behavior that should be captured as an architectural decision; or
- the manual release validation procedure.

Do not include real Spotify credentials, authorization URLs, session files, or
private account data in issues, commits, tests, screenshots, or release notes.

## Releases

Pushing a tag matching `v*` runs the release workflow. It tests the project,
builds Linux amd64 and macOS amd64/arm64 archives, includes the license and
third-party notices, and publishes SHA-256 checksums. Run the manual validation
matrix against each release candidate before publishing or announcing it.
