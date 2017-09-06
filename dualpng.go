package dualpng

import (
	"image"
	"image/color"
	"image/draw"
	"io"

	"github.com/Necroforger/dualpng/gamapng"
)

// CreateMask creates a mask from the given matrix repeated over
// bounds b.
//    m : Mask matrix. Values greater than 0 will be opaque.
//           Values lower than zero will be transparent.
//    b : Bounds of the mask.
func CreateMask(m [][]int, b image.Rectangle) image.Image {
	mask := image.NewRGBA(b)
	var clr color.RGBA

	for y := b.Min.Y; y < b.Max.Y; y += len(m[0]) {
		for x := b.Min.X; x < b.Max.X; x += len(m) {

			// Draw mask
			for i := 0; i < len(m) && i+y < b.Max.Y; i++ {
				for j := 0; j < len(m[0]) && x+j < b.Max.X; j++ {
					if m[i][j] > 0 {
						clr = color.RGBA{0, 0, 0, 255}
					} else {
						clr = color.RGBA{0, 0, 0, 0}
					}
					mask.Set(x+j, y+i, clr)
				}
			}

		}
	}
	return mask
}

// MergeImages merges two images with the given maskmatrix.
// If maskmatrix is nil, it will merge by alternating every other pixel
//     img1       : first image
//     img2       : Second image
//     maskmatrix : Mask to use when merging two images together.
func MergeImages(img1, img2 image.Image, maskmatrix [][]int) image.Image {
	var maxWidth, maxHeight int
	b := img1.Bounds()
	b2 := img2.Bounds()
	if b.Dx() > b2.Dx() {
		maxWidth = b.Dx()
	} else {
		maxWidth = b2.Dx()
	}
	if b.Dy() > b2.Dy() {
		maxHeight = b.Dy()
	} else {
		maxHeight = b2.Dy()
	}

	combined := image.NewRGBA(image.Rect(0, 0, maxWidth, maxHeight))

	if maskmatrix == nil {
		b = combined.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if y%2 == 0 || x%2 == 0 {
					combined.Set(x, y, img1.At(x, y))
				} else {
					combined.Set(x, y, img2.At(x, y))
				}
			}
		}
	} else {
		mask := CreateMask(maskmatrix, combined.Bounds())
		draw.Draw(combined, img2.Bounds(), img2, image.ZP, draw.Src)
		draw.DrawMask(combined, mask.Bounds(), img1, image.ZP, mask, image.ZP, draw.Over)
	}

	return combined
}

// LevelImage sets the RGB values of the given image to be within the specified range.
//     img  : Source image
//     low  : Lowest RGB value in range
//     high : Highest RGB value in range
func LevelImage(img image.Image, low uint8, high uint8) image.Image {
	out := image.NewRGBA(img.Bounds())
	b := img.Bounds()
	level := func(n uint8, low, high uint8) uint8 {
		return uint8((float64(n)/255.0)*float64(high-low)) + low
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			out.Set(x, y, color.RGBA{
				level(uint8(r>>8), low, high),
				level(uint8(g>>8), low, high),
				level(uint8(b>>8), low, high),
				uint8(a >> 8),
			})
		}
	}
	return out
}

// Encode encodes the image with as png with a gAMA chunk
// with value gAMA. gAMA values are multiplied by 100,000
// so if you want to use a gAMA value of 0.023, you would enter
// 2,300 for gAMA.
//    w    : destination writer.
//    img  : image to encode
//    gAMA : gAMA value to give the image.
func Encode(w io.Writer, img image.Image, gAMA uint32) error {
	return gamapng.Encode(w, img, gAMA)
}
