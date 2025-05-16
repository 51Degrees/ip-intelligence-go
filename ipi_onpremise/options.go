package ipi_onpremise

import (
	"fmt"
	"github.com/51Degrees/ip-intelligence-go/ipi_interopt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type EngineOptions func(cfg *Engine) error

// WithDataFile sets the path to the local data file, this parameter is required to start the engine
func WithDataFile(path string) EngineOptions {
	return func(cfg *Engine) error {
		path := filepath.Join(path)
		_, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to get file path: %w", err)
		}

		cfg.dataFile = path
		return nil
	}
}

// WithConfigIpi allows to configure the Ipi matching algorithm.
// See ipi_interopt.ConfigIpi type for all available settings:
// PerformanceProfile, Drift, Difference, Concurrency
// By default initialized with ipi_interopt.Balanced performance profile
// ipi_interopt.NewConfigIpi(ipi_interopt.Balanced)
func WithConfigIpi(configIpi *ipi_interopt.ConfigIpi) EngineOptions {
	return func(cfg *Engine) error {
		cfg.config = configIpi
		return nil
	}
}

// WithLicenseKey sets the license key to use when pulling the data file
// this option can only be used when using the default data file url from 51Degrees, it will be appended as a query parameter
func WithLicenseKey(key string) EngineOptions {
	return func(cfg *Engine) error {
		//if !cfg.isDefaultDataFileUrl() {
		//	return errors.New("license key can only be set when using default data file url")
		//}
		//cfg.licenseKey = key
		return nil
	}
}

// WithProduct sets the product to use when pulling the data file when distributor service is used
// licenseKey has to be provided using WithLicenseKey
func WithProduct(product string) EngineOptions {
	return func(cfg *Engine) error {
		//if !cfg.isDefaultDataFileUrl() {
		//	return errors.New("product can only be set when using default data file url")
		//}
		//
		//cfg.product = product
		return nil
	}
}

// WithDataUpdateUrl sets a custom URL to download the data file from
func WithDataUpdateUrl(urlStr string) EngineOptions {
	return func(cfg *Engine) error {
		_, err := url.ParseRequestURI(urlStr)
		if err != nil {
			return err
		}

		cfg.dataFileUrl = urlStr

		return nil
	}
}

// WithMaxRetries sets the maximum number of retries to pull the data file if request fails
func WithMaxRetries(retries int) EngineOptions {
	return func(cfg *Engine) error {
		cfg.maxRetries = retries
		return nil
	}
}

// WithPollingInterval sets the interval in seconds to pull the data file
func WithPollingInterval(seconds int) EngineOptions {
	return func(cfg *Engine) error {
		cfg.dataFilePullEveryMs = seconds * 1000
		return nil
	}
}

// WithLogging enables or disables the logger
func WithLogging(enabled bool) EngineOptions {
	return func(cfg *Engine) error {
		cfg.logger.enabled = enabled
		return nil
	}
}

// WithCustomLogger sets a custom logger
func WithCustomLogger(logger LogWriter) EngineOptions {
	return func(cfg *Engine) error {
		cfg.logger = logWrapper{
			enabled: true,
			logger:  logger,
		}

		return nil
	}
}

// WithFileWatch enables or disables file watching in case 3rd party updates the data file
// engine will automatically reload the data file.  Default is true
func WithFileWatch(enabled bool) EngineOptions {
	return func(cfg *Engine) error {
		cfg.isFileWatcherEnabled = enabled
		return nil
	}
}

// WithUpdateOnStart enables or disables update on start
// if enabled, engine will pull the data file from the distributor (or custom URL) once initialized
// default is false
func WithUpdateOnStart(enabled bool) EngineOptions {
	return func(cfg *Engine) error {
		cfg.isUpdateOnStartEnabled = enabled

		return nil
	}
}

// WithAutoUpdate enables or disables auto update
// default is true
// if enabled, engine will automatically pull the data file from the distributor or custom URL
// if disabled options like WithDataUpdateUrl, WithLicenseKey will be ignored
func WithAutoUpdate(enabled bool) EngineOptions {
	return func(cfg *Engine) error {
		cfg.isAutoUpdateEnabled = enabled

		return nil
	}
}

// WithTempDataCopy enables or disables creating a temp copy of the data file
// default is true
// if enabled, engine will create a temp copy of the data file and use it for detection rather than original data file
// if disabled, engine will use the original data file to initialize the manager
// this is useful when 3rd party updates the data file on the file system
func WithTempDataCopy(enabled bool) EngineOptions {
	return func(cfg *Engine) error {
		cfg.isCreateTempDataCopyEnabled = enabled

		return nil
	}
}

// WithTempDataDir sets the directory to store the temp data file
// default is system temp directory
func WithTempDataDir(dir string) EngineOptions {
	return func(cfg *Engine) error {
		dirFileInfo, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("failed to get file path: %w", err)
		}

		if !dirFileInfo.IsDir() {
			return fmt.Errorf("path is not a directory: %s", dir)
		}

		cfg.tempDataDir = dir
		return nil
	}
}

// WithRandomization sets the randomization time in seconds
// default is 10 minutes
// if set, when scheduling the file pulling, it will add randomization time to the interval
// this is useful to avoid multiple engines pulling the data file at the same time in case of multiple engines/instances
func WithRandomization(seconds int) EngineOptions {
	return func(cfg *Engine) error {
		cfg.randomization = seconds * 1000
		return nil
	}
}

// WithProperties sets properties that the engine retrieves from the data file for each device detection result instance
// default is [] which will include all possible properties
func WithProperties(properties []string) EngineOptions {
	return func(cfg *Engine) error {
		if properties != nil {
			cfg.managerProperties = strings.Join(properties, ",")
		}
		return nil
	}
}
