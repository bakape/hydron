package main

// #cgo pkg-config: libavcodec libavutil libavformat
// #cgo CFLAGS: -std=c11
// #include <libavutil/pixdesc.h>
// #include "video.h"
import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"unsafe"
)

var errNoCompatibleStreams = errors.New("no compatible streams found")

// Thumbnail extracts the first frame of the video
func (c *ffContext) Thumbnail() (image.Image, error) {
	ci, err := c.codecContext(ffVideo)
	if err != nil {
		return nil, err
	}

	var f *C.AVFrame
	eErr := C.extract_video_image(&f, c.avFormatCtx, ci.ctx, ci.stream)
	if eErr != 0 {
		return nil, ffError(eErr)
	}
	if f == nil {
		return nil, errors.New("failed to get frame")
	}
	defer C.av_frame_free(&f)

	// TODO: This encoding step is redundant. Need to find a way to extract the
	// buffer directly from the frame.

	if C.GoString(C.av_get_pix_fmt_name(int32(f.format))) != "yuv420p" {
		return nil, fmt.Errorf(
			"expected format: %s; got: %s",
			image.YCbCrSubsampleRatio420,
			C.GoString(C.av_get_pix_fmt_name(int32(f.format))),
		)
	}
	y := C.GoBytes(unsafe.Pointer(f.data[0]), f.linesize[0]*f.height)
	u := C.GoBytes(unsafe.Pointer(f.data[1]), f.linesize[0]*f.height/4)
	v := C.GoBytes(unsafe.Pointer(f.data[2]), f.linesize[0]*f.height/4)

	return &image.YCbCr{
		Y:              y,
		Cb:             u,
		Cr:             v,
		YStride:        int(f.linesize[0]),
		CStride:        int(f.linesize[0]) / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect: image.Rectangle{
			Min: image.Point{
				X: 0,
				Y: 0,
			},
			Max: image.Point{
				X: int(f.width),
				Y: int(f.height) / 2 * 2,
			},
		},
	}, nil
}

// Extract data and thumbnail from a WebM video
func processWebm(data []byte) ([]byte, bool, error) {
	c, err := newFFContext(data)
	if err != nil {
		return nil, false, err
	}
	defer c.Close()

	return thumbnailVideo(c)
}

// Produce a thumbnail out of a video stream
func thumbnailVideo(c *ffContext) (buf []byte, isPNG bool, err error) {
	src, err := c.Thumbnail()
	if err != nil {
		return
	}

	var w bytes.Buffer
	err = png.Encode(&w, src)
	if err != nil {
		return
	}

	return processImage(w.Bytes())
}

// Verify the media container file (OGG, MP4, etc.) contains the supported
// stream codecs and produce an appropriate thumbnail
func processMediaContainer(data []byte) (thumb []byte, isPNG bool, err error) {
	c, err := newFFContext(data)
	if err != nil {
		return
	}
	defer c.Close()

	video, err := c.CodecName(ffVideo)
	switch {
	case err != nil:
		return
	case video == "":
		err = errUnsupportedFile
		return
	}

	return thumbnailVideo(c)
}
