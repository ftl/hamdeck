package hamlib

import (
	"image"
	"log"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/rigproxy/pkg/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

/*
	SetModeButton
*/

func NewSetModeButton(hamlibClient *HamlibClient, mode client.Mode, bandwidth client.Frequency, label string, icon string) *SetModeButton {
	result := &SetModeButton{
		client:       hamlibClient,
		enabled:      hamlibClient.Connected(),
		mode:         mode,
		bandwidth:    bandwidth,
		bandplanMode: mode.ToBandplanMode(),
		label:        label,
		icon:         icon,
	}
	result.longpress = hamdeck.NewLongpressHandler(result.OnLongpress)

	hamlibClient.Listen(result)

	return result
}

type SetModeButton struct {
	hamdeck.BaseButton
	client             *HamlibClient
	iconImage          image.Image
	image              image.Image
	selectedImage      image.Image
	inModePortionImage image.Image
	enabled            bool
	selected           bool
	inModePortion      bool
	mode               client.Mode
	bandplanMode       bandplan.Mode
	bandwidth          client.Frequency
	label              string
	icon               string
	currentMode        client.Mode
	currentBandwidth   client.Frequency
	currentFrequency   client.Frequency
	longpress          *hamdeck.LongpressHandler
}

func (b *SetModeButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *SetModeButton) SetMode(mode client.Mode) {
	b.currentMode = mode
	wasSelected := b.selected
	b.selected = (mode == b.mode) && (b.currentBandwidth == b.bandwidth)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SetModeButton) SetPassband(passband client.Frequency) {
	if b.bandwidth == 0 {
		b.currentBandwidth = 0
		return
	}
	b.currentBandwidth = passband
	wasSelected := b.selected
	b.selected = (passband == b.bandwidth) && (b.currentMode == b.mode)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SetModeButton) SetFrequency(frequency client.Frequency) {
	b.currentFrequency = frequency
	wasInModePortion := b.inModePortion
	b.inModePortion = isInModePortion(frequency, b.bandplanMode)
	if b.inModePortion == wasInModePortion {
		return
	}
	b.Invalidate(false)
}

func (b *SetModeButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || b.inModePortionImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	if b.inModePortion {
		return b.inModePortionImage
	}
	return b.image
}

func (b *SetModeButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	text := string(b.mode)
	if b.label != "" {
		text = b.label
	}
	b.image = b.redrawButton(gc, text)

	gc.SwapColors()
	b.selectedImage = b.redrawButton(gc, text)

	gc.SwapColors()
	gc.SetBackground(hamdeck.Blue)
	b.inModePortionImage = b.redrawButton(gc, text)
}

func (b *SetModeButton) redrawButton(gc hamdeck.GraphicContext, text string) image.Image {
	if b.icon == "" {
		return gc.DrawSingleLineTextButton(text)
	}

	gc.SetFontSize(16)
	if b.iconImage == nil {
		iconFile := b.icon + ".png"
		b.iconImage = gc.LoadIconAsset(iconFile)
	}
	return gc.DrawIconLabelButton(b.iconImage, text)
}

func (b *SetModeButton) Pressed() {
	b.longpress.Pressed()
	if !b.enabled {
		return
	}
	ctx := b.client.WithRequestTimeout()
	err := b.client.Conn.SetModeAndPassband(ctx, b.mode, hamradio.Frequency(b.bandwidth))
	if err != nil {
		log.Printf("cannot set mode: %v", err)
	}
}

func (b *SetModeButton) Released() {
	b.longpress.Released()
}

func (b *SetModeButton) OnLongpress() {
	if !b.enabled {
		return
	}
	frequency := findModePortionCenter(b.currentFrequency, b.bandplanMode)
	ctx := b.client.WithRequestTimeout()
	err := b.client.Conn.SetFrequency(ctx, frequency)
	if err != nil {
		log.Printf("cannot jump to the beginning of the %s band portion: %v", b.mode, err)
	}
}

func isInModePortion(f client.Frequency, mode bandplan.Mode) bool {
	currentPortion, ok := findPortion(f)
	if !ok {
		return false
	}
	return currentPortion.Mode == mode
}

func findModePortionCenter(f client.Frequency, mode bandplan.Mode) client.Frequency {
	band := bandplan.IARURegion1.ByFrequency(f)
	var modePortion bandplan.Portion
	var currentPortion bandplan.Portion
	for _, portion := range band.Portions {
		if (portion.Mode == mode && portion.From < f) || modePortion.Mode != mode {
			modePortion = portion
		}
		if portion.Contains(f) {
			currentPortion = portion
		}
		if modePortion.Mode == mode && currentPortion.Mode != "" {
			break
		}
	}
	if currentPortion.Mode == mode {
		return currentPortion.Center()
	}
	if modePortion.Mode == mode {
		return modePortion.Center()
	}
	return band.Center()
}

func findPortion(f client.Frequency) (bandplan.Portion, bool) {
	band := bandplan.IARURegion1.ByFrequency(f)
	for _, portion := range band.Portions {
		if portion.Contains(f) {
			return portion, true
		}
	}
	return bandplan.Portion{}, false
}

/*
	ToggleModeButton
*/

func NewToggleModeButton(hamlibClient *HamlibClient, mode1 client.Mode, label1 string, mode2 client.Mode, label2 string) *ToggleModeButton {
	result := &ToggleModeButton{
		client:        hamlibClient,
		enabled:       hamlibClient.Connected(),
		modes:         []client.Mode{mode1, mode2},
		bandplanModes: []bandplan.Mode{mode1.ToBandplanMode(), mode2.ToBandplanMode()},
		labels:        []string{label1, label2},
	}
	result.longpress = hamdeck.NewLongpressHandler(result.OnLongpress)

	hamlibClient.Listen(result)

	return result
}

type ToggleModeButton struct {
	hamdeck.BaseButton
	client             *HamlibClient
	image              image.Image
	selectedImage      image.Image
	inModePortionImage image.Image
	enabled            bool
	selected           bool
	inModePortion      bool
	modes              []client.Mode
	bandplanModes      []bandplan.Mode
	labels             []string
	currentMode        int
	currentFrequency   client.Frequency
	longpress          *hamdeck.LongpressHandler
}

func (b *ToggleModeButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
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
	b.Invalidate(true)
}

func (b *ToggleModeButton) SetFrequency(frequency client.Frequency) {
	b.currentFrequency = frequency
	wasInModePortion := b.inModePortion
	b.inModePortion = false
	for _, mode := range b.bandplanModes {
		if isInModePortion(frequency, mode) {
			b.inModePortion = true
		}
	}
	if b.inModePortion == wasInModePortion {
		return
	}
	b.Invalidate(false)
}

func (b *ToggleModeButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || b.inModePortionImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	if b.inModePortion {
		return b.inModePortionImage
	}
	return b.image
}

func (b *ToggleModeButton) redrawImages(gc hamdeck.GraphicContext) {
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
	b.image = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.currentMode+1)

	gc.SwapColors()
	b.selectedImage = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.currentMode+1)

	gc.SwapColors()
	gc.SetBackground(hamdeck.Blue)
	b.inModePortionImage = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.currentMode+1)
}

