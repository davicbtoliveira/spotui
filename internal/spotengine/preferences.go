package spotengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type preferenceStore struct {
	path string
}

type preferences struct {
	Autoplay *bool `json:"autoplay"`
}

func newPreferenceStore(path string) *preferenceStore {
	return &preferenceStore{path: path}
}

func (s *preferenceStore) LoadAutoplay() (bool, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read preferences: %w", err)
	}

	var prefs preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return false, fmt.Errorf("decode preferences: %w", err)
	}
	if prefs.Autoplay == nil {
		return true, nil
	}
	return *prefs.Autoplay, nil
}

func (s *preferenceStore) SaveAutoplay(enabled bool) error {
	data, err := json.Marshal(preferences{Autoplay: &enabled})
	if err != nil {
		return fmt.Errorf("encode preferences: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create preferences directory: %w", err)
	}
	file, err := os.CreateTemp(dir, ".settings-*.tmp")
	if err != nil {
		return fmt.Errorf("create preferences replacement: %w", err)
	}
	tempPath := file.Name()
	defer func() { _ = os.Remove(tempPath) }()

	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return fmt.Errorf("secure preferences replacement: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write preferences replacement: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close preferences replacement: %w", err)
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace preferences: %w", err)
	}
	return nil
}
