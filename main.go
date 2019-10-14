package main

/*
#cgo LDFLAGS: -lavcodec -lavutil

// copyright (c) 2001 Fabrice Bellard
//
// This file is part of Libav.
//
// Libav is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 2.1 of the License, or (at your option) any later version.
//
// Libav is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with Libav; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "libavcodec/avcodec.h"
#include "libavutil/frame.h"
#include "libavutil/imgutils.h"
static void encode(AVCodecContext *enc_ctx, AVFrame *frame, AVPacket *pkt,
				FILE *outfile)
{
	int ret;
	// send the frame to the encoder
	ret = avcodec_send_frame(enc_ctx, frame);
	if (ret < 0) {
		fprintf(stderr, "error sending a frame for encoding\n");
		exit(1);
	}
	while (ret >= 0) {
		ret = avcodec_receive_packet(enc_ctx, pkt);
		if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF)
			return;
		else if (ret < 0) {
			fprintf(stderr, "error during encoding\n");
			exit(1);
		}
		printf("encoded frame %3"PRId64" (size=%5d)\n", pkt->pts, pkt->size);
		fwrite(pkt->data, 1, pkt->size, outfile);
		av_packet_unref(pkt);
	}
}

struct Encoder {
	int width;
	int height;
	int bit_rate;
	int gop_size;
	int fps;
	int frame;
};

int onFrame(struct Encoder*, uint8_t*, uint8_t*, uint8_t*, int, int, int);

static int save_video(struct Encoder *encoder, const char *filename)
{
	const AVCodec *codec;
	AVCodecContext *c= NULL;
	int i, ret, x, y;
	FILE *f;
	AVFrame *picture;
	AVPacket *pkt;
	uint8_t endcode[] = { 0, 0, 1, 0xb7 };
	avcodec_register_all();
	// find the mpeg1video encoder
	codec = avcodec_find_encoder(AV_CODEC_ID_H264);
	if (!codec) {
		fprintf(stderr, "codec not found\n");
		exit(1);
	}
	c = avcodec_alloc_context3(codec);
	picture = av_frame_alloc();
	pkt = av_packet_alloc();
	if (!pkt)
		exit(1);
	// put sample parameters
	c->bit_rate = encoder->bit_rate;
	// resolution must be a multiple of two
	c->width = encoder->width;
	c->height = encoder->height;
	// frames per second
	c->time_base = (AVRational){1, encoder->fps};
	c->framerate = (AVRational){encoder->fps, 1};
	c->gop_size = encoder->gop_size; // emit one intra frame every these frames
	c->max_b_frames=1;
	c->pix_fmt = AV_PIX_FMT_YUV420P;
	// open it
	if (avcodec_open2(c, codec, NULL) < 0) {
		fprintf(stderr, "could not open codec\n");
		exit(1);
	}
	f = fopen(filename, "wb");
	if (!f) {
		fprintf(stderr, "could not open %s\n", filename);
		exit(1);
	}
	picture->format = c->pix_fmt;
	picture->width  = c->width;
	picture->height = c->height;
	ret = av_frame_get_buffer(picture, 32);
	if (ret < 0) {
		fprintf(stderr, "could not alloc the frame data\n");
		exit(1);
	}
	// encode 1 second of video
	int end = 0;
	int frame = 0;
	while (!end) {
		fflush(stdout);
		// make sure the frame data is writable
		ret = av_frame_make_writable(picture);
		if (ret < 0)
			exit(1);
		end = onFrame(encoder, picture->data[0], picture->data[1], picture->data[2], picture->linesize[0], picture->linesize[1], picture->linesize[2]);
		picture->pts = i;
		// encode the image
		encode(c, picture, pkt, f);
	}
	// flush the encoder
	encode(c, NULL, pkt, f);
	// add sequence end code to have a real MPEG file
	fwrite(endcode, 1, sizeof(endcode), f);
	fclose(f);
	avcodec_free_context(&c);
	av_frame_free(&picture);
	av_packet_free(&pkt);
	return 0;
}

*/
import "C"

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"
)

func uint8CArrayToGoSlice(p *C.uint8_t, l int) []uint8 {
	var result []uint8
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&result))
	slice.Cap = l
	slice.Len = l
	slice.Data = uintptr(unsafe.Pointer(p))
	return result
}

type encoder = C.struct_Encoder

func newEncoder(width, height, bitRate, gopSize, fps int) *encoder {
	return &encoder{
		width:    C.int(width),
		height:   C.int(height),
		bit_rate:   C.int(bitRate),
		gop_size:   C.int(gopSize),
		fps:      C.int(fps),
	}
}

func (e *encoder) EncodeTo(filename string) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	C.save_video(e, cFilename)
}

func (e *encoder) onDraw(dataY, dataU, dataV []uint8, linesizeY, linesizeU, linesizeV, width, height, frame int) bool {
	// Y
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dataY[y*linesizeY+x] = uint8(x + y + frame*3)
		}
	}
	// Cb and Cr
	for y := 0; y < height/2; y++ {
		for x := 0; x < width/2; x++ {
			dataU[y*linesizeU+x] = uint8(128 + y + frame*2)
			dataV[y*linesizeV+x] = uint8(64 + x + frame*5)
		}
	}
	return frame+1 == 30
}

//export onFrame
func onFrame(e *encoder, dataY, dataU, dataV *C.uint8_t, linesizeY, linesizeU, linesizeV C.int) C.int {
	width := int(e.width)
	height := int(e.height)
	frame := int(e.frame)
	ly := int(linesizeY)
	lu := int(linesizeU)
	lv := int(linesizeV)
	dy := uint8CArrayToGoSlice(dataY, height*ly)
	du := uint8CArrayToGoSlice(dataU, height*lu)
	dv := uint8CArrayToGoSlice(dataV, height*lv)
	result := e.onDraw(dy, du, dv, ly, lu, lv, width, height, frame)
	e.frame++
	if result {
		return 1
	}
	return 0
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output file>\n", os.Args[0])
		os.Exit(1)
	}
	e := newEncoder(1280, 720, 4 * 1024 * 1024, 10, 30)
	e.EncodeTo(os.Args[1])
}
