package tci

import (
	"fmt"
	"image"
	"log"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/tci/client"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

/*
	SetModeButton
*/

func NewSetModeButton(tciClient *Client, mode client.Mode, label string) *SetModeButton {
	result := &SetModeButton{
		client:           tciClient,
		enabled:          tciClient.Connected(),
		mode:             mode,
		bandplanMode:     toBandplanMode(mode),
		label:            label,
		currentTRX:       0,
		currentMode:      make(map[int]client.Mode),
		currentFrequency: make(map[int]int),
	}
	result.longpress = hamdeck.NewLongpressHandler(result.OnLongpress)

	tciClient.Notify(result)

	return result
}

type SetModeButton struct {
	hamdeck.BaseButton
	client             *Client
	image              image.Image
	selectedImage      image.Image
	inModePortionImage image.Image
	enabled            bool
	selected           bool
	inModePortion      bool
	mode               client.Mode
	bandplanMode       bandplan.Mode
	label              string
	currentTRX         int
	currentMode        map[int]client.Mode
	currentFrequency   map[int]int
	longpress          *hamdeck.LongpressHandler
}

func (b *SetModeButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *SetModeButton) SetTRX(trx int) {
	b.currentTRX = trx

	wasSelected := b.selected
	b.selected = (b.currentMode[trx] == b.mode)

	wasInModePortion := b.inModePortion
	b.inModePortion = isInModePortion(b.currentFrequency[trx], b.bandplanMode)

	if (b.selected != wasSelected) || (b.inModePortion != wasInModePortion) {
		b.Invalidate(false)
	}
}

func (b *SetModeButton) SetMode(trx int, mode client.Mode) {
	b.currentMode[trx] = mode
	if trx != b.currentTRX {
		return
	}

	wasSelected := b.selected
	b.selected = (mode == b.mode)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SetModeButton) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if vfo == client.VFOA {
		b.currentFrequency[trx] = frequency
	}
	if trx != b.currentTRX || vfo != client.VFOA {
		return
	}

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

	b.image = gc.DrawSingleLineTextButton(text)

	gc.SwapColors()
	b.selectedImage = gc.DrawSingleLineTextButton(text)

	gc.SwapColors()
	gc.SetBackground(hamdeck.Blue)
	b.inModePortionImage = gc.DrawSingleLineTextButton(text)
}

func (b *SetModeButton) Pressed() {
	b.longpress.Pressed()
	if !b.enabled {
		return
	}
	err := b.client.SetMode(b.currentTRX, b.mode)
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
	frequency := findModePortionCenter(b.currentFrequency[b.currentTRX], b.bandplanMode)
	err := b.client.SetDDS(b.currentTRX, frequency)
	if err != nil {
		log.Printf("cannot jump to the center of the %s band portion: %v", b.bandplanMode, err)
	}
	err = b.client.SetVFOFrequency(b.currentTRX, client.VFOA, frequency)
	if err != nil {
		log.Printf("cannot jump to the center of the %s band portion: %v", b.bandplanMode, err)
	}
}

func isInModePortion(f int, mode bandplan.Mode) bool {
	currentPortion, ok := findPortion(f)
	if !ok {
		return false
	}
	return currentPortion.Mode == mode
}

func findModePortionCenter(f int, mode bandplan.Mode) int {
	frequency := hamradio.Frequency(f)
	band := bandplan.IARURegion1.ByFrequency(frequency)
	var modePortion bandplan.Portion
	var currentPortion bandplan.Portion
	for _, portion := range band.Portions {
		if (portion.Mode == mode && portion.From < frequency) || modePortion.Mode != mode {
			modePortion = portion
		}
		if portion.Contains(frequency) {
			currentPortion = portion
		}
		if modePortion.Mode == mode && currentPortion.Mode != "" {
			break
		}
	}
	if currentPortion.Mode == mode {
		return int(currentPortion.Center())
	}
	if modePortion.Mode == mode {
		return int(modePortion.Center())
	}
	return int(band.Center())
}

func findPortion(f int) (bandplan.Portion, bool) {
	frequency := hamradio.Frequency(f)
	band := bandplan.IARURegion1.ByFrequency(frequency)
	for _, portion := range band.Portions {
		if portion.Contains(frequency) {
			return portion, true
		}
	}
	return bandplan.Portion{}, false
}

/*
	ToggleModeButton
*/

