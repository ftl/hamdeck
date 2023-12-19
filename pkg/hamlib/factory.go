package hamlib

import (
	"fmt"
	"log"

	"github.com/ftl/rigproxy/pkg/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const (
	ConfigAddress   = "address"
	ConfigCommand   = "command"
	ConfigArgs      = "args"
	ConfigMode      = "mode"
	ConfigLabel     = "label"
	ConfigMode1     = "mode1"
	ConfigLabel1    = "label1"
	ConfigMode2     = "mode2"
	ConfigLabel2    = "label2"
	ConfigIcon      = "icon"
	ConfigBandwidth = "bandwidth"
	ConfigBand      = "band"
	ConfigValue     = "value"
	ConfigVFO       = "vfo"
	ConfigUseUpDown = "use_up_down"
)

const (
	ConnectionType          = "hamlib"
	SetModeButtonType       = "hamlib.SetMode"
	ToggleModeButtonType    = "hamlib.ToggleMode"
	SetButtonType           = "hamlib.Set"
	SwitchToBandButtonType  = "hamlib.SwitchToBand"
	SetPowerLevelButtonType = "hamlib.SetPowerLevel"
	MOXButtonType           = "hamlib.MOX"
	SetVFOButtonType        = "hamlib.SetVFO"
)

func NewButtonFactory(provider hamdeck.ConnectionConfigProvider, legacyAddress string) *Factory {
	result := &Factory{}
	result.connections = hamdeck.NewConnectionManager(ConnectionType, provider, result.createHamlibClient)

	if legacyAddress != "" {
		client := NewClient(legacyAddress)
		client.KeepOpen()
		result.connections.SetLegacy(client)
	}

	return result
}

type Factory struct {
	connections *hamdeck.ConnectionManager[*HamlibClient]
}

func (f *Factory) createHamlibClient(name string, config hamdeck.ConnectionConfig) (*HamlibClient, error) {
	address, ok := hamdeck.ToString(config[ConfigAddress])
	if !ok {
		return nil, fmt.Errorf("no address defined for hamlib connection %s", name)
	}

	client := NewClient(address)
	client.KeepOpen()

	return client, nil
}

func (f *Factory) Close() {
	f.connections.ForEach(func(client *HamlibClient) {
		client.Close()
	})
}

func (f *Factory) CreateButton(config map[string]interface{}) hamdeck.Button {
	switch config[hamdeck.ConfigType] {
	case SetModeButtonType:
		return f.createSetModeButton(config)
	case ToggleModeButtonType:
		return f.createToggleModeButton(config)
	case SetButtonType:
		return f.createSetButton(config)
	case SwitchToBandButtonType:
		return f.createSwitchToBandButton(config)
	case SetPowerLevelButtonType:
		return f.createSetPowerLevelButton(config)
	case MOXButtonType:
		return f.createMOXButton(config)
	case SetVFOButtonType:
		return f.createSetVFOButton(config)
	default:
		return nil
	}
}

func (f *Factory) createSetModeButton(config map[string]interface{}) hamdeck.Button {
	mode, haveMode := hamdeck.ToString(config[ConfigMode])
	bandwidth, haveBandwidth := hamdeck.ToInt(config[ConfigBandwidth])
	label, _ := hamdeck.ToString(config[ConfigLabel])
	icon, haveIcon := hamdeck.ToString(config[ConfigIcon])
	if !haveMode {
		log.Print("A hamlib.SetMode button must have a mode field.")
		return nil
	}
	if !haveBandwidth {
		bandwidth = 0
	}
	if !haveIcon {
		icon = ""
	}

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	hamlibClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create hamlib.SetMode button: %v", err)
		return nil
	}

	return NewSetModeButton(hamlibClient, client.Mode(mode), client.Frequency(bandwidth), label, icon)
}

func (f *Factory) createToggleModeButton(config map[string]interface{}) hamdeck.Button {
	mode1, haveMode1 := hamdeck.ToString(config[ConfigMode1])
	label1, _ := hamdeck.ToString(config[ConfigLabel1])
	mode2, haveMode2 := hamdeck.ToString(config[ConfigMode2])
	label2, _ := hamdeck.ToString(config[ConfigLabel])
	if !(haveMode1 && haveMode2) {
		log.Print("A hamlib.ToggleMode button must have mode1 and mode2 fields.")
		return nil
	}

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	hamlibClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create hamlib.ToggleMode button: %v", err)
		return nil
	}

	return NewToggleModeButton(hamlibClient, client.Mode(mode1), label1, client.Mode(mode2), label2)
}

func (f *Factory) createSetButton(config map[string]interface{}) hamdeck.Button {
	command, haveCommand := hamdeck.ToString(config[ConfigCommand])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	args, _ := hamdeck.ToStringArray(config[ConfigArgs])
	if !(haveCommand && haveLabel) {
		log.Print("A hamlib.Set button must have command and label fields.")
		return nil
	}

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	hamlibClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create hamlib.Set button: %v", err)
		return nil
	}

	return NewSetButton(hamlibClient, label, command, args...)
}

func (f *Factory) createSwitchToBandButton(config map[string]interface{}) hamdeck.Button {
	band, haveBand := hamdeck.ToString(config[ConfigBand])
	label, _ := hamdeck.ToString(config[ConfigLabel])
	useUpDown, _ := hamdeck.ToBool(config[ConfigUseUpDown])
	if !(haveBand) {
		log.Print("A hamlib.SwitchToBand button must have a band field.")
		return nil
	}

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	hamlibClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create hamlib.SwitchToBand button: %v", err)
		return nil
	}

	return NewSwitchToBandButton(hamlibClient, label, band, useUpDown)
}

func (f *Factory) createSetPowerLevelButton(config map[string]interface{}) hamdeck.Button {
	value, haveValue := hamdeck.ToFloat(config[ConfigValue])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	if !(haveValue && haveLabel) {
		log.Print("A hamlib.SetPowerLevel button must have value and label fields.")
		return nil
	}

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	hamlibClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create hamlib.SetPowerLevel button: %v", err)
		return nil
	}

	return NewSetPowerLevelButton(hamlibClient, label, value)
}

func (f *Factory) createMOXButton(config map[string]interface{}) hamdeck.Button {
	label, _ := hamdeck.ToString(config[ConfigLabel])

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	hamlibClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create hamlib.MOX button: %v", err)
		return nil
	}

	return NewMOXButton(hamlibClient, label)
}

func (f *Factory) createSetVFOButton(config map[string]interface{}) hamdeck.Button {
	vfo, haveVFO := hamdeck.ToString(config[ConfigVFO])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	if !(haveVFO && haveLabel) {
		log.Print("A hamlib.SetVFO button must have vfo and label fields.")
		return nil
	}

	connection, _ := hamdeck.ToString(config[hamdeck.ConfigConnection])
	hamlibClient, err := f.connections.Get(connection)
	if err != nil {
		log.Printf("Cannot create hamlib.SetVFO button: %v", err)
		return nil
	}

	return NewSetVFOButton(hamlibClient, label, client.VFO(vfo))
}
