package tci

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/ftl/hamdeck/pkg/hamdeck"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/tci/client"
)

const (
	ConfigCommand         = "command"
	ConfigArgs            = "args"
	ConfigMode            = "mode"
	ConfigLabel           = "label"
	ConfigMode1           = "mode1"
	ConfigLabel1          = "label1"
	ConfigMode2           = "mode2"
	ConfigLabel2          = "label2"
	ConfigIcon            = "icon"
	ConfigBand            = "band"
	ConfigValue           = "value"
	ConfigIncrement       = "increment"
	ConfigBottomFrequency = "bottom_frequency"
	ConfigTopFrequency    = "top_frequency"
)

const (
	SetModeButtonType         = "tci.SetMode"
	ToggleModeButtonType      = "tci.ToggleMode"
	SetFilterButtonType       = "tci.SetFilter"
	MOXButtonType             = "tci.MOX"
	TuneButtonType            = "tci.Tune"
	MuteButtonType            = "tci.Mute"
	SetDriveButtonType        = "tci.SetDrive"
	IncrementDriveButtonType  = "tci.IncrementDrive"
	IncrementVolumeButtonType = "tci.IncrementVolume"
	SwitchToBandButtonType    = "tci.SwitchToBand"
)

func NewButtonFactory(address string) *Factory {
	host, err := parseTCPAddr(address)
	if err != nil {
		return &Factory{}
	}

	return &Factory{client: NewClient(host)}
}

type Factory struct {
	client *Client
}

func (f *Factory) Close() {
	f.client.Disconnect()
}

func (f *Factory) CreateButton(config map[string]interface{}) hamdeck.Button {
	if f.client == nil {
		return nil
	}

	switch config[hamdeck.ConfigType] {
	case SetModeButtonType:
		return f.createSetModeButton(config)
	case ToggleModeButtonType:
		return f.createToggleModeButton(config)
	case SetFilterButtonType:
		return f.createSetFilterButton(config)
	case MOXButtonType:
		return f.createMOXButton(config)
	case TuneButtonType:
		return f.createTuneButton(config)
	case MuteButtonType:
		return f.createMuteButton(config)
	case SetDriveButtonType:
		return f.createSetDriveButton(config)
	case IncrementDriveButtonType:
		return f.createIncrementDriveButton(config)
	case IncrementVolumeButtonType:
		return f.createIncrementVolumeButton(config)
	case SwitchToBandButtonType:
		return f.createSwitchToBandButton(config)
	default:
		return nil
	}
}

func (f *Factory) createSetModeButton(config map[string]interface{}) hamdeck.Button {
	mode, haveMode := hamdeck.ToString(config[ConfigMode])
	label, _ := hamdeck.ToString(config[ConfigLabel])

	mode = strings.ToLower(strings.TrimSpace(mode))

	if !haveMode {
		log.Print("A tci.SetMode button must have a mode field.")
		return nil
	}

	return NewSetModeButton(f.client, client.Mode(mode), label)
}

func (f *Factory) createToggleModeButton(config map[string]interface{}) hamdeck.Button {
	mode1, haveMode1 := hamdeck.ToString(config[ConfigMode1])
	label1, _ := hamdeck.ToString(config[ConfigLabel1])
	mode2, haveMode2 := hamdeck.ToString(config[ConfigMode2])
	label2, _ := hamdeck.ToString(config[ConfigLabel2])

	mode1 = strings.ToLower(strings.TrimSpace(mode1))
	mode2 = strings.ToLower(strings.TrimSpace(mode2))

	if !(haveMode1 && haveMode2) {
		log.Print("A tci.ToggleMode button must have mode1 and mode2 fields.")
		return nil
	}

	return NewToggleModeButton(f.client, client.Mode(strings.ToLower(mode1)), label1, client.Mode(strings.ToLower(mode2)), label2)
}

