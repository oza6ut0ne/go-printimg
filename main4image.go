package main

import (
	"flag"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Println("error: no src")
		return
	}

	src, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Println("error:", err)
		return
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		log.Println("error: ", err)
	}

	resizedImg := ResizeImg(&img)
	PrintImg(resizedImg)
	return
}
