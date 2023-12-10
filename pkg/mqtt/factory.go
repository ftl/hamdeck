package mqtt

import (
	"fmt"
	"log"
	"strings"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const (
	ConfigAddress     = "address"
	ConfigUsername    = "username"
	ConfigPassword    = "password"
	ConfigLabel       = "label"
	ConfigPath        = "path"
	ConfigInputTopic  = "inputTopic"
	ConfigOutputTopic = "outputTopic"
	ConfigOnPayload   = "onPayload"
	ConfigOffPayload  = "offPayload"
	ConfigMode        = "mode"
)

const (
	ConnectionType   = "mqtt"
	TuneButtonType   = "mqtt.AT100Tune"
	SwitchButtonType = "mqtt.Switch"
)

func NewButtonFactory(provider hamdeck.ConnectionConfigProvider, legacyAddress string, username string, password string) *Factory {
	result := &Factory{}
	result.connections = hamdeck.NewConnectionManager(ConnectionType, provider, result.createMQTTClient)

	if legacyAddress != "" {
		result.connections.SetLegacy(NewClient(legacyAddress, username, password))
	}

	return result
}

type Factory struct {
	connections *hamdeck.ConnectionManager[*Client]
}

func (f *Factory) createMQTTClient(name string, config hamdeck.ConnectionConfig) (*Client, error) {
	address, ok := hamdeck.ToString(config[ConfigAddress])
	if !ok {
		return nil, fmt.Errorf("no address defined for mqtt connection %s", name)
	}
	username, _ := hamdeck.ToString(config[ConfigUsername])
	password, _ := hamdeck.ToString(config[ConfigPassword])

	client := NewClient(address, username, password)

	return client, nil
}

func (f *Factory) Close() {
	f.connections.ForEach(func(client *Client) {
		client.Disconnect()
	})
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

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	mqttClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create mqtt.ATU100Tune button: %v", err)
		return nil
	}

	return NewTuneButton(mqttClient, label, path)
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

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	mqttClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create mqtt.Switch button: %v", err)
		return nil
	}

	return NewSwitchButton(mqttClient, label, inputTopic, outputTopic, onPayload, offPayload, SwitchMode(strings.TrimSpace(strings.ToUpper(mode))))
}
