package atu100

import (
	"log"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const (
	ConfigLabel = "label"
	ConfigPath  = "path"
)

const (
	TuneButtonType = "atu100.Tune"
)

func NewButtonFactory(address string, username string, password string) *Factory {
	client := NewClient(address, username, password)

	return &Factory{
		client: client,
	}
}

type Factory struct {
	client *Client
}

func (f *Factory) Close() {
	f.client.Disconnect()
}

func (f *Factory) CreateButton(config map[string]interface{}) hamdeck.Button {
	switch config[hamdeck.ConfigType] {
	case TuneButtonType:
		return f.createTuneButton(config)
	default:
		return nil
	}
}

func (f *Factory) createTuneButton(config map[string]interface{}) hamdeck.Button {
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	path, havePath := hamdeck.ToString(config[ConfigPath])

	if !(haveLabel && havePath) {
		log.Print("A atu100.Tune button must have label and path fields.")
		return nil
	}

	return NewTuneButton(f.client, label, path)
}
