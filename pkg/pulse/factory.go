package pulse

import "github.com/ftl/hamdeck/pkg/hamdeck"

const (
	ConfigSinkID   = "sink"
	ConfigSourceID = "source"
	ConfigLabel    = "label"
)

const (
	ToggleMuteButtonType = "pulse.ToggleMute"
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
	label, _ := hamdeck.ToString(config[ConfigLabel])
	if !(haveSinkID || haveSourceID) {
		return nil
	}

	return NewToggleMuteButton(f.client, sinkID, sourceID, label)
}
