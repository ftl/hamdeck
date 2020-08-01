package hamlib

import (
	"context"
	"image"
	"log"

	"github.com/ftl/hamdeck/pkg/hamdeck"

	"github.com/ftl/rigproxy/pkg/client"
)

func NewSetModeButton(hamlibClient *HamlibClient, mode client.Mode, label string) *SetModeButton {
	result := &SetModeButton{
		client:  hamlibClient,
		enabled: true,
		mode:    mode,
		label:   label,
	}

	result.updateSelection()
	hamlibClient.Listen(result)

	return result
}

type SetModeButton struct {
	hamdeck.BaseButton
	client        *HamlibClient
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	mode          client.Mode
	label         string
}

func (b *SetModeButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.image = nil
	b.selectedImage = nil
	b.Invalidate()
}

func (b *SetModeButton) updateSelection() {
	mode, _, err := b.client.Conn.ModeAndPassband(context.Background())
	if err != nil {
		log.Printf("cannot retrieve current mode: %v", err)
		return
	}
	b.SetMode(mode)
}

func (b *SetModeButton) SetMode(mode client.Mode) {
	wasSelected := b.selected
	b.selected = (mode == b.mode)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate()
}

func (b *SetModeButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	text := string(b.mode)
	if b.label != "" {
		text = b.label
	}
	if b.image == nil {
		b.image = gc.DrawSingleLineTextButton(text)
	}
	if b.selectedImage == nil {
		gc.SwapColors()
		b.selectedImage = gc.DrawSingleLineTextButton(text)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *SetModeButton) Pressed() {
	if !b.enabled {
		return
	}
	ctx := context.Background()
	err := b.client.Conn.SetModeAndPassband(ctx, b.mode, 0)
	if err != nil {
		log.Printf("cannot set mode: %v", err)
	}
}

func (b *SetModeButton) Released() {
	// ignore
}
