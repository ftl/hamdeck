package pulse

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/jfreymuth/pulse/proto"
)

const (
	// subscription masks: https://freedesktop.org/software/pulseaudio/doxygen/def_8h.html#ad4e7f11f879e8c77ae5289145ecf6947
	paSubscriptionMaskSink         = 0x1
	paSubscriptionMaskSource       = 0x2
	paSubscriptionMaskSinkInput    = 0x4
	paSubscriptionMaskSourceOutput = 0x8
	paSubscriptionMaskModule       = 0x10
	paSubscriptionMaskClient       = 0x20
	paSubscriptionMaskSampleCache  = 0x40
	paSubscriptionMaskServer       = 0x80
	paSubscriptionMaskAutoload     = 0x100
	paSubscriptionMaskCard         = 0x200
	paSubscriptionMaskAll          = 0x2ff

	// subscription events: https://freedesktop.org/software/pulseaudio/doxygen/def_8h.html#a6bedfa147a9565383f1f44642cfef6a3
	paSubscriptionEventSink         = 0x0
	paSubscriptionEventSource       = 0x1
	paSubscriptionEventSinkInput    = 0x2
	paSubscriptionEventSourceOutput = 0x3
	paSubscriptionEventModule       = 0x4
	paSubscriptionEventClient       = 0x5
	paSubscriptionEventSampleCache  = 0x6
	paSubscriptionEventServer       = 0x7
	paSubscriptionEventAutoload     = 0x8
	paSubscriptionEventCard         = 0x9

	paSubscriptionEventNew    = 0x0
	paSubscriptionEventChange = 0x10
	paSubscriptionEventRemove = 0x20

	paSubscriptionEventFacilityMask = 0xf
	paSubscriptionEventTypeMask     = 0x30
)

type MuteListener interface {
	SetMute(id string, mute bool)
}

type MuteListenerFunc func(string, bool)

func (f MuteListenerFunc) SetMute(id string, mute bool) {
	f(id, mute)
}

func NewClient() (*PulseClient, error) {
	result := &PulseClient{
		props: proto.PropList{
			"media.name":                 proto.PropListString("HamDeck Audio Control"),
			"application.name":           proto.PropListString(path.Base(os.Args[0])),
			"application.icon_name":      proto.PropListString("audio-x-generic"),
			"application.process.id":     proto.PropListString(fmt.Sprintf("%d", os.Getpid())),
			"application.process.binary": proto.PropListString(os.Args[0]),
			"window.x11.display":         proto.PropListString(os.Getenv("DISPLAY")),
		},
		subscribeEvents: make(chan *proto.SubscribeEvent, 100),
	}

	err := result.reconnect()
	if err != nil {
		return nil, err
	}

	go result.handleSubscribeEvents()

	return result, nil
}

type PulseClient struct {
	conn            net.Conn
	client          *proto.Client
	props           proto.PropList
	subscribeEvents chan *proto.SubscribeEvent
	listeners       []interface{}
}

func (c *PulseClient) reconnect() error {
	var err error
	c.client, c.conn, err = proto.Connect("")
	if err != nil {
		return fmt.Errorf("cannot connect to pulse audio: %w", err)
	}

	err = c.client.Request(&proto.SetClientName{Props: c.props}, &proto.SetClientNameReply{})
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("cannot set the client properties: %w", err)
	}

	var mask uint32 = paSubscriptionMaskSink | paSubscriptionMaskSource
	err = c.client.Request(&proto.Subscribe{Mask: mask}, nil)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("cannot subscribe to sink and source events: %w", err)
	}

	c.client.Callback = func(msg interface{}) {
		switch msg := msg.(type) {
		case *proto.SubscribeEvent:
			c.subscribeEvents <- msg
		default:
			log.Print("unknown message type ", msg)
		}
	}

	c.client.OnConnectionClosed = func() {
		log.Print("connection to pulseaudio lost, trying to reconnect")
		retry := 1
		for {
			time.Sleep(2 * time.Second)
			err = c.reconnect()
			if err == nil {
				log.Printf("reconnected to pulseaudio after #%d retries", retry)
				return
			}
			log.Printf("reconnect to pulseaudio failed, waiting until next retry: %v", err)
			retry++
		}
	}
	return nil
}

