package simplevid_test

import (
	"fmt"
	"os"

	"github.com/but80/simplevid-go"
)

func ExampleEncoder() {
	filename := "example.mp4"
	os.Remove(filename)

	opts := simplevid.EncoderOptions{
		Width:   1280,
		Height:  720,
		BitRate: 4 * 1024 * 1024,
		GOPSize: 10,
		FPS:     30,
	}
	e := simplevid.NewCustomEncoder(opts, func(e simplevid.Encoder) bool {
		frame := e.Frame()
		if frame == 30 {
			return false
		}
		width := e.Width()
		height := e.Height()
		dataY := e.Data(0)
		dataU := e.Data(1)
		dataV := e.Data(2)
		lineSizeY := e.LineSize(0)
		lineSizeU := e.LineSize(1)
		lineSizeV := e.LineSize(2)
		// Y
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dataY[y*lineSizeY+x] = uint8(x + y + frame*3)
			}
		}
		// Cb and Cr
		for y := 0; y < height/2; y++ {
			for x := 0; x < width/2; x++ {
				dataU[y*lineSizeU+x] = uint8(128 + y + frame*2)
				dataV[y*lineSizeV+x] = uint8(64 + x + frame*5)
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
	// example.mp4 is created.
}
