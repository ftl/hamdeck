package hamdeck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig_Connections(t *testing.T) {
	runWithConfigString(t, `{
	"connections": {
		"testsdrA": {
			"type": "test",
			"some_config": "some_value"
		},
		"testsdrB": {
			"type": "test",
			"some_config": "some_other_value"
		}
	}
}`, func(t *testing.T, deck *HamDeck, device *testDevice, _ chan struct{}) {
		assert.Equal(t, 2, len(deck.connections))

		configA, ok := deck.GetConnection("testsdrA", "test")
		assert.True(t, ok)
		assert.Equal(t, "some_value", configA["some_config"])

		configB, ok := deck.GetConnection("testsdrB", "test")
		assert.True(t, ok)
		assert.Equal(t, "some_other_value", configB["some_config"])

		_, ok = deck.GetConnection("testsdrA", "undefined")
		assert.False(t, ok)
	})
}

func TestConnectionManager(t *testing.T) {
	provider := &testConnectionProvider{
		name: "blah",
		config: ConnectionConfig{
			"type":        "test",
			"some_config": "some_value",
		},
	}
	manager := NewConnectionManager[*testConnection]("test", provider, provider.CreateConnection)

	connection, err := manager.Get("blah")
	assert.NoError(t, err)
	assert.Equal(t, "some_value", connection.config["some_config"])

	_, err = manager.Get("undefined")
	assert.Error(t, err)
}

type testConnection struct {
	config ConnectionConfig
}

type testConnectionProvider struct {
	name   string
	config ConnectionConfig
}

func (p *testConnectionProvider) GetConnection(name string, connectionType string) (ConnectionConfig, bool) {
	if name != p.name || connectionType != "test" {
		return ConnectionConfig{}, false
	}
	return p.config, true
}

func (p *testConnectionProvider) CreateConnection(config ConnectionConfig) (*testConnection, error) {
	return &testConnection{
		config: config,
	}, nil
}
