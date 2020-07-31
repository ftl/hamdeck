package bindata

import (
	"bytes"
	"io"
)

func AssetReader(name string) io.Reader {
	return bytes.NewBuffer(MustAsset(name))
}
