module github.com/ftl/hamdeck

go 1.14

// replace github.com/ftl/rigproxy => ../rigproxy

// replace github.com/muesli/streamdeck => ../streamdeck
replace github.com/jfreymuth/pulse => ../pulse

require (
	github.com/fogleman/gg v1.3.0
	github.com/ftl/hamradio v0.0.0-20200721200456-334cc249f095
	github.com/ftl/rigproxy v0.0.0-20200804083623-21aea16acd5a
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jfreymuth/pulse v0.0.0-20200608153616-84b2d752b9d4
	github.com/muesli/streamdeck v0.0.0-20200514174954-dd59ecb861aa
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/image v0.0.0-20200801110659-972c09e46d76 // indirect
)
