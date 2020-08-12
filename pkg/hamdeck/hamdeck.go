package hamdeck

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"time"
)

type Key struct {
	Index   int
	Pressed bool
}

type Device interface {
	Close() error
	ID() string
	Serial() string
	FirmwareVersion() string
	Pixels() int
	Rows() int
	Columns() int
	Clear() error
	Reset() error
	SetBrightness(int) error
	SetImage(int, image.Image) error
	ReadKeys() (chan Key, error)
}

type GraphicContext interface {
	Pixels() int
	Reset()
	SetBackground(background color.Color)
	SetForeground(foreground color.Color)
	SwapColors()
	SetFont(filename string)
	SetFontSize(points float64)
	DrawNoButton() image.Image
	DrawSingleLineTextButton(text string) image.Image
	DrawDoubleLineToggleTextButton(text1, text2 string, activeLine int) image.Image
	LoadIconFromFile(filename string) (image.Image, error)
	LoadIconFromReader(r io.Reader) (image.Image, error)
	LoadIconAsset(name string) image.Image
	DrawIconButton(icon image.Image) image.Image
	DrawIconLabelButton(icon image.Image, label string) image.Image
}

type ButtonContext interface {
	Invalidate(bool)
}

type Button interface {
	Image(GraphicContext, bool) image.Image
	Pressed()
	Released()
	Attached(ButtonContext)
	Detached()
}

type FlashingButton interface {
	Flash(on bool)
}

const FlashingInterval = 500 * time.Millisecond

type Enabler interface {
	Enable(enabled bool)
}

func NotifyEnablers(listeners []interface{}, enabled bool) {
	for _, listener := range listeners {
		enabler, ok := listener.(Enabler)
		if ok {
			enabler.Enable(enabled)
		}
	}
}

type ButtonFactory interface {
	Close()
	CreateButton(config map[string]interface{}) Button
}

type HamDeck struct {
	device    Device
	gc        GraphicContext
	buttons   []Button
	noButton  Button
	flashOn   bool
	factories []ButtonFactory
}

func New(device Device) *HamDeck {
	buttonCount := device.Columns() * device.Rows()
	result := &HamDeck{
		device:  device,
		gc:      NewGraphicContext(device.Pixels()),
		buttons: make([]Button, buttonCount),
	}
	result.noButton = &noButton{image: result.gc.DrawNoButton()}
	for i := range result.buttons {
		result.buttons[i] = result.noButton
	}

	result.device.Clear()
	result.RedrawAll(true)

	return result
}

func (d *HamDeck) RegisterFactory(factory ButtonFactory) {
	d.factories = append(d.factories, factory)
}

func (d *HamDeck) RedrawAll(redrawImages bool) {
	for i, b := range d.buttons {
		d.gc.Reset()
		d.device.SetImage(i, b.Image(d.gc, redrawImages))
	}
}

func (d *HamDeck) Redraw(index int, redrawImages bool) {
	d.gc.Reset()
	d.device.SetImage(index, d.buttons[index].Image(d.gc, redrawImages))
}

func (d *HamDeck) Attach(index int, button Button) {
	d.buttons[index] = button

	ctx := &buttonContext{index: index, deck: d}
	button.Attached(ctx)
	d.Redraw(index, true)
}

func (d *HamDeck) Detach(index int) {
	d.buttons[index].Detached()
	d.buttons[index] = d.noButton
	d.Redraw(index, true)
}

func (d *HamDeck) Run(stop <-chan struct{}) error {
	keys, err := d.device.ReadKeys()
	if err != nil {
		return fmt.Errorf("cannot read keys from Stream Deck: %w", err)
	}

	flashTicker := time.NewTicker(FlashingInterval)
	defer flashTicker.Stop()

MainLoop:
	for {
		select {
		case key, ok := <-keys:
			if !ok {
				log.Print("The Stream Deck device closed the connection.")
				return nil
			}
			d.handleKey(key)
		case <-flashTicker.C:
			d.flash()
		case <-stop:
			break MainLoop
		}
	}

	err = d.device.Reset()
	if err != nil {
		return fmt.Errorf("cannot reset Stream Deck: %w", err)
	}
	return nil
}

func (d *HamDeck) handleKey(key Key) {
	if (key.Index < 0) || (int(key.Index) >= len(d.buttons)) {
		return
	}
	button := d.buttons[key.Index]

	if key.Pressed {
		button.Pressed()
	} else {
		button.Released()
	}
}

func (d *HamDeck) flash() {
	d.flashOn = !d.flashOn
	for _, button := range d.buttons {
		flashingButton, ok := button.(FlashingButton)
		if ok {
			flashingButton.Flash(d.flashOn)
		}
	}
}

type BaseButton struct {
	ctx ButtonContext
}

func (b *BaseButton) Invalidate(redrawImages bool) {
	if b.ctx == nil {
		return
	}
	b.ctx.Invalidate(redrawImages)
}

func (b *BaseButton) Attached(ctx ButtonContext) {
	b.ctx = ctx
}

func (b *BaseButton) Detached() {
	b.ctx = nil
}

type noButton struct {
	image image.Image
}

func (b *noButton) Image(GraphicContext, bool) image.Image {
	return b.image
}

func (b *noButton) Pressed()               {}
func (b *noButton) Released()              {}
func (b *noButton) Attached(ButtonContext) {}
func (b *noButton) Detached()              {}

type buttonContext struct {
	index int
	deck  *HamDeck
}

func (c *buttonContext) Invalidate(redrawImages bool) {
	c.deck.Redraw(c.index, redrawImages)
}
