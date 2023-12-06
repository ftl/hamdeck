package hamdeck

import (
	"image"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHamDeckRoundtrip(t *testing.T) {
	device := newDefaultTestDevice()
	deck := New(device)
	deck.RegisterFactory(new(testButtonFactory))

	reader, err := openTestConfig("testRoundtrip")
	require.NoError(t, err)
	defer reader.Close()
	err = deck.ReadConfig(reader)
	require.NoError(t, err)

	require.Equal(t, 32, len(deck.buttons))
	button := deck.buttons[12].(*testButton)

	assert.Equal(t, "some_value", button.config["some_config"])
	assert.True(t, button.attached)
	assert.False(t, button.detached)
	assert.False(t, button.pressed)
	assert.False(t, button.released)

	stopper := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		deck.Run(stopper)
		wg.Done()
	}()

	device.Press(12)
	device.WaitForLastKey()
	assert.True(t, button.pressed)
	assert.False(t, button.released)

	device.Release(12)
	device.WaitForLastKey()
	assert.True(t, button.released)

	close(stopper)
	wg.Wait()
}

/* Test Harness */

func openTestConfig(name string) (io.ReadCloser, error) {
	return os.Open(filepath.Join("testdata", name+".json"))
}

type testDevice struct {
	id              string
	serial          string
	firmwareVersion string
	pixels          int
	rows            int
	columns         int
	keys            chan Key
}

func newDefaultTestDevice() *testDevice {
	return newTestDevice(128, 4, 8)
}

func newTestDevice(pixels int, rows int, columns int) *testDevice {
	return &testDevice{
		pixels:  pixels,
		rows:    rows,
		columns: columns,
		keys:    make(chan Key),
	}
}

func (d *testDevice) Close() error                    { return nil }
func (d *testDevice) ID() string                      { return d.id }
func (d *testDevice) Serial() string                  { return d.serial }
func (d *testDevice) FirmwareVersion() string         { return d.firmwareVersion }
func (d *testDevice) Pixels() int                     { return d.pixels }
func (d *testDevice) Rows() int                       { return d.rows }
func (d *testDevice) Columns() int                    { return d.columns }
func (d *testDevice) Clear() error                    { return nil }
func (d *testDevice) Reset() error                    { return nil }
func (d *testDevice) SetBrightness(int) error         { return nil }
func (d *testDevice) SetImage(int, image.Image) error { return nil }
func (d *testDevice) ReadKeys() (chan Key, error)     { return d.keys, nil }
func (d *testDevice) Press(index int) {
	key := Key{
		Index:   index,
		Pressed: true,
	}
	d.keys <- key
}
func (d *testDevice) Release(index int) {
	key := Key{
		Index:   index,
		Pressed: false,
	}
	d.keys <- key
}
func (d *testDevice) WaitForLastKey() {
	d.keys <- Key{}
}

const (
	testButtonType = "test.Button"
)

type testButtonFactory struct{}

func (f *testButtonFactory) Close() {}

func (f *testButtonFactory) CreateButton(config map[string]any) Button {
	switch config[ConfigType] {
	case testButtonType:
		return f.createTestButton(config)
	default:
		return nil
	}
}

func (f *testButtonFactory) createTestButton(config map[string]any) *testButton {
	return &testButton{
		config: config,
	}
}

type testButton struct {
	config map[string]any

	pressed  bool
	released bool
	attached bool
	detached bool
}

func (b *testButton) Image(GraphicContext, bool) image.Image { return nil }
func (b *testButton) Pressed()                               { b.pressed = true }
func (b *testButton) Released()                              { b.released = true }
func (b *testButton) Attached(ButtonContext)                 { b.attached = true }
func (b *testButton) Detached()                              { b.detached = true }
