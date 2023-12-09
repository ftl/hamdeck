package hamdeck

import "image"

/*
	PageButton
*/

type PageButton struct {
	BaseButton
	image image.Image

	pageSwitcher PageSwitcher
	id           string
	label        string
}

func NewPageButton(pageSwitcher PageSwitcher, id string, label string) *PageButton {
	return &PageButton{
		pageSwitcher: pageSwitcher,
		id:           id,
		label:        label,
	}
}

func (b *PageButton) Image(gc GraphicContext, redrawImages bool) image.Image {
	if b.image == nil || redrawImages {
		gc.SetForeground(White)
		gc.SetBackground(Black)
		b.image = gc.DrawSingleLineTextButton(b.label)
	}
	return b.image
}

func (b *PageButton) Pressed() {
	b.pageSwitcher.AttachPage(b.id)
}

func (b *PageButton) Released() {
	// nop
}