func NewToggleModeButton(tciClient *Client, mode1 client.Mode, label1 string, mode2 client.Mode, label2 string) *ToggleModeButton {
	result := &ToggleModeButton{
		client:           tciClient,
		enabled:          tciClient.Connected(),
		modes:            []client.Mode{mode1, mode2},
		bandplanModes:    []bandplan.Mode{toBandplanMode(mode1), toBandplanMode(mode2)},
		labels:           []string{label1, label2},
		currentTRX:       0,
		currentMode:      make(map[int]client.Mode),
		currentFrequency: make(map[int]int),
	}
	result.longpress = hamdeck.NewLongpressHandler(result.OnLongpress)

	tciClient.Notify(result)

	return result
}

type ToggleModeButton struct {
	hamdeck.BaseButton
	client             *Client
	image              image.Image
	selectedImage      image.Image
	inModePortionImage image.Image
	enabled            bool
	selected           bool
	inModePortion      bool
	modes              []client.Mode
	selectedModeIndex  int
	bandplanModes      []bandplan.Mode
	labels             []string
	currentTRX         int
	currentMode        map[int]client.Mode
	currentFrequency   map[int]int
	longpress          *hamdeck.LongpressHandler
}

func (b *ToggleModeButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *ToggleModeButton) SetTRX(trx int) {
	b.currentTRX = trx

	wasSelected := b.selected
	lastModeIndex := b.selectedModeIndex

	b.selected = false
	for i, m := range b.modes {
		if b.currentMode[trx] == m {
			b.selectedModeIndex = i
			b.selected = true
			break
		}
	}

	wasInModePortion := b.inModePortion
	b.inModePortion = isInModePortion(b.currentFrequency[b.currentTRX], b.bandplanModes[b.selectedModeIndex])

	if (b.selected != wasSelected) || (b.selectedModeIndex != lastModeIndex) || (b.inModePortion != wasInModePortion) {
		b.Invalidate(false)
	}
}

func (b *ToggleModeButton) SetMode(trx int, mode client.Mode) {
	b.currentMode[b.currentTRX] = mode
	if trx != b.currentTRX {
		return
	}

	wasSelected := b.selected
	lastModeIndex := b.selectedModeIndex

	b.selected = false
	for i, m := range b.modes {
		if mode == m {
			b.selectedModeIndex = i
			b.selected = true
			break
		}
	}
	if (b.selected == wasSelected) && (b.selectedModeIndex == lastModeIndex) {
		return
	}
	b.Invalidate(true)
}

func (b *ToggleModeButton) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if vfo == client.VFOA {
		b.currentFrequency[trx] = frequency
	}
	if trx != b.currentTRX || vfo != client.VFOA {
		return
	}

	wasInModePortion := b.inModePortion
	b.inModePortion = isInModePortion(frequency, b.bandplanModes[b.selectedModeIndex])
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
	b.image = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.selectedModeIndex+1)

	gc.SwapColors()
	b.selectedImage = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.selectedModeIndex+1)

	gc.SwapColors()
	gc.SetBackground(hamdeck.Blue)
	b.inModePortionImage = gc.DrawDoubleLineToggleTextButton(text[0], text[1], b.selectedModeIndex+1)
}

func (b *ToggleModeButton) Pressed() {
	b.longpress.Pressed()
	if !b.enabled {
		return
	}
	modeIndex := b.selectedModeIndex
	if b.selected {
		modeIndex = (modeIndex + 1) % len(b.modes)
	}
	err := b.client.SetMode(b.currentTRX, b.modes[modeIndex])
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
	frequency := findModePortionCenter(b.currentFrequency[b.currentTRX], b.bandplanModes[b.selectedModeIndex])
	err := b.client.SetDDS(b.currentTRX, frequency)
	if err != nil {
		log.Printf("cannot jump to the beginning of the %s band portion: %v", b.bandplanModes[b.selectedModeIndex], err)
	}
	err = b.client.SetVFOFrequency(b.currentTRX, client.VFOA, frequency)
	if err != nil {
		log.Printf("cannot jump to the beginning of the %s band portion: %v", b.bandplanModes[b.selectedModeIndex], err)
	}
}

/*
	SetFilterButton
*/

func NewSetFilterButton(tciClient *Client, bottomFrequency int, topFrequency int, label string, icon string) *SetFilterButton {
	result := &SetFilterButton{
		client:          tciClient,
		enabled:         tciClient.Connected(),
		selected:        make(map[int]bool),
		bottomFrequency: bottomFrequency,
		topFrequency:    topFrequency,
		label:           label,
		icon:            icon,
		currentTRX:      0,
	}

	tciClient.Notify(result)

	return result
}

