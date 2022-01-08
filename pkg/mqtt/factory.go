package mqtt

import (
	"log"
	"strings"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const (
	ConfigLabel       = "label"
	ConfigPath        = "path"
	ConfigInputTopic  = "inputTopic"
	ConfigOutputTopic = "outputTopic"
	ConfigOnPayload   = "onPayload"
	ConfigOffPayload  = "offPayload"
	ConfigMode        = "mode"
)

const (
	TuneButtonType   = "mqtt.AT100Tune"
	SwitchButtonType = "mqtt.Switch"
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
	case SwitchButtonType:
		return f.createSwitchButton(config)
	default:
		return nil
	}
}

func (f *Factory) createTuneButton(config map[string]interface{}) hamdeck.Button {
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	path, havePath := hamdeck.ToString(config[ConfigPath])

	if !(haveLabel && havePath) {
		log.Print("A mqtt.ATU100Tune button must have label and path fields.")
		return nil
	}

	return NewTuneButton(f.client, label, path)
}

func (f *Factory) createSwitchButton(config map[string]interface{}) hamdeck.Button {
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	inputTopic, haveInputTopic := hamdeck.ToString(config[ConfigInputTopic])
	outputTopic, haveOutputTopic := hamdeck.ToString(config[ConfigOutputTopic])
	onPayload, haveOnPayload := hamdeck.ToString(config[ConfigOnPayload])
	offPayload, haveOffPayload := hamdeck.ToString(config[ConfigOffPayload])
	mode, haveMode := hamdeck.ToString(config[ConfigMode])

	if !(haveLabel && haveInputTopic && haveOutputTopic && haveOnPayload && haveOffPayload && haveMode) {
		log.Print("A mqtt.Switch button must have label, inputTopic, outputTopic, onPayload, offPayload, and mode fields.")
		return nil
	}

	return NewSwitchButton(f.client, label, inputTopic, outputTopic, onPayload, offPayload, SwitchMode(strings.TrimSpace(strings.ToUpper(mode))))
}
