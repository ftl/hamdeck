package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ftl/hamradio/cfg"
	"github.com/spf13/cobra"

	"github.com/ftl/hamdeck/pkg/hamdeck"
	"github.com/ftl/hamdeck/pkg/hamlib"
	"github.com/ftl/hamdeck/pkg/pulse"
	"github.com/ftl/hamdeck/pkg/streamdeck"
)

var rootFlags = struct {
	serial        string
	brightness    int
	configFile    string
	hamlibAddress string
}{}

var rootCmd = &cobra.Command{
	Use:   "hamdeck",
	Short: "HamDeck - control your ham radio station with a Stream Deck",
	Run:   run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootFlags.serial, "serial", "", "the serial number of the Stream Deck device that should be used")
	rootCmd.PersistentFlags().IntVar(&rootFlags.brightness, "brightness", 100, "the initial brightness of the Stream Deck device")
	rootCmd.PersistentFlags().StringVar(&rootFlags.configFile, "config", "", "the configuration file that should be used (default: .config/hamradio/conf.json)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.hamlibAddress, "hamlib", "", "the address of the rigctld server (default: localhost:4532)")
}

func run(cmd *cobra.Command, args []string) {
	shutdown := monitorShutdownSignals()

	device, err := streamdeck.Open(rootFlags.serial)
	if err != nil {
		log.Fatalf("Cannot open Stream Deck: %v", err)
	}
	defer func() {
		log.Print("Closing device")
		err := device.Close()
		if err != nil {
			log.Printf("Cannot close Stream Deck: %v", err)
		} else {
			log.Print("Device closed")
		}
	}()

	log.Printf("Using Stream Deck %s %dx%d %s", device.ID(), device.Columns(), device.Rows(), device.Serial())
	log.Printf("Firmware Version %s", device.FirmwareVersion())
	device.SetBrightness(rootFlags.brightness)

	deck := hamdeck.New(device)
	deck.RegisterFactory(pulse.NewButtonFactory())
	deck.RegisterFactory(hamlib.NewButtonFactory(rootFlags.hamlibAddress))

	err = configureHamDeck(deck, rootFlags.configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = deck.Run(shutdown)
	if err != nil {
		log.Fatal(err)
	}
}

func monitorShutdownSignals() <-chan struct{} {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	shutdown := make(chan struct{})
	go func() {
		rude := false
		for {
			<-signals
			if rude {
				log.Fatal("graceful shutdown failed")
			}
			rude = true
			close(shutdown)
		}
	}()
	return shutdown
}

func configureHamDeck(deck *hamdeck.HamDeck, config string) error {
	if config == "" {
		return useDefaultConfiguration(deck)
	}

	file, err := os.Open(config)
	if err != nil {
		return fmt.Errorf("cannot open configuration file: %w", err)
	}
	defer file.Close()

	return deck.ReadConfig(file)
}

func useDefaultConfiguration(deck *hamdeck.HamDeck) error {
	hamradioConfig, err := cfg.LoadDefault()
	if err != nil {
		return fmt.Errorf("cannot open default configuration: %w", err)
	}

	rawConfig := hamradioConfig.Get(hamdeck.ConfigMainKey, nil)
	config, ok := rawConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot find hamdeck configuration in default configuration file")
	}

	return deck.AttachConfiguredButtons(config)
}
