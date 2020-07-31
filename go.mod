module github.com/ftl/hamdeck

go 1.14

// replace github.com/ftl/rigproxy => ../rigproxy
// replace github.com/muesli/streamdeck => ../streamdeck
replace github.com/jfreymuth/pulse => ../pulse

require (
	github.com/fogleman/gg v1.3.0
	github.com/ftl/hamradio v0.0.0-20200721200456-334cc249f095
	github.com/ftl/rigproxy v0.0.0-20200524134605-8e6f179b3a88
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jfreymuth/pulse v0.0.0-20200608153616-84b2d752b9d4
	github.com/muesli/streamdeck v0.0.0-20200514174954-dd59ecb861aa
	github.com/spf13/cobra v1.0.0
	golang.org/x/image v0.0.0-20200618115811-c13761719519 // indirect
)
