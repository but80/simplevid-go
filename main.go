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

	const AVCodec *codec;
	AVCodecContext *c;
	FILE *f;
	AVFrame *picture;
	AVPacket *pkt;
	int frame;
};

static void initialize(struct Encoder *e, const char *filename) {
	avcodec_register_all();
	// find the mpeg1video encoder
	e->codec = avcodec_find_encoder(AV_CODEC_ID_H264);
	if (!e->codec) {
		fprintf(stderr, "codec not found\n");
		exit(1);
	}
	e->c = avcodec_alloc_context3(e->codec);
	e->picture = av_frame_alloc();
	e->pkt = av_packet_alloc();
	if (!e->pkt) exit(1);
	// put sample parameters
	e->c->bit_rate = e->bit_rate;
	// resolution must be a multiple of two
	e->c->width = e->width;
	e->c->height = e->height;
	// frames per second
	e->c->time_base = (AVRational){1, e->fps};
	e->c->framerate = (AVRational){e->fps, 1};
	e->c->gop_size = e->gop_size; // emit one intra frame every these frames
	e->c->max_b_frames=1;
	e->c->pix_fmt = AV_PIX_FMT_YUV420P;
	// open it
	if (avcodec_open2(e->c, e->codec, NULL) < 0) {
		fprintf(stderr, "could not open codec\n");
		exit(1);
	}
	e->f = fopen(filename, "wb");
	if (!e->f) {
		fprintf(stderr, "could not open %s\n", filename);
		exit(1);
	}
	e->picture->format = e->c->pix_fmt;
	e->picture->width  = e->c->width;
	e->picture->height = e->c->height;
	if (av_frame_get_buffer(e->picture, 32) < 0) {
		fprintf(stderr, "could not alloc the frame data\n");
		exit(1);
	}
	e->frame = 0;
}

static void begin_frame(struct Encoder *e) {
	fflush(stdout);
	// make sure the frame data is writable
	if (av_frame_make_writable(e->picture) < 0) exit(1);
}

static void end_frame(struct Encoder *e) {
	e->picture->pts = e->frame;
	e->frame++;
	// encode the image
	encode(e->c, e->picture, e->pkt, e->f);
}

static uint8_t endcode[] = { 0, 0, 1, 0xb7 };

static void finalize(struct Encoder *e) {
	// flush the encoder
	encode(e->c, NULL, e->pkt, e->f);
	// add sequence end code to have a real MPEG file
	fwrite(endcode, 1, sizeof(endcode), e->f);
	fclose(e->f);
	avcodec_free_context(&e->c);
	av_frame_free(&e->picture);
	av_packet_free(&e->pkt);
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
		bit_rate: C.int(bitRate),
		gop_size: C.int(gopSize),
		fps:      C.int(fps),
	}
}

func (e *encoder) EncodeTo(filename string) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	C.initialize(e, cFilename)
	end := false
	for !end {
		C.begin_frame(e)
		width := int(e.width)
		height := int(e.height)
		frame := int(e.frame)
		ly := int(e.picture.linesize[0])
		lu := int(e.picture.linesize[1])
		lv := int(e.picture.linesize[2])
		dy := uint8CArrayToGoSlice(e.picture.data[0], height*ly)
		du := uint8CArrayToGoSlice(e.picture.data[1], height*lu)
		dv := uint8CArrayToGoSlice(e.picture.data[2], height*lv)
		end = e.onDraw(dy, du, dv, ly, lu, lv, width, height, frame)
		C.end_frame(e)
	}
	C.finalize(e)
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

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output file>\n", os.Args[0])
		os.Exit(1)
	}
	e := newEncoder(1280, 720, 4*1024*1024, 10, 30)
	e.EncodeTo(os.Args[1])
}
