package spotengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	librespot "github.com/devgianlu/go-librespot"
)

type fileStateStore struct {
	path string
}

func newFileStateStore(path string) *fileStateStore {
	return &fileStateStore{path: path}
}

func (s *fileStateStore) Load() (*librespot.AppState, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read local session: %w", err)
	}

	var state librespot.AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode local session: %w", err)
	}
	return &state, nil
}

func (s *fileStateStore) Save(state *librespot.AppState) error {
	state.Lock()
	data, err := json.Marshal(state)
	state.Unlock()
	if err != nil {
		return fmt.Errorf("encode local session: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create local session directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("secure local session directory: %w", err)
	}

	file, err := os.CreateTemp(dir, ".session-*.tmp")
	if err != nil {
		return fmt.Errorf("create local session replacement: %w", err)
	}
	tempPath := file.Name()
	defer func() { _ = os.Remove(tempPath) }()

	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("secure local session replacement: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write local session replacement: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync local session replacement: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close local session replacement: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace local session: %w", err)
	}
	if err := os.Chmod(s.path, 0o600); err != nil {
		return fmt.Errorf("secure local session: %w", err)
	}
	return nil
}
