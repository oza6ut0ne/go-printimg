package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/image/draw"
)

const (
	DEFAULT_TERM_W int = 80
	DEFAULT_TERM_H int = 24
)

func ResizeRGBA(img *image.RGBA) *image.RGBA {
	fd := int(os.Stdin.Fd())
	termW, termH, err := terminal.GetSize(fd)
	if err != nil {
		termW, termH = DEFAULT_TERM_W, DEFAULT_TERM_H
	}

	termH -= 1

	bounds := (*img).Bounds()
	imgW := bounds.Max.X
	imgH := bounds.Max.Y

	h := termH
	w := imgW * termH / imgH

	if newW := termW / 2; w > newW {
		h = h * newW / w
		w = newW
	}

	resizedImg := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.NearestNeighbor.Scale(resizedImg, resizedImg.Bounds(), img, bounds, draw.Over, nil)
	return resizedImg
}

func ResizeImg(img *image.Image) *image.RGBA {
	rgba := image.NewRGBA((*img).Bounds())
	draw.Draw(rgba, rgba.Rect, *img, image.Pt(0, 0), draw.Src)
	return ResizeRGBA(rgba)
}

func PrintImg(img *image.RGBA) {
	rect := img.Rect
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8
			fmt.Printf("\x1b[48;2;%d;%d;%dm  ", r, g, b)
		}
		fmt.Println()
	}
	fmt.Print("\x1b[0m")
}
