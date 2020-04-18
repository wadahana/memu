package memu

import (
//	"bytes"
//	"fmt"
	"image"
	"syscall"
	"unsafe"
	"io"

	"github.com/wadahana/memu/log"
)

/*
#cgo LDFLAGS: -L/root/shared/ga-server/third_party/MEmuVCapKey -lMEmuVCapKey -lstdc++
//#cgo LDFLAGS: -L/Users/eric/workspace/docker/ubuntu-dev/shared/ga-server/third_party/MEmuVCapKey -lMEmuVCapKey -lstdc++
#include <stdio.h>
#include <stdlib.h>
#include <inttypes.h>
int makeVCapKey(const char * name, char * key, size_t key_len);

///// "MEmu_1_ImgSharedMemory\200" =  "qipc_sharedmemory_MEmuImgSharedMemory3714cfc48f1c4cfbaeb392a8ce8888766577229f"

*/
import "C"

type Grabber struct {
	io.Closer
	key 		string
	handleMap  	syscall.Handle
	buffer      uintptr
	width       uint32
	height      uint32
	bitrate     int;
	frameRate   int;
	bounds      image.Rectangle
	imageSize   uint32
};



const TRUE    BOOL = 1
const FALSE   BOOL = 0
const NULL    uintptr = 0

type BOOL		int32

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var procOpenFileMapping = kernel32.NewProc("OpenFileMappingW")

func typeToBool(b BOOL) bool {
	if b != FALSE {
		return true
	}
	return false
}
func typeToBOOL(b bool) BOOL {
	if b {
		return TRUE
	}
	return FALSE
}
// errno to error
func errnoErr(e syscall.Errno) *MEmuError {
	return 	NewError(RC_SystemError, e.Error());
}

func openFileMapping(access uint32, possession bool, name *uint16) (handle syscall.Handle, err *MEmuError) {
	err = nil;
	ret, _, errno := procOpenFileMapping.Call(uintptr(access), uintptr(typeToBOOL(possession)), uintptr(unsafe.Pointer(name)))
	handle = syscall.Handle(ret)
	log.Debugf("openFileMapping-> handle(%x), errno(%d)", handle, errno);
	if handle == 0 {
		if errno.(syscall.Errno) != 0 {
			err = errnoErr(errno.(syscall.Errno))
		} else {
			err = ErrorInvalidArgument
		}
	}
	return handle, err
}

func makeVCapKey(name string) (string, *MEmuError) {
	c_name := C.CString(name);
	defer C.free(unsafe.Pointer(c_name));
	key := make([]byte, 256);
	result := int(C.makeVCapKey(c_name, (*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key))));
	if result <= 0 {
		return "", ErrorMakeGrabberKeyFail
	}
	key = key[:result];
	return string(key), nil;
}

func createImage(rect image.Rectangle) (img *image.RGBA, e *MEmuError) {
	img = nil
	e = ErrorCreateImageFail;

	defer func() {
		err := recover()
		if err == nil {
			e = nil
		}
	}()
	// image.NewRGBA may panic if rect is too large.
	img = image.NewRGBA(rect)

	return img, e
}

func newGrabber(name string, framerate int, bitrate int) (*Grabber, *MEmuError) {
	var err *MEmuError = nil;
	var _err error = nil;
	grabber := new(Grabber);
	grabber.key, err = makeVCapKey(name);
	if err != nil {
		log.Errorf("Can't make Grabber mmap key, err: %v", err);
		return nil, ErrorMakeGrabberKeyFail;
	}
	log.Infof("key: %s\n", grabber.key);
	var mode uint32 = syscall.FILE_MAP_WRITE | syscall.FILE_MAP_READ;
	grabber.handleMap, err = openFileMapping(mode, false, syscall.StringToUTF16Ptr(grabber.key));
	if err != nil || grabber.handleMap == syscall.InvalidHandle {
		log.Errorf("Grabber OpenFileMapping fail, err: %v", err);
		return nil, ErrorOpenFileMapFail;
	}
	grabber.buffer, _err = syscall.MapViewOfFile(grabber.handleMap, mode, 0, 0, 0);
	if _err != nil || grabber.buffer == NULL {
		log.Errorf("grabber MapViewOfFile fail, err: %v", _err);
		return nil, ErrorOpenFileMapFail;
	}
	var pW unsafe.Pointer = unsafe.Pointer(grabber.buffer);
	var pH unsafe.Pointer = unsafe.Pointer(grabber.buffer + 4);
	var width uint32 = *(*uint32)(pW);
	var height uint32 = *(*uint32)(pH);
	grabber.imageSize = width * height * 4;
	grabber.bounds.Min.X = 0;
	grabber.bounds.Min.Y = 0;
	grabber.bounds.Max.X = int(width);
	grabber.bounds.Max.Y = int(height);
	grabber.frameRate = framerate;
	grabber.bitrate = bitrate;
	return grabber, nil;
}

func (grabber *Grabber) FrameRate() int {
	return grabber.frameRate;
}

func (grabber *Grabber) Bitrate() int {
	return grabber.bitrate;
}

func (grabber *Grabber) Close() *MEmuError {
	var err error = nil;
	if grabber.buffer != NULL {
		err = syscall.UnmapViewOfFile(grabber.buffer);
		grabber.buffer = NULL;
		if err != nil {
			log.Infof("UnmapViewOfFile fail, err: %v", err);
		}
	}
	if grabber.handleMap != syscall.InvalidHandle {
		syscall.CloseHandle(grabber.handleMap);
		grabber.handleMap = syscall.InvalidHandle;
		if err != nil {
			log.Infof("CloseHandle fail, err: %v", err);
		}
	}
	if err != nil {
		return NewError(RC_SystemError, err.Error());
	}
	return nil;
}

func (grabber *Grabber) CaptureVideo() (*image.RGBA, *MEmuError) {
	if grabber.buffer != NULL && grabber.imageSize > 0 {
		img, err := createImage(grabber.GetBounds())
		if err != nil {
			return nil, err
		}
		width := grabber.GetWidth();
		height := grabber.GetHeight();
		src := grabber.buffer + 8;
		for y := height - 1; y >= 0; y-- {
			i := y * width * 4;
			for x := 0; x < width; x++ {
				v0 := *(*uint8)(unsafe.Pointer(src))
				v1 := *(*uint8)(unsafe.Pointer(src + 1))
				v2 := *(*uint8)(unsafe.Pointer(src + 2))
				
				// BGRA => RGBA, and set A to 255
				img.Pix[i+0] = v0
				img.Pix[i+1] = v1
				img.Pix[i+2] = v2
				img.Pix[i+3] = 255
				i += 4
				src += 4
			}
		}
		return img, nil;
	}
	return nil, ErrorGrabberNotInit;
}

func (grabber *Grabber) GetWidth() int {
	return grabber.bounds.Max.X - grabber.bounds.Min.X;
}

func (grabber *Grabber) GetHeight() int {
	return grabber.bounds.Max.Y - grabber.bounds.Min.Y;
}

func (grabber *Grabber) GetBounds() image.Rectangle { 	
	return grabber.bounds
}

func (grabber *Grabber) Running() bool {
	if grabber.buffer != NULL && grabber.handleMap != syscall.InvalidHandle {
		return true;
	}
	return false;
}
