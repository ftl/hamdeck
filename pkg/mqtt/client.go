package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const mqttWaitTimeout = 200 * time.Millisecond

func NewClient(address string, username string, password string) *Client {
	result := &Client{
		address: address,
		alive:   make(map[string]bool),
		tx:      make(map[string]bool),
		tuning:  make(map[string]bool),
		swr:     make(map[string]float64),
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", address))
	opts.SetClientID("hamdeck")
	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(10 * time.Second)

	opts.SetDefaultPublishHandler(result.messageReceived)
	opts.OnConnect = result.connected
	opts.OnConnectionLost = result.connectionLost

	result.client = mqtt.NewClient(opts)
	if token := result.client.Connect(); token.WaitTimeout(mqttWaitTimeout) && token.Error() != nil {
		log.Printf("cannot connect to MQTT broker: %v", token.Error())
	}

	return result
}

type Client struct {
	address   string
	client    mqtt.Client
	paths     []string
	listeners []interface{}

	alive  map[string]bool
	tx     map[string]bool
	tuning map[string]bool
	swr    map[string]float64
}

type atu100Data struct {
	TX     bool    `json:"txing"`
	Tuning bool    `json:"tuning"`
	SWR    float64 `json:"swr"`
}

func (c *Client) Connected() bool {
	return c.client.IsConnected()
}

func (c *Client) Disconnect() {
	c.client.Disconnect(250)
}

func (c *Client) connected(mqtt.Client) {
	log.Printf("connected to MQTT broker %s", c.address)
	for _, path := range c.paths {
		c.subscribePath(path)
	}
	hamdeck.NotifyEnablers(c.listeners, true)
}

func (c *Client) subscribePath(path string) {
	if !c.client.IsConnected() {
		return
	}
	c.client.Subscribe(fmt.Sprintf("%s/data", path), 1, nil).WaitTimeout(mqttWaitTimeout)
	c.client.Subscribe(fmt.Sprintf("%s/alive", path), 1, nil).WaitTimeout(mqttWaitTimeout)
}

func (c *Client) connectionLost(_ mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
	hamdeck.NotifyEnablers(c.listeners, false)
}

func (c *Client) messageReceived(_ mqtt.Client, msg mqtt.Message) {
	topic := strings.ToLower(msg.Topic())
	//log.Printf("received MQTT message from topic: %s", topic)
	path, suffix, ok := splitTopic(topic)
	if !ok {
		return
	}

	switch suffix {
	case "alive":
		c.SetAlive(path, string(msg.Payload()) == "true")
	case "data":
		var data atu100Data
		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Printf("invalid atu100 payload from %s: %v", topic, err)
			return
		}
		c.SetTX(path, data.TX)
		c.SetTune(path, data.Tuning)
		c.SetSWR(path, data.SWR)
	}
}

func splitTopic(topic string) (string, string, bool) {
	splitterIndex := strings.LastIndex(topic, "/")
	if splitterIndex < 1 || splitterIndex >= len(topic)-1 {
		return "", "", false
	}
	return topic[0:splitterIndex], topic[splitterIndex+1:], true
}

func (c *Client) AddPath(path string) {
	c.paths = append(c.paths, path)
	c.subscribePath(path)
}

func (c *Client) Tune(path string) {
	c.client.Publish(fmt.Sprintf("%s/cmd", path), 0, false, "1")
}

func (c *Client) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) SetAlive(path string, alive bool) {
	if c.alive[path] == alive {
		return
	}
	c.alive[path] = alive
	c.emitAlive(path, alive)
}

type AliveListener interface {
	SetAlive(path string, alive bool)
}

func (c *Client) emitAlive(path string, alive bool) {
	for _, l := range c.listeners {
		if listener, ok := l.(AliveListener); ok {
			listener.SetAlive(path, alive)
		}
	}
}

func (c *Client) SetTX(path string, tx bool) {
	if c.tx[path] == tx {
		return
	}
	c.tx[path] = tx
	c.emitTX(path, tx)
}

type TXListener interface {
	SetTX(path string, tx bool)
}

func (c *Client) emitTX(path string, tx bool) {
	for _, l := range c.listeners {
		if listener, ok := l.(TXListener); ok {
			listener.SetTX(path, tx)
		}
	}
}

func (c *Client) SetTune(path string, tuning bool) {
	if c.tuning[path] == tuning {
		return
	}
	c.tuning[path] = tuning
	c.emitTune(path, tuning)
}

type TuneListener interface {
	SetTune(path string, tuning bool)
}

func (c *Client) emitTune(path string, tuning bool) {
	for _, l := range c.listeners {
		if listener, ok := l.(TuneListener); ok {
			listener.SetTune(path, tuning)
		}
	}
}

func (c *Client) SetSWR(path string, swr float64) {
	if c.swr[path] == swr {
		return
	}
	c.swr[path] = swr
	c.emitSWR(path, swr)
}

type SWRListener interface {
	SetSWR(path string, swr float64)
}

func (c *Client) emitSWR(path string, swr float64) {
	for _, l := range c.listeners {
		if listener, ok := l.(SWRListener); ok {
			listener.SetSWR(path, swr)
		}
	}
}