func (f *Factory) createSetFilterButton(config map[string]interface{}) hamdeck.Button {
	bottomFrequency, haveBottomFrequency := hamdeck.ToInt(config[ConfigBottomFrequency])
	topFrequency, haveTopFrequency := hamdeck.ToInt(config[ConfigTopFrequency])
	mode, _ := hamdeck.ToString(config[ConfigMode])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	icon, haveIcon := hamdeck.ToString(config[ConfigIcon])
	mode = strings.ToLower(strings.TrimSpace(mode))
	if !haveBottomFrequency {
		log.Print("A tci.SetFilter button must have a bottom_frequency field.")
		return nil
	}
	if !haveTopFrequency {
		log.Print("A tci.SetFilter button must have a top_frequency field.")
		return nil
	}
	if !haveLabel {
		log.Print("A tci.SetFilter button must have a label field.")
		return nil
	}
	if !haveIcon {
		icon = "filter"
	}
	return NewSetFilterButton(f.client, bottomFrequency, topFrequency, client.Mode(mode), label, icon)
}

func (f *Factory) createMOXButton(config map[string]interface{}) hamdeck.Button {
	label, _ := hamdeck.ToString(config[ConfigLabel])

	return NewMOXButton(f.client, label)
}

func (f *Factory) createTuneButton(config map[string]interface{}) hamdeck.Button {
	label, _ := hamdeck.ToString(config[ConfigLabel])

	return NewTuneButton(f.client, label)
}

func (f *Factory) createMuteButton(config map[string]interface{}) hamdeck.Button {
	label, _ := hamdeck.ToString(config[ConfigLabel])

	return NewMuteButton(f.client, label)
}

func (f *Factory) createSetDriveButton(config map[string]interface{}) hamdeck.Button {
	value, haveValue := hamdeck.ToInt(config[ConfigValue])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	if !(haveValue && haveLabel) {
		log.Print("A tci.SetDrive button must have value and label fields.")
		return nil
	}

	return NewSetDriveButton(f.client, label, value)
}

func (f *Factory) createIncrementDriveButton(config map[string]interface{}) hamdeck.Button {
	increment, haveIncrement := hamdeck.ToInt(config[ConfigIncrement])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	if !(haveIncrement && haveLabel) {
		log.Print("A tci.IncrementDrive button must have increment and label fields.")
		return nil
	}

	return NewIncrementDriveButton(f.client, label, increment)
}

func (f *Factory) createIncrementVolumeButton(config map[string]interface{}) hamdeck.Button {
	increment, haveIncrement := hamdeck.ToInt(config[ConfigIncrement])
	label, haveLabel := hamdeck.ToString(config[ConfigLabel])
	if !(haveIncrement && haveLabel) {
		log.Print("A tci.IncrementVolume button must have increment and label fields.")
		return nil
	}

	return NewIncrementVolumeButton(f.client, label, increment)
}

func (f *Factory) createSwitchToBandButton(config map[string]interface{}) hamdeck.Button {
	band, haveBand := hamdeck.ToString(config[ConfigBand])
	label, _ := hamdeck.ToString(config[ConfigLabel])
	if !(haveBand) {
		log.Print("A tci.SwitchToBand button must have a band field.")
		return nil
	}

	return NewSwitchToBandButton(f.client, label, band)
}

func parseTCPAddr(arg string) (*net.TCPAddr, error) {
	host, port := splitHostPort(arg)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = strconv.Itoa(client.DefaultPort)
	}

	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, port))
}

func splitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func toBandplanMode(m client.Mode) bandplan.Mode {
	switch m {
	case client.ModeCW:
		return bandplan.ModeCW
	case client.ModeUSB, client.ModeLSB, client.ModeAM, client.ModeNFM, client.ModeWFM, client.ModeSAM, client.ModeDSB, client.ModeDRM:
		return bandplan.ModePhone
	case client.ModeDIGL, client.ModeDIGU, client.ModeSPEC:
		return bandplan.ModeDigital
	default:
		return bandplan.ModeDigital
	}
}
