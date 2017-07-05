package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
)

func mustOpen(filename string) *os.File {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func mustLoadImage(filename string) image.Image {
	f := mustOpen(filename)
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

func mustSaveImage(img image.Image, output string) {
	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	png.Encode(f, img)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rate16(c uint32, r float64) uint16 {
	fc := float64(c)
	return uint16(65535 - (65535-fc)*r)
}

func main() {
	var (
		output string
	)
	flag.StringVar(&output, "output", "diff.png", "output filename")
	flag.StringVar(&output, "o", "diff.png", "output filename")
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("usage: imagediff [<option>...] <image1> <image2>")
		os.Exit(1)
	}

	img1 := mustLoadImage(args[0])
	img2 := mustLoadImage(args[1])

	dst := diffImage(img1, img2)

	mustSaveImage(dst, output)
}
