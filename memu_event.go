package memu

import (
	"fmt"
	"golang.org/x/net/websocket"
	"github.com/wadahana/memu/log"
)

const (
	MouseDown  int = 2
	MouseUp    int = 3
	MouseWheel int = 4
	MouseMove  int = 5

	LeftButton   int = 0
	MiddleButton int = 1
	RightButton  int = 2
)

type Event interface {
	GetEventType() int
	GetPayload() []byte
}

type MouseEvent struct {
	eventType   int
	mouseType 	int
	buttonType  int
	x,y 		float32
}


func NewMouseEvent(t int, x float32, y float32) Event {
	e := &MouseEvent{eventType: 1};
	e.mouseType   = t;
	e.x           = x;
	e.y 		  = y;
	return e;
}

func (e *MouseEvent) GetEventType() int {
	return e.eventType;
}

func (e *MouseEvent) GetPayload() []byte {
	buf := make([]byte, 8);
	buf[0] = 1;
	buf[1] = byte(e.mouseType);
	buf[2] = 0;
	buf[3] = 0;
	var x, y uint16;
	x = uint16(e.x * 65535)
	y = uint16(e.y * 65535)
	buf[4] = byte((x >> 8) & 0xFF);
	buf[5] = byte(x & 0xFF);
	buf[6] = byte((y >> 8) & 0xFF);
	buf[7] = byte(y & 0xFF);
	return buf;
}


type KeyboardEvent struct {
	eventType   int
	keyType     int
	keyCode     int
}


type EventAgent struct {
	index int
	port int
	retry int
	conn *websocket.Conn
	events chan Event
	stop   chan struct{}
}

func newEventAgent(index int) *EventAgent {
	return &EventAgent{
		index: index,
		conn: nil,
		port: 11001 + index * 10,
		retry: 0,
		events: make(chan Event, 10),
		stop:   make(chan struct{}),
	};
}

func (agent *EventAgent) Start() {
	go func() {
		for {
			select {
			case <-agent.stop:
				close(agent.events)
				return
			case e := <-agent.events:
				if e.GetEventType() == 1 {
					mouseEvent := e.(*MouseEvent);

					if mouseEvent.mouseType != MouseMove {
						log.Debugf("MouseEvent: %02x,%02x, %0.4f, %0.4f\n", mouseEvent.GetEventType(), mouseEvent.mouseType, mouseEvent.x, mouseEvent.y);
					}
					
					if agent.retry < 20 {
						var err error = nil;
						if agent.conn != nil {
							//_, err = agent.conn.Write(e.GetPayload());
							err = websocket.Message.Send(agent.conn, e.GetPayload());
						}
						if err != nil || agent.conn == nil {
							log.Debugf("err: %v", err);
							err = agent.Connect()
							if err != nil {
								log.Fatalf("websock (%d) reconnect fail, retry (%d), err: %v\n", agent.index, agent.retry, err);
								agent.retry += 1;
							} else {
								log.Debugf("websock (%d) reconnect success.", agent.index);
								agent.retry = 0;
							}
						}
					}

				} else {
					log.Debugf("Unknown Event");
				}
			default:
				if agent.conn == nil && agent.retry < 20 {
					log.Debugf("default connect.");
					err := agent.Connect()
					if err != nil {
						log.Fatalf("websock (%d) connect fail, retry (%d) err: %v\n", agent.index, agent.retry, err);
						agent.retry += 1;
					} else {
						log.Infof("websock (%d) connect success.", agent.index);
						agent.retry = 0;
					}
				}
			}
		}
	}();
}

func (agent *EventAgent) Connect() error {
	var err error = nil;
	origin := fmt.Sprintf("http://127.0.0.1:%d", agent.port);
	url := fmt.Sprintf("ws://127.0.0.1:%d/", agent.port);
	agent.conn, err = websocket.Dial(url, "", origin)
	return err
}

func (agent *EventAgent) Stop() {
	close(agent.stop);
}
func (agent *EventAgent) Send(ev Event) {
	agent.events <- ev
}