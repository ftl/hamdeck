package hamlib

import (
	"log"

	"github.com/ftl/rigproxy/pkg/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

const (
	ConfigCommand = "command"
	ConfigArgs    = "args"
	ConfigMode    = "mode"
	ConfigLabel   = "label"
	ConfigMode1   = "mode1"
	ConfigLabel1  = "label1"
	ConfigMode2   = "mode2"
	ConfigLabel2  = "label2"
	ConfigBand    = "band"
	ConfigValue   = "value"
)

const (
	SetModeButtonType       = "hamlib.SetMode"
	ToggleModeButtonType    = "hamlib.ToggleMode"
	SetButtonType           = "hamlib.Set"
	SwitchToBandButtonType  = "hamlib.SwitchToBand"
	SetPowerLevelButtonType = "hamlib.SetPowerLevel"
	MOXButtonType           = "hamlib.MOX"
)

func NewButtonFactory(address string) (*Factory, error) {
	client, err := NewClient(address)
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
	case SetButtonType:
		return f.createSetButton(config)
	case SwitchToBandButtonType:
		return f.createSwitchToBandButton(config)
	case SetPowerLevelButtonType:
		return f.createSetPowerLevelButton(config)
	case MOXButtonType:
		return f.createMOXButton(config)
	default:
		return nil
	}
}

func (f *Factory) createSetModeButton(config map[string]interface{}) hamdeck.Button {
	mode, haveMode := hamdeck.ToString(config[ConfigMode])
	label, _ := hamdeck.ToString(config[ConfigLabel])
	if !haveMode {
		log.Print("A hamlib.SetMode button must have a mode field.")
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
		log.Print("A hamlib.ToggleMode button must have mode1 and mode2 fields.")
		return nil
	}

	return NewToggleModeButton(f.client, client.Mode(mode1), label1, client.Mode(mode2), label2)
}

func (f *Factory) createSetButton(config map[string]interface{}) hamdeck.Button {
	command, haveCommand := hamdeck.ToString(config[ConfigCommand])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	args, _ := hamdeck.ToStringArray(config[ConfigArgs])
	if !(haveCommand && haveLabel) {
		log.Print("A hamlib.Set button must have command and label fields.")
		return nil
	}

	return NewSetButton(f.client, label, command, args...)
}

func (f *Factory) createSwitchToBandButton(config map[string]interface{}) hamdeck.Button {
	band, haveBand := hamdeck.ToString(config[ConfigBand])
	label, _ := hamdeck.ToString(config[ConfigLabel])
	if !(haveBand) {
		log.Print("A hamlib.SwitchToBand button must have a band field.")
		return nil
	}

	return NewSwitchToBandButton(f.client, label, band)
}

func (f *Factory) createSetPowerLevelButton(config map[string]interface{}) hamdeck.Button {
	value, haveValue := hamdeck.ToFloat(config[ConfigValue])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	if !(haveValue && haveLabel) {
		log.Print("A hamlib.SetPowerLevel button must have value and label fields.")
		return nil
	}

	return NewSetPowerLevelButton(f.client, label, value)
}

func (f *Factory) createMOXButton(config map[string]interface{}) hamdeck.Button {
	label, _ := hamdeck.ToString(config[ConfigLabel])

	return NewMOXButton(f.client, label)
}
