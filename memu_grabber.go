package memu

import (
//	"bytes"
//	"fmt"
	"image"
	"syscall"
	"errors"
	"unsafe"
	"log"
	"io"
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
	bounds      image.Rectangle
	imageSize   uint32
	
};

var (
	ErrorMakeGrabberKeyFail = errors.New("Make Grabber key fail.");
 	ErrorOpenFileMapFail    = errors.New("Open File Map fail.");
	ErrorGrabberNotInit     = errors.New("Grabber Not Initialized");
	ErrorCreateImageFail    = errors.New("Cannot create image.RGBA");
) 
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
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case syscall.ERROR_IO_PENDING:
		return syscall.ERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

func openFileMapping(access uint32, possession bool, name *uint16) (handle syscall.Handle, err error) {
	ret,_,errno := procOpenFileMapping.Call(uintptr(access), uintptr(typeToBOOL(possession)), uintptr(unsafe.Pointer(name)))
	handle = syscall.Handle(ret)
	if handle == 0 {
		if errno.(syscall.Errno) != 0 {
			err = errnoErr(errno.(syscall.Errno))
		} else {
			err = syscall.EINVAL
		}
	}

	return
}

func makeVCapKey(name string) (string, error) {
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

func createImage(rect image.Rectangle) (img *image.RGBA, e error) {
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

func newGrabber(name string) (*Grabber, error) {
	var err error = nil;
	grabber := new(Grabber);
	grabber.key, err = makeVCapKey(name);
	if err != nil {
		log.Fatalf("Can't make Grabber mmap key, err: %v", err);
		return nil, ErrorMakeGrabberKeyFail;
	}
	log.Printf("key: %s\n", grabber.key);
	var mode uint32 = syscall.FILE_MAP_WRITE | syscall.FILE_MAP_READ;
	grabber.handleMap, err = openFileMapping(mode, false, syscall.StringToUTF16Ptr(grabber.key));
	if err != nil || grabber.handleMap == syscall.InvalidHandle {
		log.Fatalf("Grabber OpenFileMapping fail, err: %v", err);
		return nil, ErrorOpenFileMapFail;
	}
	grabber.buffer, err = syscall.MapViewOfFile(grabber.handleMap, mode, 0, 0, 0);
	if err != nil || grabber.buffer == NULL {
		log.Fatalf("grabber MapViewOfFile fail, err: %v", err);
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
	return grabber, nil;
}

func (grabber *Grabber) Close() error {
	var err error = nil;
	if grabber.buffer != NULL {
		err = syscall.UnmapViewOfFile(grabber.buffer);
		grabber.buffer = NULL;
		if err != nil {
			log.Printf("UnmapViewOfFile fail, err: %v", err);
		}
	}
	if grabber.handleMap != syscall.InvalidHandle {
		syscall.CloseHandle(grabber.handleMap);
		grabber.handleMap = syscall.InvalidHandle;
		if err != nil {
			log.Printf("CloseHandle fail, err: %v", err);
		}
	}
	return err;
}

func (grabber *Grabber) CaptureVideo() (*image.RGBA, error) {
	if grabber.buffer != NULL && grabber.imageSize > 0 {
		img, err := createImage(grabber.GetBounds())
		if err != nil {
			return nil, err
		}
		width := grabber.GetWidth();
		height := grabber.GetHeight();
		i := 0
		src := grabber.buffer + 8;
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				v0 := *(*uint8)(unsafe.Pointer(src))
				v1 := *(*uint8)(unsafe.Pointer(src + 1))
				v2 := *(*uint8)(unsafe.Pointer(src + 2))

				// BGRA => RGBA, and set A to 255
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255

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
