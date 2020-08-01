package hamlib

import (
	"github.com/ftl/rigproxy/pkg/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const (
	ConfigMode  = "mode"
	ConfigLabel = "label"
)

const (
	SetModeButtonType = "hamlib.SetMode"
)

func NewButtonFactory() (*Factory, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &Factory{
		client: client,
	}, nil
}

type Factory struct {
	client *HamlibClient
}

func (f *Factory) Close() {
	f.client.Close()
}

func (f *Factory) CreateButton(config map[string]interface{}) hamdeck.Button {
	switch config[hamdeck.ConfigType] {
	case SetModeButtonType:
		return f.createSetModeButton(config)
	default:
		return nil
	}
}

func (f *Factory) createSetModeButton(config map[string]interface{}) hamdeck.Button {
	mode, haveMode := hamdeck.ToString(config[ConfigMode])
	label, _ := hamdeck.ToString(config[ConfigLabel])
	if !haveMode {
		return nil
	}

	return NewSetModeButton(f.client, client.Mode(mode), label)
}
