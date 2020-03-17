package memu

import (
	"testing"
	"image"
	"image/jpeg"
	"os"
	"fmt"
)

func saveJpeg(name string, img *image.RGBA) error {
	file, err := os.Create(name);
	if err != nil {
		return err;
	}
	return jpeg.Encode(file, img, nil);
}

func Test_makeVCapKey(t *testing.T) {
	name := "MEmu_1";
	key, err := makeVCapKey(name);
	//t.Logf("name: %s, key: %s, err: %v", name, key, err);
	if err != nil || key != "qipc_sharedmemory_MEmuImgSharedMemory3714cfc48f1c4cfbaeb392a8ce8888766577229f" {
		t.Error("Test_makeVCapKey 测试失败");
	}
}

func Test_grabber(t *testing.T) {
	name := "MEmu_1";
	StartRDP(name, 1);

	vm, err := GetEmulator(name);
	if err != nil {
		t.Errorf("Test_grabber 获取虚拟机失败, err: %v", err);
	}
	bounds := vm.GetDisplayBounds();

	t.Logf("bounds: %v", bounds);

	for i := 0; i < 10; i++ {
		img, err := vm.CaptureVideo();
		if err != nil {
			//t.Printf("err: %v\n", err);
			t.Errorf("Test_grabber 抓屏失败, err: %v", err);
			break;
		} 
		t.Logf("i: %d\r\n", i);
		jpegName := fmt.Sprintf("%s_%d.jpeg", name, i);
		err = saveJpeg(jpegName, img);
		if err != nil {
			t.Errorf("Test_grabber 抓屏失败, err: %v", err);
		}
	}
	StopRDP(name);
}