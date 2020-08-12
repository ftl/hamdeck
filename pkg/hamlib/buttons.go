package hamlib

import (
	"context"
	"image"
	"log"

	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/rigproxy/pkg/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

/*
	SetModeButton
*/

func NewSetModeButton(hamlibClient *HamlibClient, mode client.Mode, label string) *SetModeButton {
	result := &SetModeButton{
		client:  hamlibClient,
		enabled: hamlibClient.Connected(),
		mode:    mode,
		label:   label,
	}

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

/*
	ToggleModeButton
*/

func NewToggleModeButton(hamlibClient *HamlibClient, mode1 client.Mode, label1 string, mode2 client.Mode, label2 string) *ToggleModeButton {
	result := &ToggleModeButton{
		client:  hamlibClient,
		enabled: hamlibClient.Connected(),
		modes:   []client.Mode{mode1, mode2},
		labels:  []string{label1, label2},
	}

	hamlibClient.Listen(result)

	return result
}

type ToggleModeButton struct {
	hamdeck.BaseButton
	client        *HamlibClient
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	modes         []client.Mode
	labels        []string
	currentMode   int
}

func (b *ToggleModeButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.image = nil
	b.selectedImage = nil
	b.Invalidate()
}

func (b *ToggleModeButton) SetMode(mode client.Mode) {
	wasSelected := b.selected
	lastMode := b.currentMode

	b.selected = false
	for i, m := range b.modes {
		if mode == m {
			b.currentMode = i
			b.selected = true
			break
		}
	}

	if (b.selected == wasSelected) && (b.currentMode == lastMode) {
		return
	}

	b.image = nil
	b.selectedImage = nil
	b.Invalidate()
}

func (b *ToggleModeButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	text := make([]string, 2)
	for i := range text {
		text[i] = string(b.modes[i])
		if b.labels[i] != "" {
			text[i] = b.labels[i]
		}
	}
	if b.image == nil {
		b.image = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.currentMode+1)
	}
	if b.selectedImage == nil {
		gc.SwapColors()
		b.selectedImage = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.currentMode+1)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *ToggleModeButton) Pressed() {
	if !b.enabled {
		return
	}
	mode := b.currentMode
	if b.selected {
		mode = (mode + 1) % len(b.modes)
	}
	ctx := context.Background()
	err := b.client.Conn.SetModeAndPassband(ctx, b.modes[mode], 0)
	if err != nil {
		log.Printf("cannot set mode: %v", err)
	}
}

func (b *ToggleModeButton) Released() {
	// ignore
}

/*
	SetButton
*/

func NewSetButton(hamlibClient *HamlibClient, label string, command string, args ...string) *SetButton {
	result := &SetButton{
		client:  hamlibClient,
		enabled: hamlibClient.Connected(),
		label:   label,
		command: command,
		args:    args,
	}

	hamlibClient.Listen(result)

	return result
}

type SetButton struct {
	hamdeck.BaseButton
	client  *HamlibClient
	image   image.Image
	enabled bool
	label   string
	command string
	args    []string
}

func (b *SetButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.image = nil
	b.Invalidate()
}

func (b *SetButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	if b.image == nil {
		b.image = gc.DrawSingleLineTextButton(b.label)
	}
	return b.image
}

func (b *SetButton) Pressed() {
	if !b.enabled {
		return
	}
	ctx := context.Background()
	err := b.client.Conn.Set(ctx, b.command, b.args...)
	if err != nil {
		log.Printf("cannot execute %s: %v", b.command, err)
	}
}

func (b *SetButton) Released() {
	// ignore
}

/*
	SwitchToBandButton
*/

func NewSwitchToBandButton(hamlibClient *HamlibClient, label string, bandName string) *SwitchToBandButton {
	band, ok := bandplan.IARURegion1[bandplan.BandName(bandName)]
	if !ok {
		log.Printf("cannot find band %s in IARU Region 1 bandplan", bandName)
		return nil
	}
	result := &SwitchToBandButton{
		client:  hamlibClient,
		enabled: hamlibClient.Connected(),
		label:   label,
		band:    band,
	}

	hamlibClient.Listen(result)

	return result
}

type SwitchToBandButton struct {
	hamdeck.BaseButton
	client        *HamlibClient
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	label         string
	band          bandplan.Band
}

func (b *SwitchToBandButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.image = nil
	b.selectedImage = nil
	b.Invalidate()
}

func (b *SwitchToBandButton) SetFrequency(frequency client.Frequency) {
	wasSelected := b.selected
	b.selected = b.band.Contains(frequency)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate()
}

func (b *SwitchToBandButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	text := string(b.band.Name)
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

func (b *SwitchToBandButton) Pressed() {
	if !b.enabled {
		return
	}
	ctx := context.Background()
	err := b.client.Conn.SwitchToBand(ctx, b.band)
	if err != nil {
		log.Print(err)
	}
}

func (b *SwitchToBandButton) Released() {
	// ignore
}

/*
	SetPowerLevelButton
*/

func NewSetPowerLevelButton(hamlibClient *HamlibClient, label string, value float64) *SetPowerLevelButton {
	result := &SetPowerLevelButton{
		client:  hamlibClient,
		enabled: hamlibClient.Connected(),
		label:   label,
		value:   value,
	}

	hamlibClient.Listen(result)

	return result
}

type SetPowerLevelButton struct {
	hamdeck.BaseButton
	client        *HamlibClient
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	label         string
	value         float64
}

func (b *SetPowerLevelButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.image = nil
	b.selectedImage = nil
	b.Invalidate()
}

func (b *SetPowerLevelButton) SetPowerLevel(powerLevel float64) {
	wasSelected := b.selected
	b.selected = (powerLevel == b.value)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate()
}

func (b *SetPowerLevelButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	if b.image == nil {
		b.image = gc.DrawSingleLineTextButton(b.label)
	}
	if b.selectedImage == nil {
		gc.SwapColors()
		b.selectedImage = gc.DrawSingleLineTextButton(b.label)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *SetPowerLevelButton) Pressed() {
	if !b.enabled {
		return
	}
	ctx := context.Background()
	err := b.client.Conn.SetPowerLevel(ctx, b.value)
	if err != nil {
		log.Print(err)
	}
}

func (b *SetPowerLevelButton) Released() {
	// ignore
}

/*
	MOXButton
*/

func NewMOXButton(hamlibClient *HamlibClient, label string) *MOXButton {
	if label == "" {
		label = "TX"
	}

	result := &MOXButton{
		client:  hamlibClient,
		enabled: hamlibClient.Connected(),
		label:   label,
	}

	hamlibClient.Listen(result)

	return result
}

type MOXButton struct {
	hamdeck.BaseButton
	client        *HamlibClient
	image         image.Image
	selectedImage image.Image
	flashImage    image.Image
	enabled       bool
	selected      bool
	flashOn       bool
	label         string
}

func (b *MOXButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.image = nil
	b.selectedImage = nil
	b.Invalidate()
}

func (b *MOXButton) Flash(flashOn bool) {
	if !b.selected {
		return
	}

	b.flashOn = flashOn
	b.Invalidate()
}

func (b *MOXButton) SetPTT(ptt client.PTT) {
	wasSelected := b.selected
	b.selected = (ptt != client.PTTRx)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate()
}

func (b *MOXButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	if b.image == nil {
		b.image = gc.DrawSingleLineTextButton(b.label)
	}
	if b.flashImage == nil {
		gc.SetBackground(hamdeck.Red)
		b.flashImage = gc.DrawSingleLineTextButton(b.label)
	}
	if b.selectedImage == nil {
		gc.SwapColors()
		b.selectedImage = gc.DrawSingleLineTextButton(b.label)
	}
	if !b.selected {
		return b.image
	}
	if b.flashOn {
		return b.flashImage
	}
	return b.selectedImage
}

func (b *MOXButton) Pressed() {
	if !b.enabled {
		return
	}
	value := client.PTTTx
	if b.selected {
		value = client.PTTRx
	}
	ctx := context.Background()
	err := b.client.Conn.SetPTT(ctx, value)
	if err != nil {
		log.Print(err)
	}
}

func (b *MOXButton) Released() {
	// ignore
}
