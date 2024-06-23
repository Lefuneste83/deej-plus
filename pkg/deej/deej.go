package deej

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/omriharel/deej/pkg/deej/util"
)

const (
	// when this is set to anything, deej won't use a tray icon
	envNoTray = "DEEJ_NO_TRAY_ICON"
)

// Deej is the main entity managing access to all sub-components
type Deej struct {
	logger           *zap.SugaredLogger
	notifier         Notifier
	config           *CanonicalConfig
	sliderController SliderController
	sessions         *sessionMap

	stopChannel    chan bool
	version        string
	verbose        bool
	executablePath string
}

// NewDeej creates a Deej instance
func NewDeej(logger *zap.SugaredLogger, verbose bool) (*Deej, error) {
	logger = logger.Named("deej")

	notifier, err := NewToastNotifier(logger)
	if err != nil {
		logger.Errorw("Failed to create ToastNotifier", "error", err)
		return nil, fmt.Errorf("create new ToastNotifier: %w", err)
	}

	executablePath, err := os.Executable()
	if err != nil {
		logger.Errorw("Failed to get executable path", "error", err)
		return nil, fmt.Errorf("get executable path: %w", err)
	}

	config, err := NewConfig(logger, notifier, executablePath)
	if err != nil {
		logger.Errorw("Failed to create Config", "error", err)
		return nil, fmt.Errorf("create new Config: %w", err)
	}

	d := &Deej{
		logger:         logger,
		notifier:       notifier,
		config:         config,
		stopChannel:    make(chan bool),
		verbose:        verbose,
		executablePath: executablePath,
	}

	sessionFinder, err := newSessionFinder(logger)
	if err != nil {
		logger.Errorw("Failed to create SessionFinder", "error", err)
		return nil, fmt.Errorf("create new SessionFinder: %w", err)
	}

	sessions, err := newSessionMap(d, logger, sessionFinder)
	if err != nil {
		logger.Errorw("Failed to create sessionMap", "error", err)
		return nil, fmt.Errorf("create new sessionMap: %w", err)
	}

	d.sessions = sessions

	logger.Debug("Created deej instance")

	return d, nil
}

// Initialize sets up components and starts to run in the background
func (d *Deej) Initialize() error {
	d.logger.Debug("Initializing")

	// Find and load the config file
	configPath, err := findConfigFile(d.executablePath)
	if err != nil {
		d.logger.Errorw("Failed to find config file", "error", err)
		return fmt.Errorf("find config file: %w", err)
	}

	d.config.SetConfigFile(configPath)

	// load the config for the first time
	if err := d.config.Load(); err != nil {
		d.logger.Errorw("Failed to load config during initialization", "error", err)
		return fmt.Errorf("load config during init: %w", err)
	}

	if d.config.ControllerType == "serial" {
		serial, err := NewSerialIO(d, d.logger)
		if err != nil {
			d.logger.Errorw("Failed to create SerialIO", "error", err)
			return fmt.Errorf("create new SerialIO: %w", err)
		}

		d.sliderController = serial

		d.logger.Info("Created serial SliderController")
	} else if d.config.ControllerType == "udp" {
		udp, err := NewUdpIO(d, d.logger)
		if err != nil {
			d.logger.Errorw("Failed to create UdpIO", "error", err)
			return fmt.Errorf("create new UdpIO: %w", err)
		}

		d.sliderController = udp
		d.logger.Info("Created UDP SliderController")
	}

	// initialize the session map
	if err := d.sessions.initialize(); err != nil {
		d.logger.Errorw("Failed to initialize session map", "error", err)
		return fmt.Errorf("init session map: %w", err)
	}

	// decide whether to run with/without tray
	if _, noTraySet := os.LookupEnv(envNoTray); noTraySet {
		d.logger.Debugw("Running without tray icon", "reason", "envvar set")

		// run in main thread while waiting on ctrl+C
		d.setupInterruptHandler()
		d.run()
	} else {
		d.setupInterruptHandler()
		d.initializeTray(d.run)
	}

	return nil
}

// SetVersion causes deej to add a version string to its tray menu if called before Initialize
func (d *Deej) SetVersion(version string) {
	d.version = version
}

// Verbose returns a boolean indicating whether deej is running in verbose mode
func (d *Deej) Verbose() bool {
	return d.verbose
}

func (d *Deej) setupInterruptHandler() {
	interruptChannel := util.SetupCloseHandler()

	go func() {
		signal := <-interruptChannel
		d.logger.Debugw("Interrupted", "signal", signal)
		d.signalStop()
	}()
}

func (d *Deej) run() {
	d.logger.Info("Run loop starting")

	// watch the config file for changes
	go d.config.WatchConfigFileChanges()

	// connect to the arduino for the first time
	go func() {
		if err := d.sliderController.Start(); err != nil {
			d.logger.Warnw("Failed to start first-time slider connection", "error", err)

			if d.config.ControllerType == "serial" {
				// If the port is busy, that's because something else is connected - notify and quit
				if errors.Is(err, os.ErrPermission) {
					d.logger.Warnw("Serial port seems busy, notifying user and closing",
						"comPort", d.config.SerialConnectionInfo.COMPort)

					d.notifier.Notify(fmt.Sprintf("Can't connect to %s!", d.config.SerialConnectionInfo.COMPort),
						"This slider port is busy, make sure to close any slider monitor or other deej instance.")

					d.signalStop()

					// also notify if the COM port they gave isn't found, maybe their config is wrong
				} else if errors.Is(err, os.ErrNotExist) {
					d.logger.Warnw("Provided COM port seems wrong, notifying user and closing",
						"comPort", d.config.SerialConnectionInfo.COMPort)

					d.notifier.Notify(fmt.Sprintf("Can't connect to %s!", d.config.SerialConnectionInfo.COMPort),
						"This slider port doesn't exist, check your configuration and make sure it's set correctly.")

					d.signalStop()
				}
			} else if d.config.ControllerType == "udp" {
				d.notifier.Notify(fmt.Sprintf("Could not start UDP listener on port %d!", d.config.UdpConnectionInfo.UdpPort),
					"This UDP port is busy, make sure to close any slider monitor or other deej instance.")

				d.signalStop()
			}
		}
	}()

	// wait until stopped (gracefully)
	<-d.stopChannel
	d.logger.Debug("Stop channel signaled, terminating")

	if err := d.stop(); err != nil {
		d.logger.Warnw("Failed to stop deej", "error", err)
		os.Exit(1)
	} else {
		// exit with 0
		os.Exit(0)
	}
}

func (d *Deej) signalStop() {
	d.logger.Debug("Signalling stop channel")
	d.stopChannel <- true
}

func (d *Deej) stop() error {
	d.logger.Info("Stopping")

	d.config.StopWatchingConfigFile()
	d.sliderController.Stop()

	// release the session map
	if err := d.sessions.release(); err != nil {
		d.logger.Errorw("Failed to release session map", "error", err)
		return fmt.Errorf("release session map: %w", err)
	}

	d.stopTray()

	// attempt to sync on exit - this won't necessarily work but can't harm
	d.logger.Sync()

	return nil
}
