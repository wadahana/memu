package memu

import (
//	"bytes"
//	"fmt"
//	"image"
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

type MEmuVCap struct {
	io.Closer
	key 		string
	hVCapMap  	syscall.Handle
	width       uint32
	height      uint32
	imageSize   uint32
	pBuffer     uintptr
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
		return "", errors.New("make vcap key fail.");
	}
	key = key[:result];
	return string(key), nil;
}

func NewMEmuVCap(name string) *MEmuVCap {
	var err error = nil;
	vcap := new(MEmuVCap);
	vcap.key, err = makeVCapKey(name);
	if err != nil {
		log.Fatalf("Can't make vcap mmap key, err: %v", err);
		return nil;
	}
	log.Printf("key: %s\n", vcap.key);
	var mode uint32 = syscall.FILE_MAP_WRITE | syscall.FILE_MAP_READ;
	vcap.hVCapMap, err = OpenFileMapping(mode, false, syscall.StringToUTF16Ptr(vcap.key));
	//vcap.hVCapMap, err = syscall.CreateFileMapping(0, nil, syscall.PAGE_READWRITE, 0, 0, syscall.StringToUTF16Ptr(vcap.key));
	if err != nil || vcap.hVCapMap == syscall.InvalidHandle {
		log.Fatalf("vcap OpenFileMapping fail, err: %v", err);
		return nil;
	}
	vcap.pBuffer, err = syscall.MapViewOfFile(vcap.hVCapMap, mode, 0, 0, 0);
	if err != nil || vcap.pBuffer == NULL {
		log.Fatalf("vcap MapViewOfFile fail, err: %v", err);
		return nil;
	}
	var pW unsafe.Pointer = unsafe.Pointer(vcap.pBuffer);
	var pH unsafe.Pointer = unsafe.Pointer(vcap.pBuffer + 4);
	vcap.width = *(*uint32)(pW);
	vcap.height = *(*uint32)(pH);
	vcap.imageSize = vcap.width * vcap.height * 4;
	//log.Printf("%v, %v, %d, %d,", vcap.hVCapMap, vcap.pBuffer, vcap.width,vcap.height);
	//log.Printf("addr type: %v", reflect.TypeOf(addr));
	return vcap;
}

func (vcap *MEmuVCap) Close() error {
	var err error = nil;
	if vcap.pBuffer != NULL {
		err = syscall.UnmapViewOfFile(vcap.pBuffer);
		vcap.pBuffer = NULL;
		if err != nil {
			log.Printf("UnmapViewOfFile fail, err: %v", err);
		}
	}
	if vcap.hVCapMap != syscall.InvalidHandle {
		syscall.CloseHandle(vcap.hVCapMap);
		vcap.hVCapMap = syscall.InvalidHandle;
		if err != nil {
			log.Printf("CloseHandle fail, err: %v", err);
		}
	}
	return err;
}

func (vcap *MEmuVCap) Read() ([]byte, error){
	if vcap.pBuffer != NULL && vcap.imageSize > 0{
		rgba := C.GoBytes(vcap.pBuffer, vcap.imageSize);
		return rgba, nil;
	}
	return nil, errors.New("vcap buffer is empty");
}

func (vcap *MEmuVCap) GetWidth() uint32 {
	return vcap.width;
}

func (vcap *MEmuVCap) GetHeight() uint32 {
	return vcap.height;
}
