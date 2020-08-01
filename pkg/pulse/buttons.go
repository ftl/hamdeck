package pulse

import (
	"image"
	"log"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

func NewToggleMuteButton(client *PulseClient, sinkID, sourceID string, label string) *ToggleMuteButton {
	var muted bool
	var err error
	if sinkID != "" {
		muted, err = client.IsSinkMuted(sinkID)
		if err != nil {
			log.Print(err)
		}
	} else if sourceID != "" {
		muted, err = client.IsSourceMuted(sourceID)
		if err != nil {
			log.Print(err)
		}
	}

	result := &ToggleMuteButton{
		client:   client,
		sinkID:   sinkID,
		sourceID: sourceID,
		label:    label,
		enabled:  true,
		muted:    muted,
	}
	client.Listen(result)

	return result
}

type ToggleMuteButton struct {
	hamdeck.BaseButton
	client       *PulseClient
	sinkID       string
	sourceID     string
	label        string
	enabled      bool
	muted        bool
	mutedImage   image.Image
	unmutedImage image.Image
}

func (b *ToggleMuteButton) SetMute(id string, mute bool) {
	if id == b.sinkID || id == b.sourceID {
		b.muted = mute
		b.Invalidate()
	}
}

func (b *ToggleMuteButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.mutedImage = nil
	b.unmutedImage = nil
	b.Invalidate()
}

func (b *ToggleMuteButton) Image(gc hamdeck.GraphicContext) image.Image {
	gc.SetFontSize(16)
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	if b.mutedImage == nil {
		b.mutedImage = gc.DrawIconLabelButton(gc.LoadIconAsset("volume_off.png"), b.label)
	}
	if b.unmutedImage == nil {
		b.unmutedImage = gc.DrawIconLabelButton(gc.LoadIconAsset("volume_up.png"), b.label)
	}

	if b.muted {
		return b.mutedImage
	}
	return b.unmutedImage
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
	}
	if err != nil {
		log.Printf("cannot toggle mute state: %v", err)
	}
}

func (b *ToggleMuteButton) Released() {
	// ignore
}
