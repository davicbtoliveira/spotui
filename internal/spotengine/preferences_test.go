package spotengine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAutoplayPreferenceDefaultsOnAndPersistsOutsideSession(t *testing.T) {
	configDir := t.TempDir()
	store := newPreferenceStore(filepath.Join(configDir, "settings.json"))

	enabled, err := store.LoadAutoplay()
	if err != nil {
		t.Fatalf("load default: %v", err)
	}
	if !enabled {
		t.Fatal("autoplay default: want enabled")
	}

	if err := store.SaveAutoplay(false); err != nil {
		t.Fatalf("save autoplay: %v", err)
	}
	enabled, err = store.LoadAutoplay()
	if err != nil {
		t.Fatalf("reload autoplay: %v", err)
	}
	if enabled {
		t.Fatal("autoplay preference was not persisted")
	}
	if _, err := os.Stat(filepath.Join(configDir, "session.json")); !os.IsNotExist(err) {
		t.Fatalf("autoplay wrote session credentials: %v", err)
	}

	if err := newFileStateStore(filepath.Join(configDir, "session.json")).Clear(); err != nil {
		t.Fatalf("clear local session: %v", err)
	}
	enabled, err = store.LoadAutoplay()
	if err != nil {
		t.Fatalf("load after session clear: %v", err)
	}
	if enabled {
		t.Fatal("clearing local session removed autoplay preference")
	}
}
