package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"io"
	"log"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type HashLines struct {
	Lines []string
}

func writeUint32(w io.Writer, n uint32) {
	w.Write([]byte{
		byte((n >> 24) & 0xff),
		byte((n >> 16) & 0xff),
		byte((n >> 8) & 0xff),
		byte((n >> 0) & 0xff),
	})
}

func readUint32(r io.Reader) uint32 {
	var bs [4]byte
	r.Read(bs[:])
	n := uint32(bs[0])
	for i := 1; i < 4; i++ {
		n = n << 8
		n += uint32(bs[i])
	}

	return n
}

func encodeLine(img image.Image, y int) string {
	buf := &bytes.Buffer{}
	w := img.Bounds().Size().X
	for x := 0; x < w; x++ {
		r, g, b, a := img.At(x, y).RGBA()
		writeUint32(buf, r)
		writeUint32(buf, g)
		writeUint32(buf, b)
		writeUint32(buf, a)
	}

	return base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

func decodeLine(s string) []color.Color {
	bs, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(bs)
	colors := make([]color.Color, 0)
	for buf.Len() != 0 {
		r := readUint32(buf)
		g := readUint32(buf)
		b := readUint32(buf)
		a := readUint32(buf)
		c := color.NRGBA64{uint16(r), uint16(g), uint16(b), uint16(a)}
		colors = append(colors, c)
	}

	return colors
}

func disassembleDiffs(diffs []diffmatchpatch.Diff) []diffmatchpatch.Diff {
	res := make([]diffmatchpatch.Diff, 0, len(diffs))

	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for _, line := range lines {
			res = append(res, diffmatchpatch.Diff{Type: diff.Type, Text: line})
		}
	}

	return res
}

func diff(from, to []string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(strings.Join(from, "\n"), strings.Join(to, "\n"))
	diffs := dmp.DiffMain(a, b, false)
	result := dmp.DiffCharsToLines(diffs, c)

	return disassembleDiffs(result)
}

func blend(dst, src color.Color) color.Color {
	srcR, srcG, srcB, srcA := src.RGBA()
	dstR, dstG, dstB, _ := dst.RGBA()
	srcAp := float64(srcA) / 0xffff

	outR := float64(srcR)*srcAp + float64(dstR)*(1-srcAp)
	outG := float64(srcG)*srcAp + float64(dstG)*(1-srcAp)
	outB := float64(srcB)*srcAp + float64(dstB)*(1-srcAp)

	return color.NRGBA64{
		uint16(outR),
		uint16(outG),
		uint16(outB),
		0xffff,
	}
}

func diffImage(img1, img2 image.Image) image.Image {
	sz1 := img1.Bounds().Size()
	sz2 := img2.Bounds().Size()
	w := max(sz1.X, sz2.X)

	lines1 := make([]string, 0, sz1.Y)
	for y := 0; y < sz1.Y; y++ {
		lines1 = append(lines1, encodeLine(img1, y))
	}

	lines2 := make([]string, 0, sz2.Y)
	for y := 0; y < sz2.Y; y++ {
		lines2 = append(lines2, encodeLine(img2, y))
	}

	diffs := diff(lines1, lines2)
	diffImg := image.NewRGBA(image.Rect(0, 0, w, len(diffs)))
	for y, diff := range diffs {
		colors := decodeLine(diff.Text)

		switch diff.Type {
		case diffmatchpatch.DiffDelete:
			red := color.RGBA64{0xffff, 0, 0, 0x4000}
			for x, c := range colors {
				diffImg.Set(x, y, blend(c, red))
			}
		case diffmatchpatch.DiffInsert:
			green := color.RGBA64{0, 0xffff, 0, 0x4000}
			for x, c := range colors {
				diffImg.Set(x, y, blend(c, green))
			}
		case diffmatchpatch.DiffEqual:
			for x, c := range colors {
				diffImg.Set(x, y, c)
			}
		}
	}

	return diffImg
}
