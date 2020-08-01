package hamlib

import (
	"github.com/ftl/rigproxy/pkg/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const (
	ConfigMode   = "mode"
	ConfigLabel  = "label"
	ConfigMode1  = "mode1"
	ConfigLabel1 = "label1"
	ConfigMode2  = "mode2"
	ConfigLabel2 = "label2"
)

const (
	SetModeButtonType    = "hamlib.SetMode"
	ToggleModeButtonType = "hamlib.ToggleMode"
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
	case ToggleModeButtonType:
		return f.createToggleModeButton(config)
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

func (f *Factory) createToggleModeButton(config map[string]interface{}) hamdeck.Button {
	mode1, haveMode1 := hamdeck.ToString(config[ConfigMode1])
	label1, _ := hamdeck.ToString(config[ConfigLabel1])
	mode2, haveMode2 := hamdeck.ToString(config[ConfigMode2])
	label2, _ := hamdeck.ToString(config[ConfigLabel])
	if !(haveMode1 && haveMode2) {
		return nil
	}

	return NewToggleModeButton(f.client, client.Mode(mode1), label1, client.Mode(mode2), label2)
}
