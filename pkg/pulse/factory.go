package pulse

import "github.com/ftl/hamdeck/pkg/hamdeck"

const (
	ConfigSinkID   = "sink"
	ConfigSourceID = "source"
	ConfigLabel    = "label"
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
	if config[hamdeck.ConfigType] != "pulse.ToggleMute" {
		return nil
	}
	sinkID, haveSinkID := toString(config[ConfigSinkID])
	sourceID, haveSourceID := toString(config[ConfigSourceID])
	label, _ := toString(config[ConfigLabel])
	if !(haveSinkID || haveSourceID) {
		return nil
	}

	return NewToggleMuteButton(f.client, sinkID, sourceID, label)
}

func toString(raw interface{}) (string, bool) {
	if raw == nil {
		return "", false
	}
	s, ok := raw.(string)
	if !ok {
		return "", false
	}
	return s, ok
}
