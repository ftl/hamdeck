package pulse

import "github.com/ftl/hamdeck/pkg/hamdeck"

const (
	ConfigSinkID           = "sink"
	ConfigSourceID         = "source"
	ConfigSinkInputName    = "sinkInput"
	ConfigSourceOutputName = "sourceOutput"
	ConfigLabel            = "label"
)

const (
	ToggleMuteButtonType = "pulse.ToggleMute"
)

func NewButtonFactory() *Factory {
	client := NewClient()
	client.KeepOpen()

	return &Factory{
		client: client,
	}
}

type Factory struct {
	client *PulseClient
}

func (f *Factory) Close() {
	f.client.Close()
}

func (f *Factory) CreateButton(config map[string]interface{}) hamdeck.Button {
	switch config[hamdeck.ConfigType] {
	case ToggleMuteButtonType:
		return f.createToggleMuteButton(config)
	default:
		return nil
	}
}

func (f *Factory) createToggleMuteButton(config map[string]interface{}) hamdeck.Button {
	sinkID, haveSinkID := hamdeck.ToString(config[ConfigSinkID])
	sourceID, haveSourceID := hamdeck.ToString(config[ConfigSourceID])
	sinkInputName, haveSinkInputName := hamdeck.ToString(config[ConfigSinkInputName])
	sourceOutputName, haveSourceOutputName := hamdeck.ToString(config[ConfigSourceOutputName])
	label, _ := hamdeck.ToString(config[ConfigLabel])
	if !(haveSinkID || haveSourceID || haveSinkInputName || haveSourceOutputName) {
		return nil
	}

	return NewToggleMuteButton(f.client, sinkID, sourceID, sinkInputName, sourceOutputName, label)
}
