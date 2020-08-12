# HamDeck

HamDeck allows you to control and automate your ham radio station using an Elgato Stream Deck device. You can define buttons using a JSON configuration file. HamDeck connects to the local pulseaudio server and the local rigctld server on the default ports. Currently the following actions are implemented as Stream Deck buttons:

* Toggle the mute state of a pulseaudio sink our source.
* Call any simple hamlib set command (e.g. `vfo_op BAND_UP`).
* Set the mode of your radio through hamlib.
* Switch to a specific frequency band through hamlib (if `vfo_op BAND_UP/BAND_DOWN` is supported by your radio).
* Set the output power level of your radio through hamlib.
* Control the TX state (MOX) of your radio through hamlib.
* Select the VFO of your radio through hamlib.
* Get an indication on the mode buttons which modes are suitable to the current frequency according to the IARU Region 1 bandplan.
* Jump to the beginning of the closest band portion suitable for your currently selected mode (press the mode button > 1s).

This tool is written in Go on Linux. It might also work on OSX or Windows, but I did not try that out.

## Build

Binary data, e.g. icons, are stored in the sub-directories of `pkg/bindata`. All files are embedded using `go-bindata`. If you make any changes, you need to execute

```
go generate ./...
```

To build the `hamdeck` binary simply run

```
go build
```

## Install

The following describes the steps how to install `hamdeck` on an Ubuntu 20.04 LTS (Focal Fossa) to start automatically when you plug-in your the Stream Deck device.

1. Installing the binary:

```
go install github.com/ftl/hamdeck
which hamdeck
```

The last command will give you the path to the installed binary, which you will need later.

2. Adding UDEV rules for the stream deck

Create `/etc/udev/rules.d/99-streamdeck.rules` with the following content:

```
ACTION=="add", SUBSYSTEM=="usb", ATTRS{idVendor}=="0fd9", ATTRS{idProduct}=="0060", MODE:="666", GROUP="plugdev", SYMLINK="streamdeck", TAG+="systemd", ENV{SYSTEMD_WANTS}="hamdeck.service"
ACTION=="add", SUBSYSTEM=="usb", ATTRS{idVendor}=="0fd9", ATTRS{idProduct}=="0063", MODE:="666", GROUP="plugdev", SYMLINK="streamdeck", TAG+="systemd", ENV{SYSTEMD_WANTS}="hamdeck.service"
ACTION=="add", SUBSYSTEM=="usb", ATTRS{idVendor}=="0fd9", ATTRS{idProduct}=="006c", MODE:="666", GROUP="plugdev", SYMLINK="streamdeck", TAG+="systemd", ENV{SYSTEMD_WANTS}="hamdeck.service"
ACTION=="add", SUBSYSTEM=="usb", ATTRS{idVendor}=="0fd9", ATTRS{idProduct}=="006d", MODE:="666", GROUP="plugdev", SYMLINK="streamdeck", TAG+="systemd", ENV{SYSTEMD_WANTS}="hamdeck.service"

ACTION=="remove", SUBSYSTEM=="usb", ENV{PRODUCT}=="fd9/60/*", TAG+="systemd"
ACTION=="remove", SUBSYSTEM=="usb", ENV{PRODUCT}=="fd9/63/*", TAG+="systemd"
ACTION=="remove", SUBSYSTEM=="usb", ENV{PRODUCT}=="fd9/6c/*", TAG+="systemd"
ACTION=="remove", SUBSYSTEM=="usb", ENV{PRODUCT}=="fd9/6d/*", TAG+="systemd"
```

(See [systemd devices](https://www.freedesktop.org/software/systemd/man/systemd.device.html) and [systemd #7587](https://github.com/systemd/systemd/issues/7587) for a little bit of background about the UDEV rules.)

3. Adding a service definition for `hamdeck`

Create `/etc/systemd/system/hamdeck.service` with the following content:

```
[Unit]
Description=HamDeck
After=syslog.target dev-streamdeck.device
BindsTo=dev-streamdeck.device

[Service]
ExecStart=<hamdeck binary, see above> --syslog --config=<your config file>
User=<your username>
```

## Disclaimer

I develop this tools for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome).

## Links

* [Wiki](https://github.com/ftl/hamdeck/wiki)

## License

This tool is published under the [MIT License](https://www.tldrlegal.com/l/mit).

This repository and also the binary contains images from [https://material.io](https://material.io/resources/icons/), which are licensed under the [Apache license version 2.0](https://www.apache.org/licenses/LICENSE-2.0.html).

Copyright [Florian Thienel](http://thecodingflow.com/)
