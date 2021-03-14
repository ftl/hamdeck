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
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

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
	DarkGreen    = color.RGBA{0, 150, 0, 255}
	Orange       = color.RGBA{255, 165, 0, 255}
)

var (
	DefaultBackground = Black
	DefaultForeground = White
	DefaultFont       = "DejaVuSans.ttf"
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
	fontName   string
	fontSize   float64
}

func (gc *GC) Pixels() int {
	return gc.pixels
}

func (gc *GC) Reset() {
	gc.background = DefaultBackground
	gc.foreground = DefaultForeground
	gc.fontName = DefaultFont
	gc.fontSize = DefaultFontSize
}

func (gc *GC) SetBackground(background color.Color) {
	gc.background = background
}

func (gc *GC) SetForeground(foreground color.Color) {
	gc.foreground = foreground
}

func (gc *GC) SwapColors() {
	temp := gc.foreground
	gc.foreground = gc.background
	gc.background = temp
}

func (gc *GC) SetFont(filename string) {
	gc.fontName = filename
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

	ctx.SetFontFace(gc.LoadFontAsset(gc.fontName, gc.fontSize))
	ctx.SetColor(gc.background)
	ctx.Clear()
	ctx.SetColor(gc.foreground)
	ctx.DrawStringAnchored(text, float64(gc.pixels)/2, float64(gc.pixels)/2, 0.5, 0.5)

	return result
}

func (gc *GC) DrawDoubleLineToggleTextButton(text1, text2 string, activeLine int) image.Image {
	result, ctx := gc.newImage()

	bigSize := gc.fontSize
	smallSize := 0.75 * bigSize

	ctx.SetColor(gc.background)
	ctx.Clear()
	ctx.SetColor(gc.foreground)

	fontSize := bigSize
	if activeLine != 1 {
		fontSize = smallSize
	}
	ctx.SetFontFace(gc.LoadFontAsset(gc.fontName, fontSize))
	ctx.DrawStringAnchored(text1, 0.5*float64(gc.pixels), 0.25*float64(gc.pixels), 0.5, 0.5)

	fontSize = bigSize
	if activeLine != 2 {
		fontSize = smallSize
	}
	ctx.SetFontFace(gc.LoadFontAsset(gc.fontName, fontSize))
	ctx.DrawStringAnchored(text2, 0.5*float64(gc.pixels), 0.75*float64(gc.pixels), 0.5, 0.5)

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
	asset, err := bindata.Assets.Open(assetName)
	if err != nil {
		log.Fatalf("cannot open asset %s: %v", assetName, err)
	}
	defer asset.Close()

	icon, err := gc.LoadIconFromReader(asset)
	if err != nil {
		log.Fatalf("cannot load asset %s: %v", assetName, err)
	}
	return icon
}

func (gc *GC) LoadFontFaceFromReader(r io.Reader, points float64) (font.Face, error) {
	fontBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: points,
	})
	return face, nil
}

func (gc *GC) LoadFontAsset(name string, points float64) font.Face {
	assetName := fmt.Sprintf("fonts/%s", name)
	asset, err := bindata.Assets.Open(assetName)
	if err != nil {
		log.Fatalf("cannot open asset %s: %v", assetName, err)
	}
	defer asset.Close()

	face, err := gc.LoadFontFaceFromReader(asset, points)
	if err != nil {
		log.Fatalf("cannot load asset %s: %v", assetName, err)
	}
	return face
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

	ctx.SetFontFace(gc.LoadFontAsset(gc.fontName, gc.fontSize))
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
