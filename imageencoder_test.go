package simplevid_test

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/but80/simplevid-go"
)

func ExampleImageEncoder() {
	filename := "example-image-encoder.mp4"
	os.Remove(filename)

	opts := simplevid.EncoderOptions{
		Width:   1280,
		Height:  720,
		BitRate: 4 * 1024 * 1024,
		GOPSize: 10,
		FPS:     30,
	}
	ch := make(chan image.Image, 10)
	go func() {
		for frame := 0; frame < 30; frame++ {
			img := image.NewRGBA(image.Rect(0, 0, opts.Width, opts.Height))
			for y := 0; y < opts.Height; y++ {
				for x := 0; x < opts.Width; x++ {
					img.Set(x, y, color.RGBA{
						R: uint8(x + y + frame*3),
						G: uint8(128 + y/2 + frame*2),
						B: uint8(64 + x/2 + frame*5),
						A: 255,
					})
				}
			}
			ch <- img
		}
		close(ch)
	}()
	e := simplevid.NewImageEncoder(opts, ch)
	if err := e.EncodeToFile(filename); err != nil {
		panic(err)
	}

	if _, err := os.Stat(filename); err != nil {
		panic(err)
	}
	fmt.Printf("%s is created.\n", filename)

	// Output:
	// example-image-encoder.mp4 is created.
}
