package streamdeck

import (
	"fmt"
	"image"
	"log"

	"github.com/muesli/streamdeck"

	"github.com/ftl/hamdeck/pkg/hamdeck"
)

func Open(serial string) (*Device, error) {
	devices, err := streamdeck.Devices()
	if err != nil {
		return nil, fmt.Errorf("cannot enumerate the Stream Deck devices: %w", err)
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no Stream Deck devices found")
	}
	log.Printf("Found %d Stream Deck devices", len(devices))

	device := devices[0]
	if serial != "" {
		found := false
		for _, d := range devices {
			if d.Serial == serial {
				device = d
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("no device found with serial %s", serial)
		}
	}

	err = device.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot open Stream Deck: %w", err)
	}

	err = device.Reset()
	if err != nil {
		return nil, fmt.Errorf("cannot reset Stream Deck: %w", err)
	}

	firmwareVersion, err := device.FirmwareVersion()
	if err != nil {
		log.Printf("Cannot read firmware version from Stream Deck with serial %v: %v", device.Serial, err)
		firmwareVersion = "n/a"
	}

	return &Device{
		device:          &device,
		firmwareVersion: firmwareVersion,
	}, nil
}

type Device struct {
	device          *streamdeck.Device
	firmwareVersion string
}

func (d *Device) Close() error {
	return d.device.Close()
}

func (d *Device) ID() string {
	return d.device.ID
}

func (d *Device) Serial() string {
	return d.device.Serial
}

func (d *Device) FirmwareVersion() string {
	return d.firmwareVersion
}

func (d *Device) Pixels() int {
	return int(d.device.Pixels)
}

func (d *Device) Rows() int {
	return int(d.device.Rows)

}

func (d *Device) Columns() int {
	return int(d.device.Columns)
}

func (d *Device) Clear() error {
	return d.device.Clear()
}

func (d *Device) Reset() error {
	return d.device.Reset()
}

func (d *Device) SetBrightness(brightness int) error {
	return d.device.SetBrightness(uint8(brightness))
}

func (d *Device) SetImage(button int, image image.Image) error {
	return d.device.SetImage(uint8(button), image)
}

func (d *Device) ReadKeys() (chan hamdeck.Key, error) {
	in, err := d.device.ReadKeys()
	if err != nil {
		return nil, fmt.Errorf("cannot get source key channel: %w", err)
	}

	out := make(chan hamdeck.Key, 1)
	go func() {
		for {
			select {
			case key, ok := <-in:
				if !ok {
					close(out)
					return
				}
				out <- hamdeck.Key{Index: int(key.Index), Pressed: key.Pressed}
			}
		}
	}()
	return out, nil
}
