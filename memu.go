package memu

import (
	"image"
	"errors"
	"log"
)

var ErrorEmulatorNotFound error = errors.New("Emulator Not Found!");

type MEmulator struct {
//	GetDisplayBounds() image.Rectangle
//	CaptureVideo() *image.RGBA

	name string;
	index int;
	grabber *Grabber;
	agent *EventAgent;
}

func (e *MEmulator) Init() error {
	var err error = nil;
	e.grabber, err = newGrabber(e.GetName());
	if err != nil {
		return err;
	}
	e.agent = newEventAgent(e.GetIndex());
	e.agent.Start();
	return nil;
}

func (e *MEmulator) Close() {
	if e.grabber != nil {
		e.grabber.Close();
		e.grabber = nil;
	}
	if e.agent != nil {
		e.agent.Stop();
		e.agent = nil;
	}
}
func (e *MEmulator) GetDisplayBounds() image.Rectangle {
	if e.grabber != nil {
		return e.grabber.GetBounds();
	}
	return image.Rectangle{};
}

func (e *MEmulator) CaptureVideo() (*image.RGBA, error) {
	if e.grabber != nil {
		return e.grabber.CaptureVideo();
	}
	return nil, ErrorGrabberNotInit;
}

func (e *MEmulator) SendEvent(ev Event) {
	if e.agent != nil {
		e.agent.Send(ev);
	}
}
func (e *MEmulator) GetName() string {
	return e.name;
}

func (e *MEmulator) GetIndex() int {
	return e.index;
}



var emulatorMap map[string]*MEmulator = make(map[string]*MEmulator);

func StartRDP(name string, index int) {
	if _, ok := emulatorMap[name]; ok {
		return;
	}
	e := &MEmulator{};
	e.name = name;
	e.index = index;
	e.Init();
	emulatorMap[name] = e;
	// 
}

func StopRDP(name string) {
	if e, ok := emulatorMap[name]; ok {
		delete(emulatorMap, name);
		e.Close();
	}
}

func CaptureVideo(name string) (*image.RGBA, error) {
	if e, ok := emulatorMap[name]; ok {
		return e.CaptureVideo();
	} 
	return nil, ErrorEmulatorNotFound;
}

func GetEmulator(name string) (*MEmulator, error) {
	if e, ok := emulatorMap[name]; ok {
		return e, nil
	}
	return nil, ErrorEmulatorNotFound;
}

func GetEmulators() *map[string]*MEmulator {
	return &emulatorMap;
}

func init() {
    log.Println("memu init.");

}