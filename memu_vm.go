package memu

import (
	"image"
)

type MEmuInfo struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"` 
	Running  bool    `json:"running"`
	Storage  int64   `json:"storage"` 
}

type MEmulator struct {
//	GetDisplayBounds() image.Rectangle
//	CaptureVideo() *image.RGBA
	index int;
	name string;
	grabber *Grabber;
	agent *EventAgent;
}

func NewMEmulator(index int, name string) *MEmulator {
	e := &MEmulator {
		index: index,
		name: name, 
		grabber: nil,
		agent: nil,
	};
	return e;
}

func (e *MEmulator) StartRDP(frameRate int, bitrate int) *MEmuError {
	var err *MEmuError = nil;
	e.grabber, err = newGrabber(e.GetName(), frameRate, bitrate);
	if err != nil {
		return err;
	}
	e.agent = newEventAgent(e.GetIndex());
	e.agent.Start();
	return nil
}

func (e * MEmulator) StopRDP() {
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

func (e *MEmulator) CaptureVideo() (*image.RGBA, *MEmuError) {
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

func (e *MEmulator) GetFrameRate() int {
	if e.grabber != nil {
		return e.grabber.FrameRate()
	}
	return 0;
}

func (e *MEmulator) GetBitrate() int {
	if e.grabber != nil {
		return e.grabber.Bitrate()
	}
	return 0;
}