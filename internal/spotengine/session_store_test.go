package spotengine

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	librespot "github.com/devgianlu/go-librespot"
)

func TestFileStateStoreMakesCredentialsReusable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spotui", "session.json")
	store := newFileStateStore(path)
	state := &librespot.AppState{DeviceId: "device-id"}
	state.Credentials.Username = "listener"
	state.Credentials.Data = []byte("reusable-credential")

	if err := store.Save(state); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded == nil {
		t.Fatal("loaded session is nil")
	}
	if loaded.DeviceId != state.DeviceId ||
		loaded.Credentials.Username != state.Credentials.Username ||
		string(loaded.Credentials.Data) != string(state.Credentials.Data) {
		t.Fatalf("loaded session does not match saved session: %#v", loaded)
	}
}

func TestFileStateStoreUsesPrivateModesAndLeavesNoReplacement(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file modes are not enforced on Windows")
	}

	dir := filepath.Join(t.TempDir(), "spotui")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("create permissive directory: %v", err)
	}
	path := filepath.Join(dir, "session.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatalf("create permissive session: %v", err)
	}

	state := &librespot.AppState{DeviceId: "device-id"}
	state.Credentials.Data = []byte("credential")
	if err := newFileStateStore(path).Save(state); err != nil {
		t.Fatalf("save: %v", err)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat directory: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("directory mode: want 0700, got %04o", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat session: %v", err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("session mode: want 0600, got %04o", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read directory: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "session.json" {
		t.Fatalf("unexpected local session files: %v", entries)
	}
}

func TestAdapterDetectsReusableLocalSession(t *testing.T) {
	configDir := t.TempDir()
	path := filepath.Join(configDir, "session.json")
	state := &librespot.AppState{DeviceId: "device-id"}
	state.Credentials.Username = "listener"
	state.Credentials.Data = []byte("credential")
	if err := newFileStateStore(path).Save(state); err != nil {
		t.Fatalf("save: %v", err)
	}

	adapter, err := newAdapterAtDir(configDir)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}
	if !adapter.HasSession() {
		t.Fatal("adapter did not detect reusable Local Session")
	}
}