func (b *ToggleModeButton) Pressed() {
	b.longpress.Pressed()
	if !b.enabled {
		return
	}
	mode := b.currentMode
	if b.selected {
		mode = (mode + 1) % len(b.modes)
	}
	ctx := b.client.WithRequestTimeout()
	err := b.client.Conn.SetModeAndPassband(ctx, b.modes[mode], 0)
	if err != nil {
		log.Printf("cannot set mode: %v", err)
	}
}

func (b *ToggleModeButton) Released() {
	b.longpress.Released()
}

func (b *ToggleModeButton) OnLongpress() {
	if !b.enabled {
		return
	}
	frequency := findModePortionCenter(b.currentFrequency, b.bandplanModes[b.currentMode])
	ctx := b.client.WithRequestTimeout()
	err := b.client.Conn.SetFrequency(ctx, frequency)
	if err != nil {
		log.Printf("cannot jump to the beginning of the %s band portion: %v", b.modes[b.currentMode], err)
	}
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
	b.Invalidate(true)
}

func (b *SetButton) Image(gc hamdeck.GraphicContext, redrawImage bool) image.Image {
	if b.image == nil || redrawImage {
		if b.enabled {
			gc.SetForeground(hamdeck.White)
		} else {
			gc.SetForeground(hamdeck.DisabledGray)
		}
		b.image = gc.DrawSingleLineTextButton(b.label)
	}
	return b.image
}

