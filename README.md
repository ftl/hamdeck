# HamDeck

HamDeck allows you to control and automate your ham radio station using an Elgato Stream Deck device. You can define buttons using a JSON configuration file. Currently the following actions are implemented:

* Toggle the mute state of a pulseaudio sink our source.

The tool is written Go on Linux. It might also work on OSX or Windows, but I did not try that out.

## Build

Binary data, e.g. icons, are stored in the sub-directories of `pkg/bindata`. All files are embedded using `go-bindata`. If you make any changes, you need to execute

```
go generate ./...
```

To build the `hamdeck` binary simply run

```
go build
```

## Disclaimer

I develop this tools for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome).

## Links

* [Wiki](https://github.com/ftl/hamradio/wiki)

## License

This tool is published under the [MIT License](https://www.tldrlegal.com/l/mit).

This repository and also the binary contains imaged from [https://material.io](https://material.io/resources/icons/), which are licensed under the [Apache license version 2.0](https://www.apache.org/licenses/LICENSE-2.0.html).

Copyright [Florian Thienel](http://thecodingflow.com/)
