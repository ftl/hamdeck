package pulse

import (
	"image"
	"log"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

func NewToggleMuteButton(client *PulseClient, sinkID, sourceID, sinkInputName, sourceOutputName string, label string) *ToggleMuteButton {
	result := &ToggleMuteButton{
		client:           client,
		sinkID:           sinkID,
		sourceID:         sourceID,
		sinkInputName:    sinkInputName,
		sourceOutputName: sourceOutputName,
		label:            label,
		enabled:          client.Connected(),
	}

	result.updateSelection()
	client.Listen(result)

	return result
}

type ToggleMuteButton struct {
	hamdeck.BaseButton
	client           *PulseClient
	sinkID           string
	sourceID         string
	sinkInputName    string
	sourceOutputName string
	label            string
	enabled          bool
	muted            bool
	mutedImage       image.Image
	unmutedImage     image.Image
}

func (b *ToggleMuteButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *ToggleMuteButton) updateSelection() {
	if !b.client.Connected() {
		b.Enable(false)
		return
	}

	var id string
	var muted bool
	var err error
	if b.sinkID != "" {
		id = b.sinkID
		muted, err = b.client.IsSinkMuted(b.sinkID)
	} else if b.sourceID != "" {
		id = b.sourceID
		muted, err = b.client.IsSourceMuted(b.sourceID)
	} else if b.sinkInputName != "" {
		muted, err = b.client.IsSinkInputMuted(b.sinkInputName)
	} else if b.sourceOutputName != "" {
		muted, err = b.client.IsSourceOutputMuted(b.sourceOutputName)
	}
	if err != nil {
		log.Print(err)
		return
	}

	b.SetMute(id, muted)
}

func (b *ToggleMuteButton) SetMute(id string, mute bool) {
	if id == b.sinkID || id == b.sourceID || id == b.sinkInputName || id == b.sourceOutputName {
		b.muted = mute
		b.Invalidate(false)
	}
}

func (b *ToggleMuteButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.mutedImage == nil || b.unmutedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.muted {
		return b.mutedImage
	}
	return b.unmutedImage
}

func (b *ToggleMuteButton) redrawImages(gc hamdeck.GraphicContext) {
	gc.SetFontSize(16)
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	b.mutedImage = gc.DrawIconLabelButton(gc.LoadIconAsset("volume_off.png"), b.label)
	b.unmutedImage = gc.DrawIconLabelButton(gc.LoadIconAsset("volume_up.png"), b.label)
}

func (b *ToggleMuteButton) Pressed() {
	if !b.enabled {
		return
	}

	var err error
	if b.sinkID != "" {
		_, err = b.client.ToggleMuteSink(b.sinkID)
	} else if b.sourceID != "" {
		_, err = b.client.ToggleMuteSource(b.sourceID)
	} else if b.sinkInputName != "" {
		_, err = b.client.ToggleMuteSinkInput(b.sinkInputName)
	} else if b.sourceOutputName != "" {
		_, err = b.client.ToggleMuteSourceOutput(b.sourceOutputName)
	}
	if err != nil {
		log.Printf("cannot toggle mute state: %v", err)
	}
}

func (b *ToggleMuteButton) Released() {
	// ignore
}
