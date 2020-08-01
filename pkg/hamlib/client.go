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

func NewClient() (*HamlibClient, error) {
	result := &HamlibClient{
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

	pollingInterval time.Duration
	pollingTimeout  time.Duration

	listeners []interface{}
}

func (c *HamlibClient) reconnect() error {
	var err error

	c.Conn, err = client.Open("")
	if err != nil {
		return err
	}

	c.Conn.StartPolling(c.pollingInterval, c.pollingTimeout,
		client.PollCommand(client.OnModeAndPassband(c.setModeAndPassband), "get_mode"),
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

func (c *HamlibClient) setModeAndPassband(mode client.Mode, passband float64) {
	NotifyModeListeners(c.listeners, mode)
}

func (c *HamlibClient) Listen(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}
