package main

/*
EXAMPLE USAGE:
	`go run dualpng.go -g 1000 -w 1024 -m "[[1,1],[1,0]]" ika_musume.jpeg hakase_trumpet.png`
	`-w 1024` Resizes both images to have a width of 1000
	`-m "[[1,1],[1,0]]"` Creates a checkerboard mask. If you were to leave it empty,
	it would default to the checkerboard pattern.
*/

import (
	"encoding/json"
	"errors"
	"flag"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"

	dp "github.com/Necroforger/dualpng"
)

// Flags
var (
	Width      = flag.Uint("w", 0, "Width to resize both images to")
	Height     = flag.Uint("h", 0, "Height to resize both images to")
	Range1     = flag.String("r1", "0-230", "RGB Colour range for the first image")
	Range2     = flag.String("r2", "230-255", "RGB Colour range for the second image")
	Gama       = flag.Uint("g", 2300, "gAMA value")
	OutputPath = flag.String("o", "", "Output file name")
	MaskMatrix = flag.String("m", "", "Mask matrix to use for masking images. Ex [[1, 1],[1,0]] will create a checkerboard pattern")
)

func handle(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func parseRange(txt string) (from int, to int, err error) {
	numbers := strings.Split(txt, "-")
	switch len(numbers) {
	case 1:
		to, err = strconv.Atoi(numbers[0])
	case 2:
		to, err = strconv.Atoi(numbers[0])
		from, err = strconv.Atoi(numbers[1])
	default:
		err = errors.New("Invalid range")
		return
	}
	return
}

func getImage(path string) (image.Image, error) {
	var (
		source io.Reader
		err    error
	)

	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		source = resp.Body
		defer resp.Body.Close()
	} else {
		source, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}

	img, _, err := image.Decode(source)
	return img, err
}

func createUniformImage(clr color.Color, b image.Rectangle) *image.RGBA {
	out := image.NewRGBA(b)
	draw.Draw(out, b, image.NewUniform(clr), image.ZP, draw.Src)
	return out
}

func main() {
	var (
		img1, img2 image.Image
		err        error
		out        *os.File
		mask       [][]float64
	)

	flag.Parse()

	// Obtain colour ranges
	r1From, r1To, err := parseRange(*Range1)
	handle(err)
	r2From, r2To, err := parseRange(*Range2)
	handle(err)

	// Parse mask
	if *MaskMatrix != "" {
		err = json.Unmarshal([]byte(*MaskMatrix), &mask)
		if err != nil {
			log.Println("Error parsing mask : ", err)
			return
		}
	}

	// Decode images
	// If no image path is provided use a uniformly coloured background.
	if len(flag.Args()) > 0 {
		img1, err = getImage(flag.Arg(0))
	} else {
		img1 = createUniformImage(color.White, image.Rect(0, 0, 500, 500))
	}
	if len(flag.Args()) > 1 {
		img2, err = getImage(flag.Arg(1))
	} else {
		img2 = createUniformImage(color.Black, img1.Bounds())
	}
	handle(err)

	// Set output destination
	if *OutputPath == "" {
		*OutputPath = "output.png"
	}
	out, err = os.Create(*OutputPath)
	handle(err)
	defer out.Close()

	if *Width > 0 || *Height > 0 {
		img2 = resize.Resize(1024, 0, img2, resize.Lanczos3)
		img1 = resize.Resize(1024, 0, img1, resize.Lanczos3)
	}

	dp.Encode(
		out,
		dp.MergeImages(
			dp.LevelImage(img1, uint8(r2From), uint8(r2To)),
			dp.LevelImage(img2, uint8(r1From), uint8(r1To)),
			mask,
		),
		uint32(*Gama),
	)
}
