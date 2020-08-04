package hamlib

import (
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

func NewClient(address string) (*HamlibClient, error) {
	result := &HamlibClient{
		address:         address,
		pollingInterval: 500 * time.Millisecond,
		pollingTimeout:  2 * time.Second,
	}

	err := result.reconnect()
	if err != nil {
		return nil, err
	}

	return result, nil
}

type HamlibClient struct {
	Conn *client.Conn

	address         string
	pollingInterval time.Duration
	pollingTimeout  time.Duration

	listeners []interface{}
}

func (c *HamlibClient) reconnect() error {
	var err error

	c.Conn, err = client.Open(c.address)
	if err != nil {
		return err
	}

	c.Conn.StartPolling(c.pollingInterval, c.pollingTimeout,
		client.PollCommand(client.OnModeAndPassband(c.setModeAndPassband)),
		client.PollCommand(client.OnFrequency(c.setFrequency)),
		client.PollCommand(client.OnPowerLevel(c.setPowerLevel)),
	)

	c.Conn.WhenClosed(func() {
		log.Print("connection to hamlib lost, trying to reconnect")
		hamdeck.NotifyEnablers(c.listeners, false)
		retry := 1
		for {
			time.Sleep(2 * time.Second)
			err = c.reconnect()
			if err == nil {
				log.Printf("reconnected to hamlib after #%d retries", retry)
				hamdeck.NotifyEnablers(c.listeners, true)
				return
			}
			log.Printf("reconnect to hamlib failed, waiting until next retry: %v", err)
			retry++
		}
	})

	return nil
}

func (c *HamlibClient) Close() {
	c.Conn.Close()
}

func (c *HamlibClient) setModeAndPassband(mode client.Mode, passband client.Frequency) {
	NotifyModeListeners(c.listeners, mode)
}

func (c *HamlibClient) setFrequency(frequency client.Frequency) {
	NotifyFrequencyListeners(c.listeners, frequency)
}

func (c *HamlibClient) setPowerLevel(powerLevel float64) {
	NotifyPowerLevelListeners(c.listeners, powerLevel)
}

func (c *HamlibClient) Listen(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}
