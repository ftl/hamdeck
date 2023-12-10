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

		configA, ok := deck.GetConnection("testsdrA")
		assert.True(t, ok)
		assert.Equal(t, "some_value", configA["some_config"])

		configB, ok := deck.GetConnection("testsdrB")
		assert.True(t, ok)
		assert.Equal(t, "some_other_value", configB["some_config"])

		_, ok = deck.GetConnection("undefined")
		assert.False(t, ok)
	})
}
