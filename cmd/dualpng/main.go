package main

import (
	"image"
	_ "image/jpeg"
	"log"
	"os"

	"github.com/nfnt/resize"

	dp "github.com/Necroforger/dualpng"
)

func handle(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func main() {
	f, err := os.Open("ika_musume.jpeg")
	handle(err)

	f2, err := os.Open("hakase.png")
	handle(err)

	img, _, err := image.Decode(f)
	handle(err)

	img2, _, err := image.Decode(f2)
	handle(err)

	out, err := os.Create("dual.png")
	handle(err)

	dp.Encode(
		out,
		dp.MergeImages(
			dp.LevelImage(resize.Resize(1024, 0, img, resize.Lanczos3), 0, 230),
			dp.LevelImage(resize.Resize(1024, 0, img2, resize.Lanczos3), 230, 255),
			[][]int{
				{1, 1},
				{1, 0},
			},
		),
		1000,
	)
}
