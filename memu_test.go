package memu

import (
	"testing"
)


func Test_makeVCapKey(t *testing.T) {
	vmName := "MEmu_1";
	key, err := makeVCapKey(vmName);
	//t.Logf("name: %s, key: %s, err: %v", vmName, key, err);
	if err != nil || key != "qipc_sharedmemory_MEmuImgSharedMemory3714cfc48f1c4cfbaeb392a8ce8888766577229f" {
		t.Error("Test_makeVCapKey 测试失败");
	}
}