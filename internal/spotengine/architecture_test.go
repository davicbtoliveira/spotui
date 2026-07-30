package spotengine_test

import (
	"go/parser"
	"go/token"
	"io/fs"
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
