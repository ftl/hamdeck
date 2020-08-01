package hamdeck

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"os"

	"github.com/fogleman/gg"

	"github.com/ftl/hamdeck/pkg/bindata"
)

var (
	Black        = color.Black
	White        = color.White
	DisabledGray = color.RGBA{54, 60, 62, 255}
	Red          = color.RGBA{255, 0, 0, 255}
	Green        = color.RGBA{0, 255, 0, 255}
	Blue         = color.RGBA{0, 0, 255, 255}
	Yellow       = color.RGBA{255, 255, 0, 255}
	Magenta      = color.RGBA{255, 0, 255, 255}
	Cyan         = color.RGBA{0, 255, 255, 255}
)

var (
	DefaultBackground = Black
	DefaultForeground = White
	DefaultFont       = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	DefaultFontSize   = 24.0
)

func NewGraphicContext(pixels int) GraphicContext {
	result := &GC{
		pixels: pixels,
	}
	result.Reset()
	return result
}

type GC struct {
	pixels     int
	background color.Color
	foreground color.Color
	fontFile   string
	fontSize   float64
}

func (gc *GC) Pixels() int {
	return gc.pixels
}

func (gc *GC) Reset() {
	gc.background = DefaultBackground
	gc.foreground = DefaultForeground
	gc.fontFile = DefaultFont
	gc.fontSize = DefaultFontSize
}

func (gc *GC) SetBackground(background color.Color) {
	gc.background = background
}

func (gc *GC) SetForeground(foreground color.Color) {
	gc.foreground = foreground
}

func (gc *GC) SetFont(filename string) {
	gc.fontFile = filename
}

func (gc *GC) SetFontSize(points float64) {
	gc.fontSize = points
}

func (gc *GC) newImage() (*image.RGBA, *gg.Context) {
	result := image.NewRGBA(image.Rect(0, 0, gc.pixels, gc.pixels))
	ctx := gg.NewContextForRGBA(result)
	return result, ctx
}

func (gc *GC) DrawNoButton() image.Image {
	result, ctx := gc.newImage()
	ctx.SetColor(Black)
	ctx.Clear()
	return result
}

func (gc *GC) DrawSingleLineTextButton(text string) image.Image {
	result, ctx := gc.newImage()

	err := ctx.LoadFontFace(gc.fontFile, gc.fontSize)
	if err != nil {
		log.Print(err)
		return gc.DrawNoButton()
	}

	ctx.SetColor(gc.background)
	ctx.Clear()
	ctx.SetColor(gc.foreground)
	ctx.DrawStringAnchored(text, float64(gc.pixels)/2, float64(gc.pixels)/2, 0.5, 0.5)

	return result
}

func (gc *GC) LoadIconFromFile(filename string) (image.Image, error) {
	iconFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open icon: %v", err)
	}
	defer iconFile.Close()
	return gc.LoadIconFromReader(iconFile)
}

func (gc *GC) LoadIconFromReader(r io.Reader) (image.Image, error) {
	icon, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("cannot decode icon: %v", err)
	}
	return icon, nil
}

func (gc *GC) LoadIconAsset(name string) image.Image {
	assetName := fmt.Sprintf("img/%s", name)
	icon, err := gc.LoadIconFromReader(bindata.AssetReader(assetName))
	if err != nil {
		log.Fatalf("cannot load asset %s: %v", assetName, err)
	}
	return icon
}

func (gc *GC) DrawIconButton(icon image.Image) image.Image {
	result, ctx := gc.newImage()

	if icon.Bounds().Dx() != gc.pixels || icon.Bounds().Dy() != gc.pixels {
		iconPixels := int(math.Max(float64(icon.Bounds().Dx()), float64(icon.Bounds().Dy())))
		iconDX := (iconPixels - icon.Bounds().Dx()) / 2
		iconDY := (iconPixels - icon.Bounds().Dy()) / 2
		iconBounds := image.Rect(iconDX, iconDY, icon.Bounds().Dx()+iconDX, icon.Bounds().Dy()+iconDY)
		iconScaling := float64(gc.pixels) / float64(iconPixels)

		img := image.NewRGBA(image.Rect(0, 0, iconPixels, iconPixels))
		draw.Draw(img, img.Bounds(), image.NewUniform(gc.background), image.ZP, draw.Src)
		draw.DrawMask(img, iconBounds, image.NewUniform(gc.foreground), image.ZP, icon, image.ZP, draw.Over)

		ctx.ScaleAbout(iconScaling, iconScaling, float64(gc.pixels)/2, float64(gc.pixels)/2)
		ctx.DrawImageAnchored(img, gc.pixels/2, gc.pixels/2, 0.5, 0.5)
	} else {
		draw.Draw(result, result.Bounds(), image.NewUniform(gc.background), image.ZP, draw.Src)
		draw.DrawMask(result, result.Bounds(), image.NewUniform(gc.foreground), image.ZP, icon, image.ZP, draw.Over)
	}

	return result
}

func (gc *GC) DrawIconLabelButton(icon image.Image, label string) image.Image {
	result, ctx := gc.newImage()

	err := ctx.LoadFontFace(gc.fontFile, gc.fontSize)
	if err != nil {
		log.Print(err)
		return gc.DrawNoButton()
	}

	ctx.SetColor(gc.background)
	ctx.Clear()
	ctx.SetColor(gc.foreground)
	ctx.DrawStringAnchored(label, 0.5*float64(gc.pixels), 0.75*float64(gc.pixels), 0.5, 0.5)

	iconPixels := int(math.Max(float64(icon.Bounds().Dx()), float64(icon.Bounds().Dy())))
	iconDX := (iconPixels - icon.Bounds().Dx()) / 2
	iconDY := (iconPixels - icon.Bounds().Dy()) / 2
	iconBounds := image.Rect(iconDX, iconDY, icon.Bounds().Dx()+iconDX, icon.Bounds().Dy()+iconDY)
	iconScaledHeight := float64(ctx.Height()) / 2
	iconScaling := iconScaledHeight / float64(iconPixels)

	img := image.NewRGBA(image.Rect(0, 0, iconPixels, iconPixels))
	draw.Draw(img, img.Bounds(), image.NewUniform(gc.background), image.ZP, draw.Src)
	draw.DrawMask(img, iconBounds, image.NewUniform(gc.foreground), image.ZP, icon, image.ZP, draw.Over)

	ctx.ScaleAbout(iconScaling, iconScaling, float64(ctx.Width())/2, iconScaledHeight/2)
	ctx.DrawImageAnchored(img, ctx.Width()/2, int(iconScaledHeight)/2, 0.5, 0.5)

	return result
}
