package main

import (
	by "bytes"
	"compress/zlib"
	"fmt"
	"image/color"
	"io"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var PNG_SIG = []uint8{137, 80, 78, 71, 13, 10, 26, 10}
var IHDR_CODE = []uint8{'I', 'H', 'D', 'R'}
var IDAT_CODE = []uint8{'I', 'D', 'A', 'T'}
var IEND_CODE = []uint8{'I', 'E', 'N', 'D'}

func readUint32(bytes []byte) uint32 {
	ret := uint32(bytes[0]) << 24
	ret |= uint32(bytes[1]) << 16
	ret |= uint32(bytes[2]) << 8
	ret |= uint32(bytes[3])
	return ret
}

func equalBytes(b1 []byte, b2 []byte) bool {
	l := len(b1)
	if l != len(b2) {
		return false
	}
	for b := 0; b < l; b++ {
		if b1[b] != b2[b] {
			return false
		}
	}
	return true
}

type Pixel struct {
	r uint8
	g uint8
	b uint8
}

type Image struct {
	// IHDR
	w                 uint32
	h                 uint32
	bitDepth          uint8
	colorType         uint8
	compressionMethod uint8
	filterMethod      uint8
	interlaceMethod   uint8

	// IDAT
	pixels []Pixel
}

func readChunks(img *Image, bytes []byte) {

	i := uint32(0)

	pixels := []byte{}

	for i < uint32(len(bytes)) {

		length := readUint32(bytes[i : i+4])
		i += 4

		hdr := bytes[i : i+4]
		i += 4

		if equalBytes(hdr, IHDR_CODE) {
			img.w = readUint32(bytes[i : i+4])
			i += 4
			img.h = readUint32(bytes[i : i+4])
			i += 4
			img.bitDepth = bytes[i]
			i += 1
			img.colorType = bytes[i]
			i += 1
			img.compressionMethod = bytes[i]
			i += 1
			img.filterMethod = bytes[i]
			i += 1
			img.interlaceMethod = bytes[i]
			i += 1

			fmt.Println(img.w, img.h, img.bitDepth, img.colorType, img.compressionMethod, img.filterMethod, img.interlaceMethod)
		} else if equalBytes(hdr, IDAT_CODE) {
			// fmt.Println(bytes[i])
			pixels = append(pixels, bytes[i:i+length]...)
			i += length
		} else if equalBytes(hdr, IEND_CODE) {
			i += length
		} else {
			i += length
		}

		//data := bytes[i : i+length]
		// i += length

		crc := bytes[i : i+4]
		i += 4

		if 1 == 1 {
			fmt.Println("Len:", length)
			fmt.Print("Hdr: ", hdr, fmt.Sprintf("%c ", hdr))
			fmt.Println()
			//fmt.Println("Data:", data)
			fmt.Println("Crc:", crc)
			fmt.Println()
		}
	}

	zr, err := zlib.NewReader(by.NewReader(pixels))
	if err != nil {
		fmt.Println("Error creating zlib reader:", err)
		return
	}
	defer zr.Close()

	decompressedPixels, err := io.ReadAll(zr)
	if err != nil {
		fmt.Println("Error decompressing bytes:, err")
		return
	}

	for i := 0; i < len(decompressedPixels)-3; i += 3 {
		pixel := Pixel{decompressedPixels[i], decompressedPixels[i+1], decompressedPixels[i+2]}
		img.pixels = append(img.pixels, pixel)
	}
}

func main() {
	file, err := os.Open("elephant.png")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Println("Signature:", bytes[:8])

	img := Image{}
	readChunks(&img, bytes[8:])

	// return

	rl.InitWindow(1280, 720, "PNG-Decoder (Go)")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	tx := rl.LoadTextureFromImage(rl.LoadImage("elephant.png"))

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Brown)

		y := int(0)
		for i := 0; i < len(img.pixels); i++ {
			x := i % int(img.w)
			if x == 0 {
				y++
			}
			rl.DrawPixel(int32(x), int32(y), color.RGBA{img.pixels[i].r, img.pixels[i].g, img.pixels[i].b, 255})
		}

		rl.DrawTextureRec(tx, rl.NewRectangle(0, 0, float32(img.w), float32(img.h)), rl.NewVector2(float32(img.w), 0), rl.White)

		rl.EndDrawing()
	}
}
