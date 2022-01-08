package mqtt

import (
	"fmt"
	"image"
	"image/color"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

/*
	TuneButton
*/

func NewTuneButton(atu100Client *Client, label string, path string) *TuneButton {
	if label == "" {
		label = "Tune"
	}

	result := &TuneButton{
		client:  atu100Client,
		enabled: atu100Client.Connected(),
		label:   label,
		path:    path,
	}

	atu100Client.Notify(result)
	atu100Client.AddPath(path)

	return result
}

type TuneButton struct {
	hamdeck.BaseButton
	client      *Client
	offImage    image.Image
	tuningImage image.Image
	txImage     image.Image
	enabled     bool
	label       string
	path        string
	alive       bool
	tx          bool
	tuning      bool
	swr         float64
}

func (b *TuneButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *TuneButton) SetAlive(path string, alive bool) {
	if path != b.path {
		return
	}

	wasAlive := b.alive
	b.alive = alive
	if b.alive == wasAlive {
		return
	}
	b.Invalidate(true)
}

func (b *TuneButton) SetTX(path string, tx bool) {
	if path != b.path {
		return
	}

	wasTX := b.tx
	b.tx = tx
	if b.tx == wasTX {
		return
	}
	b.Invalidate(false)
}

func (b *TuneButton) SetTune(path string, tuning bool) {
	if path != b.path {
		return
	}

	wasTuning := b.tuning
	b.tuning = tuning
	if b.tuning == wasTuning {
		return
	}
	b.Invalidate(false)
}

func (b *TuneButton) SetSWR(path string, swr float64) {
	if path != b.path {
		return
	}

	lastSWR := b.swr
	b.swr = swr
	if b.swr == lastSWR {
		return
	}
	b.Invalidate(true)
}

func (b *TuneButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.offImage == nil || b.tuningImage == nil || b.txImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	switch {
	case b.tx:
		return b.txImage
	case b.tuning:
		return b.tuningImage
	default:
		return b.offImage
	}
}

func (b *TuneButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled && b.alive {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	b.offImage = gc.DrawDoubleLineToggleTextButton("Tune", b.label, 1)
	gc.SwapColors()
	b.tuningImage = gc.DrawDoubleLineToggleTextButton("Tune", b.label, 1)

	swr := fmt.Sprintf("%3.2f", b.swr)
	var background color.Color
	switch {
	case b.swr > 3.0:
		background = hamdeck.Red
	case b.swr > 1.5:
		background = hamdeck.Orange
	default:
		background = hamdeck.DarkGreen
	}
	gc.SetBackground(background)
	b.txImage = gc.DrawDoubleLineToggleTextButton(swr, b.label, 1)
}

func (b *TuneButton) Pressed() {
	if !(b.enabled && b.alive) {
		return
	}
	b.client.Tune(b.path)
}

func (b *TuneButton) Released() {
	// ignore
}