type SetFilterButton struct {
	hamdeck.BaseButton
	client          *Client
	image           image.Image
	selectedImage   image.Image
	enabled         bool
	selected        map[int]bool
	bottomFrequency int
	topFrequency    int
	label           string
	icon            string
	currentTRX      int
}

func (b *SetFilterButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *SetFilterButton) SetTRX(trx int) {
	wasSelected := b.selected[b.currentTRX]
	b.currentTRX = trx
	if b.selected[trx] != wasSelected {
		b.Invalidate(false)
	}
}

func (b *SetFilterButton) SetRXFilterBand(trx int, min, max int) {
	wasSelected := b.selected[trx]
	b.selected[trx] = (b.bottomFrequency == min) && (b.topFrequency == max)

	if (trx == b.currentTRX) && (b.selected[trx] != wasSelected) {
		b.Invalidate(false)
	}
}

func (b *SetFilterButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected[b.currentTRX] {
		return b.selectedImage
	}
	return b.image
}

func (b *SetFilterButton) redrawImages(gc hamdeck.GraphicContext) {
	gc.SetFontSize(16)
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}

	iconFile := b.icon + ".png"
	b.image = gc.DrawIconLabelButton(gc.LoadIconAsset(iconFile), b.label)

	gc.SwapColors()
	b.selectedImage = gc.DrawIconLabelButton(gc.LoadIconAsset(iconFile), b.label)
}

func (b *SetFilterButton) Pressed() {
	if !b.enabled {
		return
	}
	err := b.client.SetRXFilterBand(b.currentTRX, b.bottomFrequency, b.topFrequency)
	if err != nil {
		log.Printf("cannot set rx filter band: %v", err)
	}
}

func (b *SetFilterButton) Released() {
	// ignore
}

/*
	MOXButton
*/

func NewMOXButton(tciClient *Client, label string) *MOXButton {
	if label == "" {
		label = "TX"
	}

	result := &MOXButton{
		client:     tciClient,
		enabled:    tciClient.Connected(),
		label:      label,
		currentPTT: make(map[int]bool),
	}

	tciClient.Notify(result)

	return result
}