func (b *SetButton) Pressed() {
	if !b.enabled {
		return
	}
	ctx := b.client.WithRequestTimeout()
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

func NewSwitchToBandButton(hamlibClient *HamlibClient, label string, bandName string, useUpDown bool) *SwitchToBandButton {
	band, ok := bandplan.IARURegion1[bandplan.BandName(bandName)]
	if !ok {
		log.Printf("cannot find band %s in IARU Region 1 bandplan", bandName)
		return nil
	}
	result := &SwitchToBandButton{
		client:    hamlibClient,
		enabled:   hamlibClient.Connected(),
		label:     label,
		band:      band,
		useUpDown: useUpDown,
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
	mode          client.Mode
	useUpDown     bool
}

func (b *SwitchToBandButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *SwitchToBandButton) SetFrequency(frequency client.Frequency) {
	wasSelected := b.selected
	b.selected = b.band.Contains(frequency)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SwitchToBandButton) SetMode(mode client.Mode) {
	b.mode = mode
}

func (b *SwitchToBandButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *SwitchToBandButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	text := string(b.band.Name)
	if b.label != "" {
		text = b.label
	}
	b.image = gc.DrawSingleLineTextButton(text)
	gc.SwapColors()
	b.selectedImage = gc.DrawSingleLineTextButton(text)
}

func (b *SwitchToBandButton) Pressed() {
	if !b.enabled {
		return
	}
	if b.useUpDown {
		ctx := b.client.WithRequestTimeout()
		err := b.client.Conn.SwitchToBand(ctx, b.band)
		if err != nil {
			log.Print(err)
		}
	} else {
		frequency := findModePortionCenter(b.band.Center(), b.mode.ToBandplanMode())
		ctx := b.client.WithRequestTimeout()
		err := b.client.Conn.SetFrequency(ctx, frequency)
		if err != nil {
			log.Printf("cannot switch to band %s: %v", b.band, err)
		}
		ctx = b.client.WithRequestTimeout()
		err = b.client.Conn.SetModeAndPassband(ctx, b.mode, 0)
		if err != nil {
			log.Printf("cannot switch band to mode %s: %v", b.mode, err)
		}
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
	b.Invalidate(true)
}

func (b *SetPowerLevelButton) SetPowerLevel(powerLevel float64) {
	wasSelected := b.selected
	b.selected = (powerLevel == b.value)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SetPowerLevelButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *SetPowerLevelButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	b.image = gc.DrawSingleLineTextButton(b.label)
	gc.SwapColors()
	b.selectedImage = gc.DrawSingleLineTextButton(b.label)
}

func (b *SetPowerLevelButton) Pressed() {
	if !b.enabled {
		return
	}
	ctx := b.client.WithRequestTimeout()
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
	b.flashOn = false
	b.Invalidate(true)
}

func (b *MOXButton) Flash(flashOn bool) {
	if !(b.selected && b.enabled) {
		return
	}
	b.flashOn = flashOn
	b.Invalidate(false)
}

func (b *MOXButton) SetPTT(ptt client.PTT) {
	wasSelected := b.selected
	b.selected = (ptt != client.PTTRx)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *MOXButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.flashImage == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if !b.selected {
		return b.image
	}
	if b.flashOn {
		return b.flashImage
	}
	return b.selectedImage
}

func (b *MOXButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	b.image = gc.DrawSingleLineTextButton(b.label)
	gc.SetBackground(hamdeck.Red)
	b.flashImage = gc.DrawSingleLineTextButton(b.label)
	gc.SwapColors()
	b.selectedImage = gc.DrawSingleLineTextButton(b.label)
}

func (b *MOXButton) Pressed() {
	if !b.enabled {
		return
	}
	value := client.PTTTx
	if b.selected {
		value = client.PTTRx
	}
	ctx := b.client.WithRequestTimeout()
	err := b.client.Conn.SetPTT(ctx, value)
	if err != nil {
		log.Print(err)
	}
}

func (b *MOXButton) Released() {
	// ignore
}

/*
	SetVFOButton
*/

func NewSetVFOButton(hamlibClient *HamlibClient, label string, vfo client.VFO) *SetVFOButton {
	result := &SetVFOButton{
		client:  hamlibClient,
		enabled: hamlibClient.Connected(),
		label:   label,
		vfo:     vfo,
	}

	hamlibClient.Listen(result)

	return result
}

type SetVFOButton struct {
	hamdeck.BaseButton
	client        *HamlibClient
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	label         string
	vfo           client.VFO
}

func (b *SetVFOButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *SetVFOButton) SetVFO(vfo client.VFO) {
	wasSelected := b.selected
	b.selected = (vfo == b.vfo)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SetVFOButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *SetVFOButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	b.image = gc.DrawSingleLineTextButton(b.label)
	gc.SwapColors()
	b.selectedImage = gc.DrawSingleLineTextButton(b.label)
}

func (b *SetVFOButton) Pressed() {
	if !b.enabled {
		return
	}
	ctx := b.client.WithRequestTimeout()
	err := b.client.Conn.SetVFO(ctx, b.vfo)
	if err != nil {
		log.Print(err)
	}
}

func (b *SetVFOButton) Released() {
	// ignore
}
