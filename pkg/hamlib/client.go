package hamlib

import (
	"context"
	"log"
	"time"

	"github.com/ftl/hamdeck/pkg/hamdeck"
	"github.com/ftl/rigproxy/pkg/client"
)

type ReconnectListener interface {
	Reconnected()
}

type ReconnectListenerFunc func()

func (f ReconnectListenerFunc) Reconnect() {
	f()
}

func NotifyReconnectListeners(listeners []interface{}) {
	for _, listener := range listeners {
		reconnectListener, ok := listener.(ReconnectListener)
		if ok {
			reconnectListener.Reconnected()
		}
	}
}

type VFOListener interface {
	SetVFO(vfo client.VFO)
}

type VFOListenerFunc func(client.VFO)

func (f VFOListenerFunc) SetVFO(vfo client.VFO) {
	f(vfo)
}

func NotifyVFOListeners(listeners []interface{}, vfo client.VFO) {
	for _, listener := range listeners {
		vfoListener, ok := listener.(VFOListener)
		if ok {
			vfoListener.SetVFO(vfo)
		}
	}
}

type FrequencyListener interface {
	SetFrequency(frequency client.Frequency)
}

type FrequencyListenerFunc func(client.Frequency)

func (f FrequencyListenerFunc) SetFrequency(frequency client.Frequency) {
	f(frequency)
}

func NotifyFrequencyListeners(listeners []interface{}, frequency client.Frequency) {
	for _, listener := range listeners {
		frequencyListener, ok := listener.(FrequencyListener)
		if ok {
			frequencyListener.SetFrequency(frequency)
		}
	}
}

type ModeListener interface {
	SetMode(mode client.Mode)
}

type ModeListenerFunc func(client.Mode)

func (f ModeListenerFunc) SetMode(mode client.Mode) {
	f(mode)
}

func NotifyModeListeners(listeners []interface{}, mode client.Mode) {
	for _, listener := range listeners {
		modeListener, ok := listener.(ModeListener)
		if ok {
			modeListener.SetMode(mode)
		}
	}
}

type PowerLevelListener interface {
	SetPowerLevel(powerLevel float64)
}

type PowerLevelListenerFunc func(float64)

func (f PowerLevelListenerFunc) SetPowerLevel(powerLevel float64) {
	f(powerLevel)
}

func NotifyPowerLevelListeners(listeners []interface{}, powerLevel float64) {
	for _, listener := range listeners {
		powerLevelListener, ok := listener.(PowerLevelListener)
		if ok {
			powerLevelListener.SetPowerLevel(powerLevel)
		}
	}
}

type PTTListener interface {
	SetPTT(ptt client.PTT)
}

type PTTListenerFunc func(client.PTT)

func (f PTTListenerFunc) SetPTT(ptt client.PTT) {
	f(ptt)
}

func NotifyPTTListeners(listeners []interface{}, ptt client.PTT) {
	for _, listener := range listeners {
		pttListener, ok := listener.(PTTListener)
		if ok {
			pttListener.SetPTT(ptt)
		}
	}
}

func NewClient(address string) *HamlibClient {
	return &HamlibClient{
		address:         address,
		pollingInterval: 500 * time.Millisecond,
		pollingTimeout:  2 * time.Second,
		retryInterval:   5 * time.Second,
		requestTimeout:  500 * time.Millisecond,
		done:            make(chan struct{}),
	}
}

type HamlibClient struct {
	Conn *client.Conn

	address         string
	pollingInterval time.Duration
	pollingTimeout  time.Duration
	retryInterval   time.Duration
	requestTimeout  time.Duration
	connected       bool
	closed          chan struct{}
	done            chan struct{}

	listeners []interface{}
}

func (c *HamlibClient) KeepOpen() {
	go func() {
		disconnected := make(chan bool, 1)
		for {
			err := c.connect(func() {
				disconnected <- true
			})
			if err == nil {
				select {
				case <-disconnected:
					log.Print("Connection lost to Hamlib, waiting for retry.")
				case <-c.done:
					log.Print("Connection to Hamlib closed.")
					return
				}
			} else {
				log.Printf("Cannot connect to Hamlib, waiting for retry: %v", err)
			}

			select {
			case <-time.After(c.retryInterval):
				log.Print("Retrying to connect to Hamlib")
			case <-c.done:
				log.Print("Connection to Hamlib closed.")
				return
			}
		}
	}()
}

func (c *HamlibClient) Connect() error {
	return c.connect(nil)
}

func (c *HamlibClient) connect(whenClosed func()) error {
	var err error

	c.Conn, err = client.Open(c.address)
	if err != nil {
		return err
	}

	c.closed = make(chan struct{})
	c.connected = true
	hamdeck.NotifyEnablers(c.listeners, true)

	c.Conn.StartPolling(c.pollingInterval, c.pollingTimeout,
		client.PollCommand(client.OnVFO(c.setVFO)),
		client.PollCommand(client.OnFrequency(c.setFrequency)),
		client.PollCommand(client.OnModeAndPassband(c.setModeAndPassband)),
		client.PollCommand(client.OnPowerLevel(c.setPowerLevel)),
		client.PollCommand(client.OnPTT(c.setPTT)),
	)

	c.Conn.WhenClosed(func() {
		c.connected = false
		hamdeck.NotifyEnablers(c.listeners, false)

		if whenClosed != nil {
			whenClosed()
		}

		close(c.closed)
	})

	return nil
}

func (c *HamlibClient) Close() {
	close(c.done)
	if c.connected {
		c.Conn.Close()
		<-c.closed
	}
}

func (c *HamlibClient) WithRequestTimeout() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), c.requestTimeout)
	return ctx
}

func (c *HamlibClient) Connected() bool {
	return c.connected
}

func (c *HamlibClient) setVFO(vfo client.VFO) {
	NotifyVFOListeners(c.listeners, vfo)
}

func (c *HamlibClient) setFrequency(frequency client.Frequency) {
	NotifyFrequencyListeners(c.listeners, frequency)
}

func (c *HamlibClient) setModeAndPassband(mode client.Mode, passband client.Frequency) {
	NotifyModeListeners(c.listeners, mode)
}

func (c *HamlibClient) setPowerLevel(powerLevel float64) {
	NotifyPowerLevelListeners(c.listeners, powerLevel)
}

func (c *HamlibClient) setPTT(ptt client.PTT) {
	NotifyPTTListeners(c.listeners, ptt)
}

func (c *HamlibClient) Listen(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}
