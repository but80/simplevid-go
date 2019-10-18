package simplevid_test

import (
	"fmt"
	"os"

	"github.com/but80/simplevid-go"
)

func ExampleCallbackEncoder() {
	filename := "example-callback-encoder.mp4"
	os.Remove(filename)

	opts := simplevid.EncoderOptions{
		Width:   1280,
		Height:  720,
		BitRate: 4 * 1024 * 1024,
		GOPSize: 10,
		FPS:     30,
	}
	e := simplevid.NewCallbackEncoder(opts, func(e simplevid.CallbackEncoder) bool {
		frame := e.Frame()
		if frame == 30 {
			return false
		}
		opts := e.Options()
		for y := 0; y < opts.Height; y++ {
			for x := 0; x < opts.Width; x++ {
				e.SetRGB(x, y,
					x+y+frame*3,
					128+y/2+frame*2,
					64+x/2+frame*5,
				)
			}
		}
		return true
	})
	if err := e.EncodeToFile(filename); err != nil {
		panic(err)
	}

	if _, err := os.Stat(filename); err != nil {
		panic(err)
	}
	fmt.Printf("%s is created.\n", filename)

	// Output:
	// example-callback-encoder.mp4 is created.
}
