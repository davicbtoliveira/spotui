package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/tui/commands"
)

const volumeDebounceDelay = 35 * time.Millisecond

func (m *RootModel) handlePlaybackKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case KeySpace:
		if m.engineTransferred {
			return nil, true
		}
		if m.enginePlaying {
			return commands.CmdPauseEngine(m.engine), true
		}
		return commands.CmdResumeEngine(m.engine), true
	case KeyNext:
		if m.engineTransferred {
			return nil, true
		}
		return commands.CmdNextEngine(m.engine), true
	case KeyPrev:
		if m.engineTransferred {
			return nil, true
		}
		return commands.CmdPreviousEngine(m.engine), true
	case KeyVolumeDown:
		if m.engineTransferred {
			return nil, true
		}
		return m.adjustVolume(-5), true
	case KeyVolumeUp:
		if m.engineTransferred {
			return nil, true
		}
		return m.adjustVolume(5), true
	case KeyAutoplay:
		if m.engineTrack == nil || m.engineTransferred {
			return nil, true
		}
		return commands.CmdSetEngineAutoplay(m.engine, !m.engineAutoplay), true
	case "s":
		if m.browseContextURI == "" || m.engineTransferred {
			return nil, true
		}
		return commands.CmdSetEngineShuffle(m.engine, !m.engineShuffle), true
	case "h":
		if m.engineTransferred || m.localProgressMs <= 0 {
			return nil, true
		}
		return commands.CmdSeekEngine(m.engine, -10000), true
	case "l":
		if m.engineTransferred {
			return nil, true
		}
		if m.engineTrack != nil && m.engineTrack.DurationMS > 0 && m.localProgressMs >= m.engineTrack.DurationMS {
			return nil, true
		}
		return commands.CmdSeekEngine(m.engine, 10000), true
	default:
		return nil, false
	}
}

func (m *RootModel) adjustVolume(delta int) tea.Cmd {
	volume := m.engineVolume + delta
	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}
	if volume == m.engineVolume {
		return nil
	}
	m.engineVolume = volume
	m.volumeDebouncePending = true
	m.volumeDebounceID++
	return commands.CmdDebounceVolume(m.volumeDebounceID, volumeDebounceDelay)
}
