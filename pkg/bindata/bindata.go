// The package bindata embeds all the needed binary data, like images and fonts.
package bindata

import (
	"embed"
)

//go:embed img/* fonts/*
var Assets embed.FS
