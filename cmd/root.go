package cmd

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ftl/hamradio/cfg"
	"github.com/spf13/cobra"

	"github.com/ftl/hamdeck/pkg/hamdeck"
	"github.com/ftl/hamdeck/pkg/hamlib"
	"github.com/ftl/hamdeck/pkg/mqtt"
	"github.com/ftl/hamdeck/pkg/pulse"
	"github.com/ftl/hamdeck/pkg/streamdeck"
	"github.com/ftl/hamdeck/pkg/tci"
)

var (
	version   string = "development"
	gitCommit string = "unknown"
	buildTime string = "unknown"
)

var rootFlags = struct {
	syslog        bool
	serial        string
	brightness    int
	configFile    string
	hamlibAddress string
	tciAddress    string
	mqttAddress   string
	mqttUsername  string
	mqttPassword  string
}{}

var rootCmd = &cobra.Command{
	Use:     "hamdeck",
	Short:   "HamDeck Version " + version + " - control your ham radio station with an Elgato Stream Deck",
	Version: version,
	Run:     run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&rootFlags.syslog, "syslog", false, "use syslog for logging")
	rootCmd.PersistentFlags().StringVar(&rootFlags.serial, "serial", "", "the serial number of the Stream Deck device that should be used")
	rootCmd.PersistentFlags().IntVar(&rootFlags.brightness, "brightness", 100, "the initial brightness of the Stream Deck device")
	rootCmd.PersistentFlags().StringVar(&rootFlags.configFile, "config", "", "the configuration file that should be used (default: .config/hamradio/hamdeck.json)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.hamlibAddress, "hamlib", "", "the address of the rigctld server (if empty, hamlib buttons are not available)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.tciAddress, "tci", "", "the address of the TCI server (if empty, tci buttons are not available)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.mqttAddress, "mqtt", "", "the address of the MQTT server (if empty, atu100 buttons are not available)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.mqttUsername, "mqttusername", "", "the username for MQTT")
	rootCmd.PersistentFlags().StringVar(&rootFlags.mqttPassword, "mqttpassword", "", "the password for MQTT")
}

func run(cmd *cobra.Command, args []string) {
	log.Printf("Hamdeck Version %s", version)
	shutdown := monitorShutdownSignals()

	if rootFlags.syslog {
		logger, err := syslog.NewLogger(syslog.LOG_INFO, 0)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logger.Writer())
	}

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
	if rootFlags.hamlibAddress != "" {
		deck.RegisterFactory(hamlib.NewButtonFactory(rootFlags.hamlibAddress))
	}
	if rootFlags.tciAddress != "" {
		deck.RegisterFactory(tci.NewButtonFactory(rootFlags.tciAddress))
	}
	if rootFlags.mqttAddress != "" {
		deck.RegisterFactory(mqtt.NewButtonFactory(rootFlags.mqttAddress, rootFlags.mqttUsername, rootFlags.mqttPassword))
	}

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
		configDirectory, err := cfg.Directory("")
		if err != nil {
			return fmt.Errorf("cannot resolve configuration directory: %w", err)
		}
		config = filepath.Join(configDirectory, hamdeck.ConfigDefaultFilename)
		log.Printf("Using default configuration file %s", config)
	}

	file, err := os.Open(config)
	if err != nil {
		return fmt.Errorf("cannot open configuration file: %w", err)
	}
	defer file.Close()

	err = deck.ReadConfig(file)
	if err != nil {
		return err
	}

	deck.CloseUnusedFactories()

	return nil
}
