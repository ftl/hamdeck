package hamdeck

import "log"

const (
	ConfigPage  = "page"
	ConfigLabel = "label"
)

const (
	PageButtonType = "hamdeck.Page"
)

type Factory struct {
	pageSwitcher PageSwitcher
}

type PageSwitcher interface {
	AttachPage(string) error
}

func NewButtonFactory(pageSwitcher PageSwitcher) *Factory {
	return &Factory{
		pageSwitcher: pageSwitcher,
	}
}

func (f *Factory) Close() {
	// nop
}

func (f *Factory) CreateButton(config map[string]any) Button {
	switch config[ConfigType] {
	case PageButtonType:
		return f.createPageButton(config)
	default:
		return nil
	}
}

func (f *Factory) createPageButton(config map[string]any) Button {
	id, haveID := ToString(config[ConfigPage])
	label, haveLabel := ToString(config[ConfigLabel])
	if !haveID {
		log.Print("A hamdeck.Page button must have a page field.")
	}
	if !haveLabel {
		log.Print("A hamdeck.Page button must have a label field.")
	}
	return NewPageButton(f.pageSwitcher, id, label)
}
