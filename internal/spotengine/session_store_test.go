package spotengine

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	librespot "github.com/devgianlu/go-librespot"
	"go.uber.org/goleak"
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

func TestCorruptLocalSessionRecoversToLoggedOutAdapter(t *testing.T) {
	configDir := t.TempDir()
	path := filepath.Join(configDir, "session.json")
	if err := os.WriteFile(path, []byte(`{"credentials":`), 0o600); err != nil {
		t.Fatalf("write corrupt session: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, ".session-stale.tmp"), []byte("secret"), 0o600); err != nil {
		t.Fatalf("write stale replacement: %v", err)
	}

	adapter, err := newAdapterAtDir(configDir)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}
	if adapter.HasSession() {
		t.Fatal("corrupt session was treated as reusable")
	}

	entries, err := os.ReadDir(configDir)
	if err != nil {
		t.Fatalf("read config directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("temporary session leaked after recovery: %s", entry.Name())
		}
	}
}

func TestAdapterLogoutRemovesSessionArtifactsAndKeepsPreferences(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	configDir := t.TempDir()
	sessionPath := filepath.Join(configDir, "session.json")
	state := &librespot.AppState{DeviceId: "device-id"}
	state.Credentials.Username = "listener"
	state.Credentials.Data = []byte("credential")
	if err := newFileStateStore(sessionPath).Save(state); err != nil {
		t.Fatalf("save session: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, ".session-stale.tmp"), []byte("secret"), 0o600); err != nil {
		t.Fatalf("write stale session replacement: %v", err)
	}
	preferences := newPreferenceStore(filepath.Join(configDir, "settings.json"))
	if err := preferences.SaveAutoplay(false); err != nil {
		t.Fatalf("save preferences: %v", err)
	}

	adapter, err := newAdapterAtDir(configDir)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := adapter.Logout(ctx); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if adapter.HasSession() {
		t.Fatal("logout retained Local Session")
	}
	for _, path := range []string{sessionPath, filepath.Join(configDir, ".session-stale.tmp")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("session artifact remains at %s: %v", path, err)
		}
	}
	autoplay, err := preferences.LoadAutoplay()
	if err != nil {
		t.Fatalf("load preferences: %v", err)
	}
	if autoplay {
		t.Fatal("logout removed Autoplay preference")
	}
	if err := adapter.Close(ctx); err != nil {
		t.Fatalf("close after logout: %v", err)
	}
}

func TestAdapterClosePreservesReusableLocalSession(t *testing.T) {
	configDir := t.TempDir()
	sessionPath := filepath.Join(configDir, "session.json")
	state := &librespot.AppState{DeviceId: "device-id"}
	state.Credentials.Username = "listener"
	state.Credentials.Data = []byte("credential")
	if err := newFileStateStore(sessionPath).Save(state); err != nil {
		t.Fatalf("save session: %v", err)
	}

	adapter, err := newAdapterAtDir(configDir)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}
	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	loaded, err := newFileStateStore(sessionPath).Load()
	if err != nil {
		t.Fatalf("load preserved session: %v", err)
	}
	if loaded == nil || len(loaded.Credentials.Data) == 0 {
		t.Fatal("close removed reusable Local Session")
	}
}

func TestAdapterDetectsSessionPersistedAfterConstruction(t *testing.T) {
	configDir := t.TempDir()
	adapter, err := newAdapterAtDir(configDir)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}
	if adapter.HasSession() {
		t.Fatal("new adapter unexpectedly has session")
	}

	state := &librespot.AppState{DeviceId: "device-id"}
	state.Credentials.Username = "listener"
	state.Credentials.Data = []byte("credential")
	if err := newFileStateStore(filepath.Join(configDir, "session.json")).Save(state); err != nil {
		t.Fatalf("persist session: %v", err)
	}
	if !adapter.HasSession() {
		t.Fatal("adapter did not observe newly persisted session")
	}
	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
}
