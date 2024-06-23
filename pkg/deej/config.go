package deej

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/omriharel/deej/pkg/deej/util"
)

// CanonicalConfig provides application-wide access to configuration fields,
// as well as loading/file watching logic for deej's configuration file
type CanonicalConfig struct {
	SliderMapping  *sliderMap
	buttonMapping  *buttonMap
	ControllerType string

	SerialConnectionInfo struct {
		COMPort  string
		BaudRate int
	}

	UdpConnectionInfo struct {
		UdpPort int
	}

	InvertSliders       bool
	NoiseReductionLevel string

	logger             *zap.SugaredLogger
	notifier           Notifier
	stopWatcherChannel chan bool

	reloadConsumers []chan bool

	userConfig     *viper.Viper
	internalConfig *viper.Viper
}

const (
	userConfigFilepath     = "config.yaml"
	internalConfigFilepath = "preferences.yaml"

	userConfigName     = "config"
	internalConfigName = "preferences"

	userConfigPath = "."

	configType = "yaml"

	configKeySliderMapping       = "slider_mapping"
	configKeyButtonMapping       = "button_mapping"
	configKeyInvertSliders       = "invert_sliders"
	configKeyCOMPort             = "com_port"
	configKeyBaudRate            = "baud_rate"
	configKeyNoiseReductionLevel = "noise_reduction"
	configKeyUdpPort             = "udp_port"
	configKeyControllerType      = "controller_type"

	defaultControllerType = "serial"

	defaultCOMPort  = "COM4"
	defaultBaudRate = 9600

	defaultUdpPort = 16990
)

// has to be defined as a non-constant because we're using path.Join
var internalConfigPath = filepath.Join(".", logDirectory)

// NewConfig creates a config instance for the deej object and sets up viper instances for deej's config files
func NewConfig(logger *zap.SugaredLogger, notifier Notifier, executablePath string) (*CanonicalConfig, error) {
	logger = logger.Named("config")

	cc := &CanonicalConfig{
		logger:             logger,
		notifier:           notifier,
		reloadConsumers:    []chan bool{},
		stopWatcherChannel: make(chan bool),
	}

	// distinguish between the user-provided config (config.yaml) and the internal config (logs/preferences.yaml)
	userConfig := viper.New()
	userConfig.SetConfigName(userConfigName)
	userConfig.SetConfigType(configType)

	// Find the config file
	configPath, err := findConfigFile(executablePath)
	if err != nil {
		return nil, fmt.Errorf("find config file: %w", err)
	}
	userConfig.SetConfigFile(configPath)

	userConfig.SetDefault(configKeySliderMapping, map[string][]string{})
	userConfig.SetDefault(configKeyButtonMapping, map[string][]string{})
	userConfig.SetDefault(configKeyInvertSliders, false)
	userConfig.SetDefault(configKeyCOMPort, defaultCOMPort)
	userConfig.SetDefault(configKeyBaudRate, defaultBaudRate)
	userConfig.SetDefault(configKeyUdpPort, defaultUdpPort)
	userConfig.SetDefault(configKeyControllerType, defaultControllerType)

	internalConfig := viper.New()
	internalConfig.SetConfigName(internalConfigName)
	internalConfig.SetConfigType(configType)
	internalConfig.AddConfigPath(internalConfigPath)

	cc.userConfig = userConfig
	cc.internalConfig = internalConfig

	logger.Debug("Created config instance")

	return cc, nil
}

// Load reads deej's config files from disk and tries to parse them
func (cc *CanonicalConfig) Load() error {
	cc.logger.Debugw("Loading config", "path", cc.userConfig.ConfigFileUsed())

	// make sure it exists
	if !util.FileExists(cc.userConfig.ConfigFileUsed()) {
		cc.logger.Warnw("Config file not found", "path", cc.userConfig.ConfigFileUsed())
		cc.notifier.Notify("Can't find configuration!",
			fmt.Sprintf("%s must be in one of the valid locations. Please re-launch", userConfigFilepath))

		return fmt.Errorf("config file doesn't exist: %s", cc.userConfig.ConfigFileUsed())
	}

	// load the user config
	if err := cc.userConfig.ReadInConfig(); err != nil {
		cc.logger.Warnw("Viper failed to read user config", "error", err)

		// if the error is yaml-format-related, show a sensible error. otherwise, show 'em to the logs
		if strings.Contains(err.Error(), "yaml:") {
			cc.notifier.Notify("Invalid configuration!",
				fmt.Sprintf("Please make sure %s is in a valid YAML format.", userConfigFilepath))
		} else {
			cc.notifier.Notify("Error loading configuration!", "Please check deej's logs for more details.")
		}

		return fmt.Errorf("read user config: %w", err)
	}

	// load the internal config - this doesn't have to exist, so it can error
	if err := cc.internalConfig.ReadInConfig(); err != nil {
		cc.logger.Debugw("Viper failed to read internal config", "error", err, "reminder", "this is fine")
	}

	// canonize the configuration with viper's helpers
	if err := cc.populateFromVipers(); err != nil {
		cc.logger.Warnw("Failed to populate config fields", "error", err)
		return fmt.Errorf("populate config fields: %w", err)
	}

	cc.logger.Info("Loaded config successfully")
	cc.logger.Infow("Config values",
		"sliderMapping", cc.SliderMapping,
		"buttonMapping", cc.buttonMapping,
		"controllerType", cc.ControllerType,
		"serialConnectionInfo", cc.SerialConnectionInfo,
		"udpConnectionInfo", cc.UdpConnectionInfo,
		"invertSliders", cc.InvertSliders)

	return nil
}

