package hamdeck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfig_SinglePage(t *testing.T) {
	runWithConfigString(t, `{
	"start_page": "main",
	"pages": {
		"main": {
			"buttons": [
				{ "type": "test.Button", "index": 0, "some_config": "some_value" }
			]
		}
	}
}`, func(t *testing.T, deck *HamDeck, device *testDevice, _ chan struct{}) {
		assert.Equal(t, "main", deck.startPageID)
		assert.Equal(t, 1, len(deck.pages))

		page := deck.pages["main"]
		require.Equal(t, len(deck.buttons), len(page.buttons))
		button, ok := page.buttons[0].(*testButton)
		require.True(t, ok)
		assert.Equal(t, "some_value", button.config["some_config"])

		assert.Same(t, button, deck.buttons[0])
		assert.True(t, button.attached)
	})
}

func TestReadConfig_SinglePagePlusLegacy(t *testing.T) {
	runWithConfigString(t, `{
	"pages": {
		"main": {
			"buttons": [
				{ "type": "test.Button", "index": 0, "some_config": "some_value" }
			]
		}
	},
	"buttons": [
		{ "type": "test.Button", "index": 1, "legacy_config": "legacy_value" }
	]
}`, func(t *testing.T, deck *HamDeck, device *testDevice, _ chan struct{}) {
		assert.Equal(t, legacyPageID, deck.startPageID)
		assert.Equal(t, 2, len(deck.pages))

		mainPage := deck.pages["main"]
		require.Equal(t, len(deck.buttons), len(mainPage.buttons))
		button, ok := mainPage.buttons[0].(*testButton)
		require.True(t, ok)
		assert.Equal(t, "some_value", button.config["some_config"])

		legacyPage := deck.pages[legacyPageID]
		require.Equal(t, len(deck.buttons), len(legacyPage.buttons))
		button, ok = legacyPage.buttons[1].(*testButton)
		require.True(t, ok)
		assert.Equal(t, "legacy_value", button.config["legacy_config"])

		assert.Same(t, button, deck.buttons[1])
		assert.True(t, button.attached)
	})
}

func TestReadConfig_OnlyLegacy(t *testing.T) {
	runWithConfigString(t, `{
	"buttons": [
		{ "type": "test.Button", "index": 1, "legacy_config": "legacy_value" }
	]
}`, func(t *testing.T, deck *HamDeck, device *testDevice, _ chan struct{}) {
		assert.Equal(t, legacyPageID, deck.startPageID)
		assert.Equal(t, 1, len(deck.pages))

		legacyPage := deck.pages[legacyPageID]
		require.Equal(t, len(deck.buttons), len(legacyPage.buttons))
		button, ok := legacyPage.buttons[1].(*testButton)
		require.True(t, ok)
		assert.Equal(t, "legacy_value", button.config["legacy_config"])

		assert.Same(t, button, deck.buttons[1])
		assert.True(t, button.attached)
	})
}

func TestAttachPage(t *testing.T) {
	runWithConfigString(t, `{
	"pages": {
		"main": {
			"buttons": [
				{ "type": "test.Button", "index": 0, "some_config": "some_value" }
			]
		}
	},
	"buttons": [
		{ "type": "test.Button", "index": 1, "legacy_config": "legacy_value" }
	]
}`, func(t *testing.T, deck *HamDeck, device *testDevice, _ chan struct{}) {
		assert.Equal(t, legacyPageID, deck.startPageID)
		assert.Equal(t, 2, len(deck.pages))

		mainPage := deck.pages["main"]
		require.Equal(t, len(deck.buttons), len(mainPage.buttons))
		mainButton, ok := mainPage.buttons[0].(*testButton)
		require.True(t, ok)
		assert.Equal(t, "some_value", mainButton.config["some_config"])

		legacyPage := deck.pages[legacyPageID]
		require.Equal(t, len(deck.buttons), len(legacyPage.buttons))
		legacyButton, ok := legacyPage.buttons[1].(*testButton)
		require.True(t, ok)
		assert.Equal(t, "legacy_value", legacyButton.config["legacy_config"])

		assert.Same(t, legacyButton, deck.buttons[1])
		assert.True(t, legacyButton.attached)

		err := deck.AttachPage("main")
		assert.NoError(t, err)

		assert.True(t, legacyButton.detached)
		assert.True(t, mainButton.attached)

		assert.Same(t, mainButton, deck.buttons[0])
		assert.Same(t, deck.noButton, deck.buttons[1])
	})
}

func TestPageButton(t *testing.T) {
	runWithConfigString(t, `{
	"pages": {
		"main": {
			"buttons": [
				{ "type": "hamdeck.Page", "index": 0, "page": "", "label": "Back" }
			]
		}
	},
	"buttons": [
		{ "type": "hamdeck.Page", "index": 0, "page": "main", "label": "Main" }
	]
}`, func(t *testing.T, deck *HamDeck, device *testDevice, _ chan struct{}) {
		assert.Equal(t, legacyPageID, deck.startPageID)
		assert.Equal(t, 2, len(deck.pages))

		mainPage := deck.pages["main"]
		require.Equal(t, len(deck.buttons), len(mainPage.buttons))
		mainButton, ok := mainPage.buttons[0].(*PageButton)
		require.True(t, ok)
		assert.Equal(t, "Back", mainButton.label)

		legacyPage := deck.pages[legacyPageID]
		require.Equal(t, len(deck.buttons), len(legacyPage.buttons))
		legacyButton, ok := legacyPage.buttons[0].(*PageButton)
		require.True(t, ok)
		assert.Equal(t, "Main", legacyButton.label)

		assert.Same(t, legacyButton, deck.buttons[0])

		device.Press(0)
		device.WaitForLastKey()
		assert.Same(t, mainButton, deck.buttons[0])

		device.Press(0)
		device.WaitForLastKey()
		assert.Same(t, legacyButton, deck.buttons[0])
	})
}
