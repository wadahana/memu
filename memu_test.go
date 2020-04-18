package memu

import (
	"testing"
	"image"
	"image/jpeg"
	"os"
	"fmt"
	"time"
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

func Test_RDP(t *testing.T) {
	name := "MEmu_1";
	StartRDP(name, 1);

	vm, err := GetEmulator(name);
	if err != nil {
		t.Errorf("Test_RDP 获取虚拟机失败, err: %v", err);
	}
	bounds := vm.GetDisplayBounds();

	t.Logf("bounds: %v", bounds);

	for i := 0; i < 10; i++ {
		img, err := vm.CaptureVideo();
		if err != nil {
			//t.Printf("err: %v\n", err);
			t.Errorf("Test_RDP 抓屏失败, err: %v", err);
			break;
		} 
		t.Logf("i: %d\r\n", i);
		jpegName := fmt.Sprintf("%s_%d.jpeg", name, i);
		err = saveJpeg(jpegName, img);
		if err != nil {
			t.Errorf("Test_RDP 抓屏失败, err: %v", err);
		}
	}
	var times = 30
	var one float32 = 1.0;
	for i := 0; i < times; i++ {
		t := 5
		if i == 0 {
			t = 2
		} else if i == (times - 1) {
			t = 3
		}
		y := float32(0.3);
		x := float32(one * float32(i) / float32(times) + 0.1) 
		ev := NewMouseEvent(t, x, y)
		vm.SendEvent(ev)
		time.Sleep(time.Millisecond * time.Duration(10));
	}

	StopRDP(name);
}
