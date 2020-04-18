package memu

import (
	"image"
	"io"
	"os"
	"github.com/wadahana/memu/log"
)

type MEmuConfig struct {
	MEmuPath   string
	LoggerFile string
}

var emulatorMap map[string]*MEmulator = make(map[string]*MEmulator);

func StartRDP(name string, index int, framerate int, bitrate int) *MEmuError {
	if _, ok := emulatorMap[name]; ok {
		return nil;
	}
	e := &MEmulator{};
	e.name = name;
	e.index = index;
	err := e.StartRDP(framerate, bitrate)
	if err == nil {
		emulatorMap[name] = e;
		log.Infof("start MEmulator(%s)", name)
	}
	return err
}

func StopRDP(name string) {
	if e, ok := emulatorMap[name]; ok {
		delete(emulatorMap, name);
		e.StopRDP();
		log.Infof("stop MEmulator(%s)", name)
	}
}

func CaptureVideo(name string) (*image.RGBA, *MEmuError) {
	if e, ok := emulatorMap[name]; ok {
		return e.CaptureVideo();
	} 
	return nil, ErrorEmulatorNotFound;
}

func GetEmulator(name string) (*MEmulator, *MEmuError) {
	if e, ok := emulatorMap[name]; ok {
		return e, nil
	}
	return nil, ErrorEmulatorNotFound;
}

func GetEmulators() *map[string]*MEmulator {
	return &emulatorMap;
}


func Init(config *MEmuConfig) {
    var logOut io.Writer = nil;
    if config.LoggerFile == "console" {
    	logOut = os.Stderr
    } else {
    	var err error = nil;
    	logOut, err = os.OpenFile(config.LoggerFile, os.O_RDWR | os.O_APPEND | os.O_CREATE, 0x666);
    	if err != nil {
    		panic(err);
    	}

    }
    log.InitLogger(logOut);
    initMEmuCmd(config.MEmuPath);
}