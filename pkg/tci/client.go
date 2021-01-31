package tci

import (
	"net"
	"time"

	"github.com/ftl/tci/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

func NewClient(host *net.TCPAddr) *Client {
	result := &Client{
		Client: client.KeepOpen(host, 10*time.Second),
	}
	result.Client.Notify(client.ConnectionListenerFunc(func(connected bool) {
		hamdeck.NotifyEnablers(result.listeners, connected)
	}))
	return result
}

type Client struct {
	*client.Client

	listeners []interface{}

	trx int
	vfo client.VFO
}

func (c *Client) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
	c.Client.Notify(listener)
}

func (c *Client) SetTRX(trx int) {
	c.trx = trx
	c.emitTRX(c.trx)
}

type TRXListener interface {
	SetTRX(trx int)
}

func (c *Client) emitTRX(trx int) {
	for _, l := range c.listeners {
		if listener, ok := l.(TRXListener); ok {
			listener.SetTRX(trx)
		}
	}
}

func (c *Client) TRX() int {
	return c.trx
}

func (c *Client) SetVFO(vfo client.VFO) {
	c.vfo = vfo
	c.emitVFO(c.vfo)
}

func (c *Client) VFO() client.VFO {
	return c.vfo
}

type VFOListener interface {
	SetVFO(vfo client.VFO)
}

func (c *Client) emitVFO(vfo client.VFO) {
	for _, l := range c.listeners {
		if listener, ok := l.(VFOListener); ok {
			listener.SetVFO(vfo)
		}
	}
}
