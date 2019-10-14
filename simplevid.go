package simplevid

// EncoderOptions は、ビデオエンコーダのオプションです。
type EncoderOptions struct {
	// Width は、ビデオ画面の横幅 [px] です。
	Width int
	// Height は、ビデオ画面の縦幅 [px] です。
	Height int
	// BitRate は、ビットレート [byte/sec] です。
	BitRate int
	// GOPSize は、GOP (Group Of Picture) フレーム数です。
	GOPSize int
	// FPS は、1秒あたりのフレーム数です。
	FPS int
}

// Encoder は、ビデオエンコーダです。
type Encoder interface {
	// Width は、ビデオ画面の横幅 [px] を返します。
	Width() int
	// Height は、ビデオ画面の縦幅 [px] を返します。
	Height() int
	// Frame は、現在エンコード中のフレーム番号を返します。
	Frame() int
	// LineSize は、チャンネル ch の画素データが1行あたりいくつのスライス要素を使用するかを返します。
	LineSize(ch int) int
	// Data は、チャンネル ch の画素データをスライスで返します。
	Data(ch int) []uint8
	// EncodeToFile は、ビデオをエンコードしてファイルに保存します。
	EncodeToFile(filename string) error
}
