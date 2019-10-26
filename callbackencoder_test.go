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
				e.SetY(x, y, uint8(x+y+frame*3))
			}
		}
		for y := 0; y < opts.Height/2; y++ {
			for x := 0; x < opts.Width/2; x++ {
				e.SetU(x, y, uint8(128+y+frame*2))
				e.SetV(x, y, uint8(64+x+frame*5))
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