func (c *PulseClient) Close() {
	c.conn.Close()
}

func (c *PulseClient) Listen(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *PulseClient) handleSubscribeEvents() {
	for msg := range c.subscribeEvents {
		eventType := msg.Event & paSubscriptionEventTypeMask
		facility := msg.Event & paSubscriptionEventFacilityMask
		index := int(msg.Index)

		if eventType == paSubscriptionEventRemove {
			continue
		}

		switch facility {
		case paSubscriptionEventSink:
			c.handleSinkChange(index)
		case paSubscriptionEventSource:
			c.handleSourceChange(index)
		default:
			log.Printf("unknown event facility: %d", facility)
		}
	}
}

func (c *PulseClient) handleSinkChange(index int) {
	infoRequest := proto.GetSinkInfo{
		SinkIndex: uint32(index),
	}
	infoReply := proto.GetSinkInfoReply{}
	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		log.Printf("cannot get sink info: %v", err)
		return
	}

	for _, listener := range c.listeners {
		muteListener, ok := listener.(MuteListener)
		if ok {
			muteListener.SetMute(infoReply.SinkName, infoReply.Mute)
		}
	}
}

func (c *PulseClient) handleSourceChange(index int) {
	infoRequest := proto.GetSourceInfo{
		SourceIndex: uint32(index),
	}
	infoReply := proto.GetSourceInfoReply{}
	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		log.Printf("cannot get source info: %v", err)
		return
	}

	for _, listener := range c.listeners {
		muteListener, ok := listener.(MuteListener)
		if ok {
			muteListener.SetMute(infoReply.SourceName, infoReply.Mute)
		}
	}
}

func (c *PulseClient) ToggleMuteSink(id string) (bool, error) {
	infoRequest := proto.GetSinkInfo{
		SinkIndex: proto.Undefined,
		SinkName:  id,
	}
	infoReply := proto.GetSinkInfoReply{}

	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		return false, fmt.Errorf("cannot get sink info: %w", err)
	}

	muteRequest := proto.SetSinkMute{
		SinkIndex: infoReply.SinkIndex,
		Mute:      !infoReply.Mute,
	}

	err = c.client.Request(&muteRequest, nil)
	if err != nil {
		return false, fmt.Errorf("cannot mute sink: %w", err)
	}

	return !infoReply.Mute, nil
}

func (c *PulseClient) ToggleMuteSource(id string) (bool, error) {
	infoRequest := proto.GetSourceInfo{
		SourceIndex: proto.Undefined,
		SourceName:  id,
	}
	infoReply := proto.GetSourceInfoReply{}

	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		return false, fmt.Errorf("cannot get source info: %w", err)
	}

	muteRequest := proto.SetSourceMute{
		SourceIndex: infoReply.SourceIndex,
		Mute:        !infoReply.Mute,
	}

	err = c.client.Request(&muteRequest, nil)
	if err != nil {
		return false, fmt.Errorf("cannot mute source: %w", err)
	}

	return !infoReply.Mute, nil
}

func (c *PulseClient) IsSinkMuted(id string) (bool, error) {
	infoRequest := proto.GetSinkInfo{
		SinkIndex: proto.Undefined,
		SinkName:  id,
	}
	infoReply := proto.GetSinkInfoReply{}

	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		return false, fmt.Errorf("cannot get sink info: %w", err)
	}

	return infoReply.Mute, nil
}

func (c *PulseClient) IsSourceMuted(id string) (bool, error) {
	infoRequest := proto.GetSourceInfo{
		SourceIndex: proto.Undefined,
		SourceName:  id,
	}
	infoReply := proto.GetSourceInfoReply{}

	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		return false, fmt.Errorf("cannot get source info: %w", err)
	}

	return infoReply.Mute, nil
}
