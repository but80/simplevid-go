# simplevid-go

simplevid-go は、とても簡単に利用できるGo言語向けビデオエンコーダです。

- エンコーダに渡したコールバックが1フレームごとに呼び出されるので、その中で画素を描画するだけでビデオを作成することができます。
- 現時点では、フォーマットは H264 YUV444P に固定されています。
- 使用例は [simplevid_test.go](simplevid_test.go) をお読みください。

## 依存ライブラリ

libavcodec, libavutil, libavformat が必要です。

```bash
# Ubuntu
sudo apt install libavcodec-dev libavutil-dev libavformat-dev
```

## ライセンス

[libav公式サンプル](https://libav.org/documentation/doxygen/master/encode_video_8c-example.html) をベースに実装しているため、これと同様に [LGPL 2.1 or later](LICENSE) とします。