type MOXButton struct {
	hamdeck.BaseButton
	client        *Client
	image         image.Image
	selectedImage image.Image
	flashImage    image.Image
	enabled       bool
	selected      bool
	flashOn       bool
	label         string
	currentTRX    int
	currentPTT    map[int]bool
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

func (b *MOXButton) SetTX(trx int, ptt bool) {
	b.currentPTT[trx] = ptt
	if trx != b.currentTRX {
		return
	}

	wasSelected := b.selected
	b.selected = ptt
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
	value := !b.selected
	err := b.client.SetTX(b.currentTRX, value, client.SignalSourceDefault)
	if err != nil {
		log.Printf("cannot set PTT to %t: %v", value, err)
	}
}

func (b *MOXButton) Released() {
	// ignore
}

/*
	TuneButton
*/

func NewTuneButton(tciClient *Client, label string) *TuneButton {
	if label == "" {
		label = "Tune"
	}

	result := &TuneButton{
		client:     tciClient,
		enabled:    tciClient.Connected(),
		label:      label,
		currentPTT: make(map[int]bool),
	}

	tciClient.Notify(result)

	return result
}

type TuneButton struct {
	hamdeck.BaseButton
	client        *Client
	image         image.Image
	selectedImage image.Image
	flashImage    image.Image
	enabled       bool
	selected      bool
	flashOn       bool
	label         string
	currentTRX    int
	currentPTT    map[int]bool
}

func (b *TuneButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.flashOn = false
	b.Invalidate(true)
}

func (b *TuneButton) Flash(flashOn bool) {
	if !(b.selected && b.enabled) {
		return
	}
	b.flashOn = flashOn
	b.Invalidate(false)
}

func (b *TuneButton) SetTune(trx int, ptt bool) {
	b.currentPTT[trx] = ptt
	if trx != b.currentTRX {
		return
	}

	wasSelected := b.selected
	b.selected = ptt
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *TuneButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
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

func (b *TuneButton) redrawImages(gc hamdeck.GraphicContext) {
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

func (b *TuneButton) Pressed() {
	if !b.enabled {
		return
	}
	value := !b.selected
	err := b.client.SetTune(b.currentTRX, value)
	if err != nil {
		log.Printf("cannot set Tune to %t: %v", value, err)
	}
}

func (b *TuneButton) Released() {
	// ignore
}

/*
	MuteButton
*/

func NewMuteButton(tciClient *Client, label string) *MuteButton {
	if label == "" {
		label = "Main"
	}

	result := &MuteButton{
		client:  tciClient,
		enabled: tciClient.Connected(),
		label:   label,
	}

	tciClient.Notify(result)

	return result
}

type MuteButton struct {
	hamdeck.BaseButton
	client        *Client
	image         image.Image
	selectedImage image.Image
	flashImage    image.Image
	enabled       bool
	selected      bool
	label         string
}

func (b *MuteButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *MuteButton) SetMute(muted bool) {
	wasSelected := b.selected
	b.selected = muted
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *MuteButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.flashImage == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if !b.selected {
		return b.image
	}
	return b.selectedImage
}

func (b *MuteButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	b.image = gc.DrawIconLabelButton(gc.LoadIconAsset("volume_off.png"), b.label)

	if b.enabled {
		gc.SetBackground(hamdeck.Red)
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.Red)
	}
	b.selectedImage = gc.DrawIconLabelButton(gc.LoadIconAsset("volume_off.png"), b.label)
}

func (b *MuteButton) Pressed() {
	if !b.enabled {
		return
	}
	value := !b.selected
	err := b.client.SetMute(value)
	if err != nil {
		log.Printf("cannot set Mute to %t: %v", value, err)
	}
}

func (b *MuteButton) Released() {
	// ignore
}

/*
	SetDriveButton
*/

func NewSetDriveButton(tciClient *Client, label string, value int) *SetDriveButton {
	result := &SetDriveButton{
		client:  tciClient,
		enabled: tciClient.Connected(),
		label:   label,
		value:   value,
	}

	tciClient.Notify(result)

	return result
}

type SetDriveButton struct {
	hamdeck.BaseButton
	client        *Client
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	label         string
	value         int
}

func (b *SetDriveButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *SetDriveButton) SetDrive(percent int) {
	wasSelected := b.selected
	b.selected = (percent == b.value)
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SetDriveButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *SetDriveButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	b.image = gc.DrawSingleLineTextButton(b.label)
	gc.SwapColors()
	b.selectedImage = gc.DrawSingleLineTextButton(b.label)
}

func (b *SetDriveButton) Pressed() {
	if !b.enabled {
		return
	}
	err := b.client.SetDrive(b.value)
	if err != nil {
		log.Printf("cannot set drive to %d: %v", b.value, err)
	}
}

func (b *SetDriveButton) Released() {
	// ignore
}

/*
	IncrementDriveButton
*/

func NewIncrementDriveButton(tciClient *Client, label string, increment int) *IncrementDriveButton {
	result := &IncrementDriveButton{
		client:    tciClient,
		enabled:   tciClient.Connected(),
		label:     label,
		increment: increment,
	}
	result.longpress = hamdeck.NewLongpressHandler(result.OnLongpress)

	tciClient.Notify(result)

	return result
}

type IncrementDriveButton struct {
	hamdeck.BaseButton
	client        *Client
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	label         string
	increment     int
	currentValue  int
	longpress     *hamdeck.LongpressHandler
}

func (b *IncrementDriveButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *IncrementDriveButton) SetDrive(percent int) {
	b.currentValue = percent
	wasSelected := b.selected
	if b.increment > 0 {
		b.selected = (percent == 100)
	} else {
		b.selected = (percent == 0)
	}
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(true)
}

func (b *IncrementDriveButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *IncrementDriveButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	text := b.label
	if b.selected {
		text = fmt.Sprintf("%d%%", b.currentValue)
	}
	b.image = gc.DrawSingleLineTextButton(text)
	gc.SwapColors()
	b.selectedImage = gc.DrawSingleLineTextButton(text)
}

func (b *IncrementDriveButton) Pressed() {
	b.longpress.Pressed()
	if !b.enabled {
		return
	}
	value := b.currentValue + b.increment
	err := b.client.SetDrive(value)
	if err != nil {
		log.Printf("cannot increment drive to %d: %v", value, err)
	}
}

func (b *IncrementDriveButton) Released() {
	b.longpress.Released()
}

func (b *IncrementDriveButton) OnLongpress() {
	if !b.enabled {
		return
	}
	var value int
	if b.increment > 0 {
		value = 100
	} else {
		value = 0
	}
	err := b.client.SetDrive(value)
	if err != nil {
		log.Printf("cannot increment drive to %d: %v", value, err)
	}
}

/*
	IncrementVolumeButton
*/

func NewIncrementVolumeButton(tciClient *Client, label string, increment int) *IncrementVolumeButton {
	result := &IncrementVolumeButton{
		client:    tciClient,
		enabled:   tciClient.Connected(),
		label:     label,
		increment: increment,
	}

	tciClient.Notify(result)

	return result
}

type IncrementVolumeButton struct {
	hamdeck.BaseButton
	client        *Client
	image         image.Image
	selectedImage image.Image
	enabled       bool
	selected      bool
	label         string
	increment     int
	currentValue  int
}

func (b *IncrementVolumeButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *IncrementVolumeButton) SetVolume(dB int) {
	b.currentValue = dB
	wasSelected := b.selected
	if b.increment > 0 {
		b.selected = (dB == 0)
	} else {
		b.selected = (dB == -60)
	}
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(true)
}

func (b *IncrementVolumeButton) Image(gc hamdeck.GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || b.selectedImage == nil || redrawImages {
		b.redrawImages(gc)
	}
	if b.selected {
		return b.selectedImage
	}
	return b.image
}

func (b *IncrementVolumeButton) redrawImages(gc hamdeck.GraphicContext) {
	if b.enabled {
		gc.SetForeground(hamdeck.White)
	} else {
		gc.SetForeground(hamdeck.DisabledGray)
	}
	text := b.label
	if b.selected {
		text = fmt.Sprintf("%ddB", b.currentValue)
	}
	var imageName string
	if b.increment > 0 {
		imageName = "volume_up.png"
	} else {
		imageName = "volume_down.png"
	}
	b.image = gc.DrawIconLabelButton(gc.LoadIconAsset(imageName), text)
	gc.DrawSingleLineTextButton(text)
	gc.SwapColors()
	b.selectedImage = gc.DrawIconLabelButton(gc.LoadIconAsset(imageName), text)
}

func (b *IncrementVolumeButton) Pressed() {
	if !b.enabled {
		return
	}
	value := b.currentValue + b.increment
	err := b.client.SetVolume(value)
	if err != nil {
		log.Printf("cannot increment volume to %d: %v", value, err)
	}
}

func (b *IncrementVolumeButton) Released() {
	// ignore
}

/*
	SwitchToBandButton
*/

func NewSwitchToBandButton(tciClient *Client, label string, bandName string) *SwitchToBandButton {
	band, ok := bandplan.IARURegion1[bandplan.BandName(bandName)]
	if !ok {
		log.Printf("cannot find band %s in IARU Region 1 bandplan", bandName)
		return nil
	}
	result := &SwitchToBandButton{
		client:           tciClient,
		enabled:          tciClient.Connected(),
		label:            label,
		band:             band,
		currentFrequency: make(map[int]int),
		currentMode:      make(map[int]client.Mode),
	}

	tciClient.Notify(result)

	return result
}

type SwitchToBandButton struct {
	hamdeck.BaseButton
	client           *Client
	image            image.Image
	selectedImage    image.Image
	enabled          bool
	selected         bool
	label            string
	band             bandplan.Band
	currentTRX       int
	currentFrequency map[int]int
	currentMode      map[int]client.Mode
}

func (b *SwitchToBandButton) Enable(enabled bool) {
	if enabled == b.enabled {
		return
	}
	b.enabled = enabled
	b.Invalidate(true)
}

func (b *SwitchToBandButton) SetTRX(trx int) {
	b.currentTRX = trx

	wasSelected := b.selected
	b.selected = b.band.Contains(hamradio.Frequency(b.currentFrequency[trx]))
	if b.selected != wasSelected {
		b.Invalidate(false)
	}
}

func (b *SwitchToBandButton) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if vfo == client.VFOA {
		b.currentFrequency[trx] = frequency
	}
	if trx != b.currentTRX || vfo != client.VFOA {
		return
	}

	wasSelected := b.selected
	b.selected = b.band.Contains(hamradio.Frequency(frequency))
	if b.selected == wasSelected {
		return
	}
	b.Invalidate(false)
}

func (b *SwitchToBandButton) SetMode(trx int, mode client.Mode) {
	b.currentMode[trx] = mode
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
	mode := b.currentMode[b.currentTRX]
	frequency := findModePortionCenter(int(b.band.Center()), toBandplanMode(mode))
	err := b.client.SetVFOFrequency(b.currentTRX, client.VFOA, frequency)
	if err != nil {
		log.Printf("cannot switch to band %s: %v", b.band, err)
	}
	err = b.client.SetMode(b.currentTRX, mode)
	if err != nil {
		log.Printf("cannot switch band to mode %s: %v", mode, err)
	}
}

func (b *SwitchToBandButton) Released() {
	// ignore
}
