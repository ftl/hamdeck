module github.com/ftl/hamdeck

go 1.16

// replace github.com/ftl/tci => ../tci

// replace github.com/ftl/rigproxy => ../rigproxy

// replace github.com/muesli/streamdeck => ../streamdeck
// replace github.com/jfreymuth/pulse => ../pulse

require (
	github.com/eclipse/paho.mqtt.golang v1.3.1
	github.com/fogleman/gg v1.3.0
	github.com/ftl/hamradio v0.0.0-20200721200456-334cc249f095
	github.com/ftl/rigproxy v0.0.0-20200812132905-1b8d78e5c89e
	github.com/ftl/tci v0.0.0-20210131212252-75860f67cedb
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/jfreymuth/pulse v0.0.0-20200804114219-7d61c4938214
	github.com/muesli/streamdeck v0.0.0-20200514174954-dd59ecb861aa
	github.com/spf13/cobra v1.1.1
	golang.org/x/image v0.0.0-20200801110659-972c09e46d76
)
