package examples

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

/*
	The HelloTextButton shows the text "Hello" while released and "World" while pressed.
*/

func NewHelloTextButton() *HelloTextButton {
	return &HelloTextButton{}
}

type HelloTextButton struct {
	hamdeck.BaseButton
	image      image.Image
	helloImage image.Image
	worldImage image.Image
}

func (b *HelloTextButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.helloImage == nil {
		b.helloImage = gc.DrawSingleLineTextButton("Hello")
	}
	if b.worldImage == nil {
		b.worldImage = gc.DrawSingleLineTextButton("World")
	}
	if b.image == nil {
		b.image = b.helloImage
	}
	return b.image
}

func (b *HelloTextButton) Pressed() {
	log.Print("Hello World pressed")
	b.image = b.worldImage
	b.Invalidate()
}

func (b *HelloTextButton) Released() {
	log.Print("Hello World released")
	b.image = b.helloImage
	b.Invalidate()
}

/*
	The ToggleBrightnessButton toggles through different brightness values
	by adding the given step to the current brightness.	When the maximum or minimum
	brightness is reached, the step is inverted.

	The button shows the current brightness value as text.
*/

func NewToggleBrightnessButton(device hamdeck.Device, initialValue int, step int) *ToggleBrightnessButton {
	result := &ToggleBrightnessButton{
		device: device,
		step:   step,
	}
	result.setBrightness(initialValue)

	return result
}

type ToggleBrightnessButton struct {
	hamdeck.BaseButton
	device     hamdeck.Device
	image      image.Image
	brightness int
	step       int
}

func (b *ToggleBrightnessButton) setBrightness(brightness int) {
	if brightness >= 100 {
		b.brightness = 100
		b.step *= -1
	} else if brightness <= 0 {
		b.brightness = 0
		b.step *= -1
	} else {
		b.brightness = brightness
	}
}

func (b *ToggleBrightnessButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.image == nil {
		b.image = gc.DrawSingleLineTextButton(fmt.Sprintf("%d", b.brightness))
	}
	return b.image
}

func (b *ToggleBrightnessButton) Pressed() {
	b.setBrightness(b.brightness + b.step)
	b.device.SetBrightness(b.brightness)
	b.image = nil
	b.Invalidate()
}

func (b *ToggleBrightnessButton) Released() {
	// ignore
}

/*
	The PowerButton shows an icon with different colors, depending on the on/off state.
	The button calls a callback function when it is pressed.
*/

func NewPowerButton(callback Switch) *PowerButton {
	return &PowerButton{
		callback: callback,
	}
}

type PowerButton struct {
	hamdeck.BaseButton
	index    int
	icon     image.Image
	onImage  image.Image
	offImage image.Image
	on       bool
	callback Switch
}

type Switch func(on bool)

func (b *PowerButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.onImage == nil {
		b.onImage = b.drawPowerIcon(gc, hamdeck.Yellow, hamdeck.Black)
	}
	if b.offImage == nil {
		b.offImage = b.drawPowerIcon(gc, hamdeck.White, hamdeck.Black)
	}

	if b.on {
		return b.onImage
	}
	return b.offImage
}

func (b *PowerButton) drawPowerIcon(gc hamdeck.GraphicContext, foreground, background color.Color) image.Image {
	if b.icon == nil {
		b.icon = gc.LoadIconAsset("power.png")
	}

	gc.SetForeground(foreground)
	gc.SetBackground(background)

	return gc.DrawIconButton(b.icon)
}

func (b *PowerButton) Pressed() {
	b.on = !b.on
	if b.callback != nil {
		b.callback(b.on)
	}
	b.Invalidate()
}

func (b *PowerButton) Released() {
	// ignore
}

/*
	The CountingButton changes its text on external input and triggers a redraw of its content on the deck.
*/

func NewCountingButton() *CountingButton {
	result := &CountingButton{}

	return result
}

type CountingButton struct {
	hamdeck.BaseButton
	image    image.Image
	value    int
	Flashing bool
	flashOn  bool
}

func (b *CountingButton) Increase() {
	b.value += 1
	b.Invalidate()
}

func (b *CountingButton) Reset() {
	b.value = 0
	b.Invalidate()
}

func (b *CountingButton) Invalidate() {
	b.image = nil
	b.BaseButton.Invalidate()
}

func (b *CountingButton) Flash(flashOn bool) {
	if !b.Flashing {
		return
	}
	b.flashOn = flashOn
	b.Invalidate()
}

func (b *CountingButton) Image(gc hamdeck.GraphicContext) image.Image {
	if b.image == nil {
		gc.SetForeground(hamdeck.Red)
		if b.Flashing && b.flashOn {
			gc.SetBackground(hamdeck.Yellow)
		} else {
			gc.SetBackground(hamdeck.Black)
		}
		b.image = gc.DrawSingleLineTextButton(fmt.Sprintf("%d", b.value))
	}
	return b.image
}

func (b *CountingButton) Pressed() {
	b.Reset()
}

func (b *CountingButton) Released() {
	// ignore
}
