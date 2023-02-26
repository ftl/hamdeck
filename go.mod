module github.com/ftl/hamdeck

go 1.19

// replace github.com/ftl/tci => ../tci

// replace github.com/ftl/rigproxy => ../rigproxy

// replace github.com/muesli/streamdeck => ../streamdeck
// replace github.com/jfreymuth/pulse => ../pulse

require (
	github.com/eclipse/paho.mqtt.golang v1.4.2
	github.com/fogleman/gg v1.3.0
	github.com/ftl/hamradio v0.2.6
	github.com/ftl/rigproxy v0.2.0
	github.com/ftl/tci v0.3.1
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/jfreymuth/pulse v0.1.0
	github.com/muesli/streamdeck v0.3.0
	github.com/spf13/cobra v1.6.1
	golang.org/x/image v0.5.0
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/karalabe/hid v1.0.1-0.20190806082151-9c14560f9ee8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
)
