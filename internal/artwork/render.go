package artwork

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi/kitty"
)

type CellRect struct{ Columns, Rows int }

func ResizeNearest(src image.Image, width, height int) *image.RGBA {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	bounds := src.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return dst
	}
	for y := 0; y < height; y++ {
		sy := bounds.Min.Y + y*bounds.Dy()/height
		for x := 0; x < width; x++ {
			sx := bounds.Min.X + x*bounds.Dx()/width
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

// ResizeContain keeps the source aspect ratio inside the requested cell grid.
func ResizeContain(src image.Image, width, height int) *image.RGBA {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	if src == nil || src.Bounds().Dx() == 0 || src.Bounds().Dy() == 0 {
		return dst
	}
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	scale := float64(width) / float64(sw)
	if v := float64(height) / float64(sh); v < scale {
		scale = v
	}
	dw, dh := maxInt(1, int(float64(sw)*scale)), maxInt(1, int(float64(sh)*scale))
	resized := ResizeNearest(src, dw, dh)
	ox, oy := (width-dw)/2, (height-dh)/2
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			dst.Set(ox+x, oy+y, resized.At(x, y))
		}
	}
	return dst
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func RenderANSI(src image.Image, rect CellRect) string {
	if src == nil || rect.Columns < 1 || rect.Rows < 1 {
		return ""
	}
	pixels := ResizeContain(src, rect.Columns, rect.Rows*2)
	var out strings.Builder
	for y := 0; y < rect.Rows*2; y += 2 {
		for x := 0; x < rect.Columns; x++ {
			top := color.RGBAModel.Convert(pixels.At(x, y)).(color.RGBA)
			bottom := top
			if y+1 < rect.Rows*2 {
				bottom = color.RGBAModel.Convert(pixels.At(x, y+1)).(color.RGBA)
			}
			fmt.Fprintf(&out, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", top.R, top.G, top.B, bottom.R, bottom.G, bottom.B)
		}
		out.WriteString("\x1b[0m")
		if y+2 < rect.Rows*2 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func EncodeKitty(src image.Image, rect CellRect, imageID, placementID int) ([]byte, error) {
	if src == nil {
		return nil, fmt.Errorf("artwork image is nil")
	}
	if imageID < 1 || placementID < 1 {
		return nil, fmt.Errorf("artwork ids must be positive")
	}
	var buffer bytes.Buffer
	err := kitty.EncodeGraphics(&buffer, src, &kitty.Options{
		Action:          kitty.TransmitAndPut,
		Format:          kitty.PNG,
		Transmission:    kitty.Direct,
		ID:              imageID,
		PlacementID:     placementID,
		Columns:         rect.Columns,
		Rows:            rect.Rows,
		Chunk:           true,
		DoNotMoveCursor: true,
	})
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DeleteKitty(imageID, placementID int) []byte {
	if imageID < 1 || placementID < 1 {
		return nil
	}
	var buffer bytes.Buffer
	_, _ = fmt.Fprintf(&buffer, "\x1b_Ga=d,d=I,i=%d,p=%d\x1b\\", imageID, placementID)
	return buffer.Bytes()
}
