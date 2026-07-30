package artwork

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestRenderANSIRendersTwoPixelsPerCell(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	got := RenderANSI(img, CellRect{Columns: 1, Rows: 1})
	if !strings.Contains(got, "38;2;255;0;0") || !strings.Contains(got, "48;2;0;0;255") || !strings.Contains(got, "▀") {
		t.Fatalf("ANSI artwork: %q", got)
	}
}

func TestEncodeKittyUsesOwnedIDs(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	data, err := EncodeKitty(img, CellRect{Columns: 2, Rows: 2}, 7, 9)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "i=7") || !strings.Contains(text, "p=9") || !strings.Contains(text, "C=1") {
		t.Fatalf("kitty sequence: %q", text)
	}
}

func TestDeleteKittyTargetsOnlyOwnedPlacement(t *testing.T) {
	if got := string(DeleteKitty(7, 9)); got != "\x1b_Ga=d,d=I,i=7,p=9\x1b\\" {
		t.Fatalf("delete: %q", got)
	}
}