// SubscribeToChanges allows external components to receive updates when the config is reloaded
func (cc *CanonicalConfig) SubscribeToChanges() chan bool {
	c := make(chan bool)
	cc.reloadConsumers = append(cc.reloadConsumers, c)
	return c
}

// WatchConfigFileChanges starts watching for configuration file changes
// and attempts reloading the config when they happen
func (cc *CanonicalConfig) WatchConfigFileChanges() {
	cc.logger.Debugw("Starting to watch user config file for changes", "path", cc.userConfig.ConfigFileUsed())

	const (
		minTimeBetweenReloadAttempts = time.Millisecond * 500
		delayBetweenEventAndReload   = time.Millisecond * 50
	)

	lastAttemptedReload := time.Now()

	cc.userConfig.WatchConfig()
	cc.userConfig.OnConfigChange(func(event fsnotify.Event) {
		if event.Op&fsnotify.Write == fsnotify.Write {
			now := time.Now()
			if lastAttemptedReload.Add(minTimeBetweenReloadAttempts).Before(now) {
				cc.logger.Debugw("Config file modified, attempting reload", "event", event)
				<-time.After(delayBetweenEventAndReload)

				if err := cc.Load(); err != nil {
					cc.logger.Warnw("Failed to reload config file", "error", err)
				} else {
					cc.logger.Info("Reloaded config successfully")
					cc.notifier.Notify("Configuration reloaded!", "Your changes have been applied.")
					cc.onConfigReloaded()
				}

				lastAttemptedReload = now
			}
		}
	})

	<-cc.stopWatcherChannel
	cc.logger.Debug("Stopping user config file watcher")
	cc.userConfig.OnConfigChange(nil)
}

// StopWatchingConfigFile signals our filesystem watcher to stop
func (cc *CanonicalConfig) StopWatchingConfigFile() {
	cc.stopWatcherChannel <- true
}

func (cc *CanonicalConfig) populateFromVipers() error {
	cc.SliderMapping = sliderMapFromConfigs(
		cc.userConfig.GetStringMapStringSlice(configKeySliderMapping),
		cc.internalConfig.GetStringMapStringSlice(configKeySliderMapping),
	)

	cc.buttonMapping = buttonrMapFromConfigs(
		cc.userConfig.GetStringMapStringSlice(configKeyButtonMapping),
		cc.internalConfig.GetStringMapStringSlice(configKeyButtonMapping),
	)

	switch cc.userConfig.GetString(configKeyControllerType) {
	case "serial", "udp":
		cc.ControllerType = cc.userConfig.GetString(configKeyControllerType)
	default:
		cc.logger.Warnw("Invalid controller type specified, using default value",
			"key", configKeyControllerType,
			"invalidValue", cc.userConfig.GetString(configKeyControllerType),
			"defaultValue", defaultControllerType)
		cc.ControllerType = defaultControllerType
	}

	cc.SerialConnectionInfo.COMPort = cc.userConfig.GetString(configKeyCOMPort)

	cc.SerialConnectionInfo.BaudRate = cc.userConfig.GetInt(configKeyBaudRate)
	if cc.SerialConnectionInfo.BaudRate <= 0 {
		cc.logger.Warnw("Invalid baud rate specified, using default value",
			"key", configKeyBaudRate,
			"invalidValue", cc.SerialConnectionInfo.BaudRate,
			"defaultValue", defaultBaudRate)
		cc.SerialConnectionInfo.BaudRate = defaultBaudRate
	}

	cc.UdpConnectionInfo.UdpPort = cc.userConfig.GetInt(configKeyUdpPort)
	if cc.UdpConnectionInfo.UdpPort <= 0 || cc.UdpConnectionInfo.UdpPort >= 65536 {
		cc.logger.Warnw("Invalid UDP port specified, using default value",
			"key", configKeyUdpPort,
			"invalidValue", cc.UdpConnectionInfo.UdpPort,
			"defaultValue", defaultUdpPort)
		cc.UdpConnectionInfo.UdpPort = defaultUdpPort
	}

	cc.InvertSliders = cc.userConfig.GetBool(configKeyInvertSliders)
	cc.NoiseReductionLevel = cc.userConfig.GetString(configKeyNoiseReductionLevel)

	cc.logger.Debug("Populated config fields from vipers")

	return nil
}

func (cc *CanonicalConfig) onConfigReloaded() {
	cc.logger.Debug("Notifying consumers about configuration reload")
	for _, consumer := range cc.reloadConsumers {
		consumer <- true
	}
}

// findConfigFile looks for the config file in multiple locations
func findConfigFile(executablePath string) (string, error) {
	// Check in the same directory as the executable
	exeDir := filepath.Dir(executablePath)
	configPath := filepath.Join(exeDir, userConfigFilepath)
	if util.FileExists(configPath) {
		return configPath, nil
	}

	// Check in the current working directory
	cwd, err := os.Getwd()
	if err == nil {
		configPath = filepath.Join(cwd, userConfigFilepath)
		if util.FileExists(configPath) {
			return configPath, nil
		}
	}

	// Check in the user's home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath = filepath.Join(homeDir, ".deej", userConfigFilepath)
		if util.FileExists(configPath) {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("config file not found")
}

// SetConfigFile sets the path for the user config file
func (cc *CanonicalConfig) SetConfigFile(path string) {
	cc.userConfig.SetConfigFile(path)
}
