package spotengine_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTUIDoesNotImportGoLibrespot(t *testing.T) {
	err := filepath.WalkDir("../tui", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, "github.com/devgianlu/go-librespot") {
				t.Errorf("%s imports go-librespot directly: %s", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect TUI imports: %v", err)
	}
}

func TestStandaloneClientHasNoOfficialSpotifyDependency(t *testing.T) {
	for _, dependency := range []string{"github.com/zmb3/spotify", "golang.org/x/oauth2"} {
		err := filepath.WalkDir("..", func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".go" {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imported := range file.Imports {
				importPath, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					return err
				}
				if strings.HasPrefix(importPath, dependency) {
					t.Errorf("%s imports removed dependency %s", path, importPath)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("inspect imports: %v", err)
		}
	}

	goMod, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, removed := range []string{"github.com/zmb3/spotify"} {
		if strings.Contains(string(goMod), removed) {
			t.Errorf("go.mod retains removed dependency %s", removed)
		}
	}
}

func TestReleaseWorkflowBuildsSupportedStandaloneArtifacts(t *testing.T) {
	workflow, err := os.ReadFile("../../.github/workflows/release.yml")
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}
	content := string(workflow)

	for _, required := range []string{
		"linux-amd64",
		"darwin-amd64",
		"darwin-arm64",
		"go-version-file: go.mod",
		"LICENSE",
		"THIRD_PARTY_NOTICES.md",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("release workflow missing %q", required)
		}
	}
	for _, forbidden := range []string{"windows", "SPOTIFY_CLIENT_ID"} {
		if strings.Contains(strings.ToLower(content), strings.ToLower(forbidden)) {
			t.Errorf("release workflow retains unsupported configuration %q", forbidden)
		}
	}
}

func TestReadmeDocumentsStandaloneClientConstraints(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	content := strings.ToLower(string(readme))

	for _, expected := range []string{
		"zero configuration",
		"spotify premium",
		"linux",
		"macos",
		"local terminal",
		"unofficial protocol",
		"not affiliated",
		"default audio output",
	} {
		if !strings.Contains(content, expected) {
			t.Errorf("README missing %q", expected)
		}
	}
	for _, stale := range []string{"spotify_client_id", "playlists tab", "toggle shuffle"} {
		if strings.Contains(content, stale) {
			t.Errorf("README retains unsupported behavior %q", stale)
		}
	}
}
