package ezvid

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
static const char* encode(AVCodecContext *enc_ctx, AVFrame *frame, AVPacket *pkt, FILE *outfile) {
	int ret;
	// send the frame to the encoder
	ret = avcodec_send_frame(enc_ctx, frame);
	if (ret < 0) return "error sending a frame for encoding";
	while (ret >= 0) {
		ret = avcodec_receive_packet(enc_ctx, pkt);
		if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF) break;
		if (ret < 0) return "error during encoding";
		// printf("encoded frame %3"PRId64" (size=%5d)\n", pkt->pts, pkt->size);
		fwrite(pkt->data, 1, pkt->size, outfile);
		av_packet_unref(pkt);
	}
	return NULL;
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

static const char* initialize(struct Encoder *e, const char *filename) {
	avcodec_register_all();
	// find the mpeg1video encoder
	e->codec = avcodec_find_encoder(AV_CODEC_ID_H264);
	if (!e->codec) {
		return "codec not found";
	}
	e->c = avcodec_alloc_context3(e->codec);
	e->picture = av_frame_alloc();
	e->pkt = av_packet_alloc();
	if (!e->pkt) return "could not allocate packet";
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
		return "could not open codec";
	}
	e->f = fopen(filename, "wb");
	if (!e->f) {
		return "could not open file";
	}
	e->picture->format = e->c->pix_fmt;
	e->picture->width  = e->c->width;
	e->picture->height = e->c->height;
	if (av_frame_get_buffer(e->picture, 32) < 0) {
		return "could not alloc the frame data";
	}
	e->frame = 0;
	return NULL;
}

static const char* begin_frame(struct Encoder *e) {
	fflush(stdout);
	// make sure the frame data is writable
	if (av_frame_make_writable(e->picture) < 0) return "frame data is not writable";
	return NULL;
}

static const char* end_frame(struct Encoder *e) {
	e->picture->pts = e->frame;
	e->frame++;
	// encode the image
	return encode(e->c, e->picture, e->pkt, e->f);
}

static uint8_t endcode[] = { 0, 0, 1, 0xb7 };

static const char* finalize(struct Encoder *e) {
	// flush the encoder
	const char* result = encode(e->c, NULL, e->pkt, e->f);
	// add sequence end code to have a real MPEG file
	fwrite(endcode, 1, sizeof(endcode), e->f);
	return result;
}

static void free_resources(struct Encoder *e) {
	fclose(e->f);
	avcodec_free_context(&e->c);
	av_frame_free(&e->picture);
	av_packet_free(&e->pkt);
}

*/
import "C"

import (
	"errors"
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

type Encoder struct {
	*encoder
	onDraw func(*Encoder) bool
}

type EncoderOptions struct {
	Width   int
	Height  int
	BitRate int
	GOPSize int
	FPS     int
}

func NewEncoder(opts EncoderOptions, onDraw func(*Encoder) bool) *Encoder {
	return &Encoder{
		encoder: &encoder{
			width:    C.int(opts.Width),
			height:   C.int(opts.Height),
			bit_rate: C.int(opts.BitRate),
			gop_size: C.int(opts.GOPSize),
			fps:      C.int(opts.FPS),
		},
		onDraw: onDraw,
	}
}

func (e *Encoder) Width() int {
	return int(e.encoder.width)
}

func (e *Encoder) Height() int {
	return int(e.encoder.height)
}

func (e *Encoder) Frame() int {
	return int(e.encoder.frame)
}

func (e *Encoder) LineSize(ch int) int {
	return int(e.encoder.picture.linesize[ch])
}

func (e *Encoder) Data(ch int) []uint8 {
	return uint8CArrayToGoSlice(e.encoder.picture.data[ch], e.Height()*e.LineSize(ch))
}

func (e *Encoder) EncodeToFile(filename string) error {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	if msg := C.initialize(e.encoder, cFilename); msg != nil {
		return errors.New(C.GoString(msg))
	}
	defer C.free_resources(e.encoder)
	for {
		if msg := C.begin_frame(e.encoder); msg != nil {
			return errors.New(C.GoString(msg))
		}
		if !e.onDraw(e) {
			break
		}
		if msg := C.end_frame(e.encoder); msg != nil {
			return errors.New(C.GoString(msg))
		}
	}
	if msg := C.finalize(e.encoder); msg != nil {
		return errors.New(C.GoString(msg))
	}
	return nil
}