package pulse

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/jfreymuth/pulse/proto"

	"github.com/ftl/hamdeck/pkg/hamdeck"
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

func NewClient() *PulseClient {
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
		retryInterval:   5 * time.Second,
		done:            make(chan struct{}),
	}

	go result.handleSubscribeEvents()

	return result
}

type PulseClient struct {
	conn            net.Conn
	client          *proto.Client
	props           proto.PropList
	subscribeEvents chan *proto.SubscribeEvent

	retryInterval           time.Duration
	connected               bool
	onPulseConnectionClosed func(interface{})
	done                    chan struct{}

	listeners []interface{}
}

func (c *PulseClient) KeepOpen() {
	go func() {
		disconnected := make(chan bool, 1)
		for {
			err := c.connect(func() {
				disconnected <- true
			})
			if err == nil {
				select {
				case <-disconnected:
					log.Print("Connection lost to pulseaudio, waiting for retry.")
				case <-c.done:
					log.Print("Connection to pulseaudio closed.")
					return
				}
			} else {
				log.Printf("Cannot connect to pulseaudio, waiting for retry: %v", err)
			}

			select {
			case <-time.After(c.retryInterval):
				log.Print("Retrying to connect to pulseaudio")
			case <-c.done:
				log.Print("Connection to pulseaudio closed.")
				return
			}
		}
	}()
}

func (c *PulseClient) Connect() error {
	return c.connect(nil)
}

func (c *PulseClient) connect(whenClosed func()) error {
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

	var mask uint32 = paSubscriptionMaskSink | paSubscriptionMaskSource | paSubscriptionMaskSinkInput | paSubscriptionMaskSourceOutput
	err = c.client.Request(&proto.Subscribe{Mask: mask}, nil)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("cannot subscribe to sink and source events: %w", err)
	}

	c.connected = true
	hamdeck.NotifyEnablers(c.listeners, true)
	log.Print("Connected to pulseaudio.")

	c.client.Callback = func(msg interface{}) {
		switch msg := msg.(type) {
		case *proto.SubscribeEvent:
			c.subscribeEvents <- msg
		default:
			log.Print("unknown message type ", msg)
		}
	}

	c.onPulseConnectionClosed = func(event interface{}) {
		if _, isConnectionClosed := event.(*proto.ConnectionClosed); !isConnectionClosed {
			return
		}

		c.connected = false
		hamdeck.NotifyEnablers(c.listeners, false)

		if whenClosed != nil {
			whenClosed()
		}
	}
	c.client.Callback = c.onPulseConnectionClosed
	return nil
}

func (c *PulseClient) Close() {
	close(c.done)
	if c.connected {
		c.conn.Close()
		c.onPulseConnectionClosed(&proto.ConnectionClosed{})
	}
}

func (c *PulseClient) Connected() bool {
	return c.connected
}

/*
	Subscibe Events
*/

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
		case paSubscriptionEventSinkInput:
			c.handleSinkInputChange(index)
		case paSubscriptionEventSourceOutput:
			c.handleSourceOutputChange(index)
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

	c.notifyMuteListeners(infoReply.SinkName, infoReply.Mute)
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

	c.notifyMuteListeners(infoReply.SourceName, infoReply.Mute)
}

func (c *PulseClient) handleSinkInputChange(index int) {
	infoRequest := proto.GetSinkInputInfo{
		SinkInputIndex: uint32(index),
	}
	infoReply := proto.GetSinkInputInfoReply{}
	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		log.Printf("cannot get sink input info: %v", err)
		return
	}

	c.notifyMuteListeners(infoReply.MediaName, infoReply.Muted)
}

func (c *PulseClient) handleSourceOutputChange(index int) {
	infoRequest := proto.GetSourceOutputInfo{
		SourceOutpuIndex: uint32(index),
	}
	infoReply := proto.GetSourceOutputInfoReply{}
	err := c.client.Request(&infoRequest, &infoReply)
	if err != nil {
		log.Printf("cannot get source output info: %v", err)
		return
	}

	c.notifyMuteListeners(infoReply.MediaName, infoReply.Muted)
}

func (c *PulseClient) notifyMuteListeners(id string, mute bool) {
	for _, listener := range c.listeners {
		muteListener, ok := listener.(MuteListener)
		if ok {
			muteListener.SetMute(id, mute)
		}
	}
}

/*
	Sink
*/

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

/*
	Source
*/

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

/*
	Sink Input
*/

func (c *PulseClient) ToggleMuteSinkInput(mediaName string) (bool, error) {
	sinkInput, err := c.findSinkInput(mediaName)
	if err != nil {
		return false, fmt.Errorf("cannot get sink input info: %w", err)
	}

	muteRequest := proto.SetSinkInputMute{
		SinkInputIndex: sinkInput.SinkInputIndex,
		Mute:           !sinkInput.Muted,
	}
	err = c.client.Request(&muteRequest, nil)
	if err != nil {
		return false, fmt.Errorf("cannot mute sink input: %w", err)
	}

	return !sinkInput.Muted, nil
}

func (c *PulseClient) IsSinkInputMuted(mediaName string) (bool, error) {
	sinkInput, err := c.findSinkInput(mediaName)
	if err != nil {
		return false, fmt.Errorf("cannot get sink input info: %w", err)
	}

	return sinkInput.Muted, nil
}

func (c *PulseClient) findSinkInput(mediaName string) (*proto.GetSinkInputInfoReply, error) {
	listRequest := proto.GetSinkInputInfoList{}
	listReply := proto.GetSinkInputInfoListReply{}

	err := c.client.Request(&listRequest, &listReply)
	if err != nil {
		return nil, fmt.Errorf("cannot get sink input list: %w", err)
	}

	for _, reply := range listReply {
		if reply.MediaName == mediaName {
			return reply, nil
		}
	}

	return nil, fmt.Errorf("sink input %s not found", mediaName)
}

/*
	Source Output
*/

func (c *PulseClient) ToggleMuteSourceOutput(mediaName string) (bool, error) {
	sourceOutput, err := c.findSourceOutput(mediaName)
	if err != nil {
		return false, fmt.Errorf("cannot get source output info: %w", err)
	}

	muteRequest := proto.SetSourceOutputMute{
		SourceOutputIndex: sourceOutput.SourceOutpuIndex,
		Mute:              !sourceOutput.Muted,
	}
	err = c.client.Request(&muteRequest, nil)
	if err != nil {
		return false, fmt.Errorf("cannot mute source output: %w", err)
	}

	return !sourceOutput.Muted, nil
}

func (c *PulseClient) IsSourceOutputMuted(mediaName string) (bool, error) {
	sourceOutput, err := c.findSourceOutput(mediaName)
	if err != nil {
		return false, fmt.Errorf("cannot get source output info: %w", err)
	}

	return sourceOutput.Muted, nil
}

func (c *PulseClient) findSourceOutput(mediaName string) (*proto.GetSourceOutputInfoReply, error) {
	listRequest := proto.GetSourceOutputInfoList{}
	listReply := proto.GetSourceOutputInfoListReply{}

	err := c.client.Request(&listRequest, &listReply)
	if err != nil {
		return nil, fmt.Errorf("cannot get source output list: %w", err)
	}

	for _, reply := range listReply {
		if reply.MediaName == mediaName {
			return reply, nil
		}
	}

	return nil, fmt.Errorf("source output %s not found", mediaName)
}
