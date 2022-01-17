package main

import (
	"fmt"
	"image"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

var (
	pattern = generateBWRGBAPattern()
)

const (
	fps = 60 // Limit display FPS to this much.
)

// generateBWRGBAPattern generates a noisy pettern of black and white pixels.
func generateBWRGBAPattern() (bw [65536][64]byte) {
	var i, j uint
	for i = 0; i < 65536; i++ {
		for j = 0; j < 16; j++ {
			if i&(1<<j) > 0 {
				// Set this pixel to white.
				bw[i][j*4+0] = 0xFF
				bw[i][j*4+1] = 0xFF
				bw[i][j*4+2] = 0xFF
			}

			// Make sure to always set the alpha channel.
			bw[i][j*4+3] = 0xFF
		}
	}

	return
}

// createNoise is a fast pattern based white noise image generator.
func createNoise(drawImg *image.RGBA) {
	var rnd, rnd2 uint64
	var rnd16a, rnd16b, rnd16c, rnd16d uint16
	img := drawImg.Pix
	// Populate the image with pixel data
	for i := 0; i < len(img); i += 256 {
		rnd = uint64(rand.Int63())
		if (i % 63) == 0 {
			rnd2 = uint64(rand.Int63())
		}

		// We have to set the 64'th bit from the rand.Int63() manualy.
		rnd |= rnd2 & 1 << 63

		// Generate our indexes in the pattern image.
		rnd16a = uint16(rnd & 0x000000000000FFFF)
		rnd16b = uint16((rnd >> 16) & 0x000000000000FFFF)
		rnd16c = uint16((rnd >> 32) & 0x000000000000FFFF)
		rnd16d = uint16((rnd >> 48) & 0x000000000000FFFF)

		// Copy pattern portions to the destination image.
		copy(img[i:i+64], pattern[rnd16a][:])
		copy(img[i+64:i+128], pattern[rnd16b][:])
		copy(img[i+128:i+192], pattern[rnd16c][:])
		copy(img[i+192:i+256], pattern[rnd16d][:])

		// Rotate to next random bit.
		rnd2 = rnd2 >> 1
	}
}

func main() {
	app := app.New()

	// Destination image width * height must be a multiple of 256. We use the
	// smallest one possible here.
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))

	// Create a canvas backed by the image we will be drawing into.
	canvasImg := canvas.NewImageFromImage(img)

	// Stretch the image with the window size.
	canvasImg.FillMode = canvas.ImageFillStretch

	// Setup our window.
	w := app.NewWindow("White Noise")
	w.SetContent(canvasImg)
	w.Resize(fyne.NewSize(720, 720))
	w.CenterOnScreen()

	throttle := time.Tick(time.Second / fps) // FPS limiter.
	go func() {
		count := 0
		start := time.Now()
		for {
			<-throttle
			createNoise(img)
			canvasImg.Refresh()

			count++
			if count == fps {
				now := time.Now()
				fmt.Println("FPS :", int(
					float64(count)/now.Sub(start).Seconds()))
				start = now
				count = 0
			}
		}
	}()

	// Display window and run app.
	w.ShowAndRun()
}
