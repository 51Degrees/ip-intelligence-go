package ipi_onpremise

import (
	"bytes"
	"github.com/51Degrees/ip-intelligence-go/ipi_interop"
	"strings"
	"testing"
)

const (
	liteDataPath  = "/path/to/51Degrees-LiteIpiV41.ipi"
	licenseKey    = "<KEY>"
	dataUpdateUrl = "https://example.com/ipi/data/51Degrees-LiteIpiV41.ipi"
)

func TestNew_WithDataFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		profile     ipi_interop.PerformanceProfile
		wantProfile ipi_interop.PerformanceProfile
	}{
		{
			name:        "with lite data file and InMemory config",
			filePath:    liteDataPath,
			profile:     ipi_interop.InMemory,
			wantProfile: ipi_interop.InMemory,
		},
		// TODO: uncomment when other profiles will be available
		//{
		//	name:       "with lite data file and HighPerformance config",
		//	filePath:   liteDataPath,
		//	profile:     ipi_interop.HighPerformance,
		//	wantProfile: ipi_interop.HighPerformance,
		//},
		//{
		//	name:       "with lite data file and LowMemory config",
		//	filePath:   liteDataPath,
		//	profile:     ipi_interop.LowMemory,
		//	wantProfile: ipi_interop.LowMemory,
		//},
		//{
		//	name:       "with lite data file and Balanced config",
		//	filePath:   liteDataPath,
		//	profile:     ipi_interop.Balanced,
		//	wantProfile: ipi_interop.Balanced,
		//},
		//{
		//	name:       "with lite data file and BalancedTemp config",
		//	filePath:   liteDataPath,
		//	profile:     ipi_interop.BalancedTemp,
		//	wantProfile: ipi_interop.BalancedTemp,
		//},
		//{
		//	name:       "with lite data file and SingleLoaded config",
		//	filePath:   liteDataPath,
		//	profile:     ipi_interop.SingleLoaded,
		//	wantProfile: ipi_interop.SingleLoaded,
		//},
		//{
		//	name:       "with lite data file and Default config",
		//	filePath:   liteDataPath,
		//	profile:     ipi_interop.Default,
		//	wantProfile: ipi_interop.Default,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(tt.profile)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}
			if config.CPtr == nil {
				t.Error("ConfigIpi.CPtr is nil")
			}
			if got := config.PerformanceProfile(); got != tt.wantProfile {
				t.Errorf("NewConfigIpi() profile = %v, want %v", got, tt.wantProfile)
			}

			engine, err := New(
				WithTempDataCopy(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config),
				WithAutoUpdate(false),
			)

			if err != nil {
				t.Fatalf("Expected no error with valid file, got: %v", err)
			}

			if engine.dataFile != liteDataPath {
				t.Errorf("Expected dataFile to be %s, got %s", liteDataPath, engine.dataFile)
			}

			engine.Stop()
		})
	}
}

func TestNew_WithInvalidDataFile(t *testing.T) {
	_, err := New(
		WithDataFile("non_existent_file"),
	)
	if err == nil {
		t.Error("Expected error with invalid file path, got nil")
	}
}

func TestWithDataUpdateUrl(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		updateUrl     string
		wantUpdateUrl string
	}{
		{
			name:          "with data update url",
			filePath:      liteDataPath,
			updateUrl:     dataUpdateUrl,
			wantUpdateUrl: dataUpdateUrl,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			engine, err := New(
				WithTempDataCopy(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config),
				WithAutoUpdate(false),
				WithDataUpdateUrl(tt.updateUrl),
			)
			if err != nil {
				t.Fatalf("Expected no error with valid file, got: %v", err)
			}

			if engine.dataFileUrl != tt.wantUpdateUrl {
				t.Errorf("Expected dataFileUrl to be %s, got %s", tt.wantUpdateUrl, engine.dataFileUrl)
			}

			engine.Stop()
		})
	}
}

func TestWithTempDataDir(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		tempDataDir     string
		wantTempDataDir string
	}{
		{
			name:            "with data update url",
			filePath:        liteDataPath,
			tempDataDir:     "test_dir",
			wantTempDataDir: "test_dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}
			engine, err := New(
				WithTempDataCopy(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config),
				WithAutoUpdate(false),
				WithTempDataDir(tt.tempDataDir),
			)
			if err != nil {
				t.Fatalf("Expected no error with valid file, got: %v", err)
			}

			if engine.tempDataDir != tt.wantTempDataDir {
				t.Errorf("Expected dataFileUrl to be %s, got %s", tt.wantTempDataDir, engine.dataFileUrl)
			}

			engine.Stop()
		})
	}
}

func TestWithPollingInterval(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		interval     int
		wantInterval int
	}{
		{
			name:         "with polling interval",
			filePath:     liteDataPath,
			interval:     60,
			wantInterval: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}
			engine, err := New(
				WithTempDataCopy(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config),
				WithAutoUpdate(false),
				WithPollingInterval(tt.interval),
			)
			if err != nil {
				t.Fatalf("Expected no error with valid file, got: %v", err)
			}

			expectedMs := tt.wantInterval * 1000
			if engine.dataFilePullEveryMs != expectedMs {
				t.Errorf("Expected polling interval to be %d ms, got %d ms", expectedMs, engine.dataFilePullEveryMs)
			}

			engine.Stop()
		})
	}
}

func TestWithAutoUpdate(t *testing.T) {
	tests := []struct {
		name            string
		autoUpdateValue bool
		wantEnabled     bool
		options         []EngineOptions
	}{
		{
			name:            "auto update enabled",
			autoUpdateValue: true,
			wantEnabled:     true,
			options: []EngineOptions{
				WithAutoUpdate(true),
				WithDataUpdateUrl(dataUpdateUrl),
				WithLicenseKey(licenseKey),
			},
		},
		{
			name:            "auto update disabled",
			autoUpdateValue: false,
			wantEnabled:     false,
			options: []EngineOptions{
				WithAutoUpdate(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			options := append(tt.options,
				WithTempDataCopy(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config))

			engine, err := New(options...)
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			if engine.isAutoUpdateEnabled != tt.wantEnabled {
				t.Errorf("WithAutoUpdate() = %v, want %v",
					engine.isAutoUpdateEnabled, tt.wantEnabled)
			}

			engine.Stop()
		})
	}
}

func TestWithLicenseKey(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		options []EngineOptions
	}{
		{
			name:    "valid license key",
			wantErr: false,
			options: []EngineOptions{
				WithLicenseKey(licenseKey),
				WithDataUpdateUrl(dataUpdateUrl),
				WithAutoUpdate(true),
			},
		},
		{
			name:    "empty license key",
			wantErr: false,
			options: []EngineOptions{
				WithLicenseKey(""),
				WithDataUpdateUrl(dataUpdateUrl),
				WithAutoUpdate(false),
			},
		},
		{
			name:    "license key with special characters",
			wantErr: false,
			options: []EngineOptions{
				WithLicenseKey("license-key-123!@#$%^&*()"),
				WithDataUpdateUrl(dataUpdateUrl),
				WithAutoUpdate(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			options := append(tt.options, WithTempDataCopy(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config))

			engine, err := New(options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithLicenseKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			engine.Stop()
		})
	}
}

func TestWithProduct(t *testing.T) {
	tests := []struct {
		name    string
		options []EngineOptions
		wantErr bool
	}{
		{
			name: "product with license key",
			options: []EngineOptions{
				WithProduct("TestProduct"),
				WithLicenseKey(licenseKey),
				WithAutoUpdate(false),
			},
			wantErr: false,
		},
		{
			name: "product with custom URL",
			options: []EngineOptions{
				WithProduct("TestProduct"),
				WithDataUpdateUrl(dataUpdateUrl),
			},
			wantErr: false,
		},
		{
			name: "product with auto update enabled",
			options: []EngineOptions{
				WithProduct("TestProduct"),
				WithAutoUpdate(true),
				WithDataUpdateUrl(dataUpdateUrl),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			// Add required WithDataFile option to test options
			options := append(tt.options, WithTempDataCopy(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config))

			engine, err := New(options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithProduct() with other options error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			engine.Stop()
		})
	}
}

func TestWithMaxRetries(t *testing.T) {
	tests := []struct {
		name        string
		retries     int
		wantRetries int
		options     []EngineOptions
	}{
		{
			name:        "positive retries",
			wantRetries: 5,
			options: []EngineOptions{
				WithMaxRetries(5),
			},
		},
		{
			name:        "zero retries",
			wantRetries: 0,
			options: []EngineOptions{
				WithMaxRetries(0),
			},
		},
		{
			name:        "negative retries",
			wantRetries: -1,
			options: []EngineOptions{
				WithMaxRetries(-1),
			},
		},
		{
			name:        "large number of retries",
			wantRetries: 1000,
			options: []EngineOptions{
				WithMaxRetries(1000),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			// Add required WithDataFile option to test options
			options := append(tt.options, WithTempDataCopy(false),
				WithAutoUpdate(false),
				WithDataFile(liteDataPath),
				WithConfigIpi(config))

			engine, err := New(options...)
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			if engine.maxRetries != tt.wantRetries {
				t.Errorf("WithMaxRetries() got = %v, want %v", engine.maxRetries, tt.wantRetries)
			}

			engine.Stop()
		})
	}
}

func TestWithLogging(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		wantEnabled bool
	}{
		{
			name:        "enable logging",
			enabled:     true,
			wantEnabled: true,
		},
		{
			name:        "disable logging",
			enabled:     false,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			engine, err := New(
				WithConfigIpi(config),
				WithAutoUpdate(false),
				WithDataFile(liteDataPath),

				WithLogging(tt.enabled),
			)
			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			if engine.logger.enabled != tt.wantEnabled {
				t.Errorf("WithLogging() got = %v, want %v", engine.logger.enabled, tt.wantEnabled)
			}

			engine.Stop()
		})
	}
}

// customTestLogger implements LogWriter interface for testing
type customTestLogger struct {
	buffer bytes.Buffer
	called bool
}

func (l *customTestLogger) Printf(format string, v ...interface{}) {
	l.called = true
}

func TestWithCustomLogger(t *testing.T) {
	tests := []struct {
		name    string
		logger  LogWriter
		wantErr bool
		testLog bool // whether to test logging functionality
	}{
		{
			name:    "valid custom logger",
			logger:  &customTestLogger{},
			wantErr: false,
			testLog: true,
		},
		{
			name:    "default logger replacement",
			logger:  DefaultLogger,
			wantErr: false,
			testLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			engine, err := New(
				WithConfigIpi(config),
				WithAutoUpdate(false),
				WithDataFile(liteDataPath),

				WithCustomLogger(tt.logger),
			)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("WithCustomLogger() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if tt.wantErr {
				t.Error("WithCustomLogger() expected error, got none")
				return
			}

			// Verify logger was set correctly
			if engine.logger.logger != tt.logger {
				t.Errorf("WithCustomLogger() logger not set correctly")
			}

			// Verify logging is enabled by default
			if !engine.logger.enabled {
				t.Error("WithCustomLogger() logger should be enabled by default")
			}

			// Test logging functionality if required
			if tt.testLog {
				if customLogger, ok := engine.logger.logger.(*customTestLogger); ok {
					engine.logger.Printf("test message")
					if !customLogger.called {
						t.Error("WithCustomLogger() logger was not called")
					}
				}
			}

			engine.Stop()
		})
	}
}

func TestWithFileWatch(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		filePath    string
		wantEnabled bool
	}{
		{
			name:        "enable file watching",
			enabled:     true,
			filePath:    liteDataPath,
			wantEnabled: true,
		},
		{
			name:        "disable file watching",
			enabled:     false,
			filePath:    liteDataPath,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			// Create engine with test configuration
			engine, err := New(
				WithConfigIpi(config),
				WithAutoUpdate(false),
				WithDataFile(liteDataPath),

				WithTempDataCopy(false),
				WithDataFile(tt.filePath),
				WithAutoUpdate(false),
				WithFileWatch(tt.enabled),
			)

			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			// Verify the flag is set correctly
			if engine.isFileWatcherEnabled != tt.wantEnabled {
				t.Errorf("WithFileWatch(%v) = %v, want %v",
					tt.enabled,
					engine.isFileWatcherEnabled,
					tt.wantEnabled)
			}

			// Verify file watcher state matches the enabled flag
			if tt.enabled {
				if !engine.fileWatcherStarted {
					t.Error("File watcher should be started when enabled")
				}
				if engine.fileWatcher == nil {
					t.Error("File watcher should not be nil when enabled")
				}
			} else {
				if engine.fileWatcherStarted {
					t.Error("File watcher should not be started when disabled")
				}
				if engine.fileWatcher != nil {
					t.Error("File watcher should be nil when disabled")
				}
			}

			engine.Stop()
		})
	}
}

func TestWithUpdateOnStart(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		dataFile    string
		wantEnabled bool
	}{
		{
			name:        "enable update on start",
			enabled:     true,
			dataFile:    liteDataPath,
			wantEnabled: true,
		},
		{
			name:        "disable update on start",
			enabled:     false,
			dataFile:    liteDataPath,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			// Create engine with test configuration
			engine, err := New(
				WithConfigIpi(config),
				WithDataFile(liteDataPath),
				WithDataUpdateUrl(dataUpdateUrl),

				WithTempDataCopy(false),
				WithDataFile(tt.dataFile),
				WithAutoUpdate(true), // Enable auto update to test update on start
				WithUpdateOnStart(tt.enabled),
			)

			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			// Verify the flag is set correctly
			if engine.isUpdateOnStartEnabled != tt.wantEnabled {
				t.Errorf("WithUpdateOnStart(%v) = %v, want %v",
					tt.enabled,
					engine.isUpdateOnStartEnabled,
					tt.wantEnabled)
			}

			engine.Stop()
		})
	}
}

func TestWithTempDataCopy(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		dataFile    string
		wantEnabled bool
	}{
		{
			name:        "enable temp data copy",
			enabled:     true,
			dataFile:    liteDataPath,
			wantEnabled: true,
		},
		{
			name:        "disable temp data copy",
			enabled:     false,
			dataFile:    liteDataPath,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			// Create engine with test configuration
			engine, err := New(
				WithConfigIpi(config),
				WithDataUpdateUrl(dataUpdateUrl),

				WithDataFile(tt.dataFile),
				WithTempDataCopy(tt.enabled),
			)

			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			// Verify the flag is set correctly
			if engine.isCreateTempDataCopyEnabled != tt.wantEnabled {
				t.Errorf("WithTempDataCopy(%v) = %v, want %v",
					tt.enabled,
					engine.isCreateTempDataCopyEnabled,
					tt.wantEnabled)
			}

			// Check if temp directory is created when enabled
			if tt.enabled {
				if engine.tempDataDir == "" {
					t.Error("Temp data directory should be set when temp data copy is enabled")
				}
			}

			engine.Stop()
		})
	}
}

func TestWithRandomization(t *testing.T) {
	tests := []struct {
		name       string
		seconds    int
		wantMillis int
		dataFile   string
	}{
		{
			name:       "zero randomization",
			seconds:    0,
			wantMillis: 0,
			dataFile:   liteDataPath,
		},
		{
			name:       "positive randomization",
			seconds:    60,
			wantMillis: 60000,
			dataFile:   liteDataPath,
		},
		{
			name:       "large randomization",
			seconds:    3600,
			wantMillis: 3600000,
			dataFile:   liteDataPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			// Create engine with test configuration
			engine, err := New(
				WithConfigIpi(config),
				WithDataUpdateUrl(dataUpdateUrl),

				WithDataFile(tt.dataFile),
				WithRandomization(tt.seconds),
			)

			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}

			// Verify the randomization is set correctly
			if engine.randomization != tt.wantMillis {
				t.Errorf("WithRandomization(%v) = %v, want %v",
					tt.seconds,
					engine.randomization,
					tt.wantMillis)
			}

			engine.Stop()
		})
	}
}

func TestWithProperties(t *testing.T) {
	tests := []struct {
		name       string
		properties []string
		wantJoined string
		dataFile   string
	}{
		{
			name:       "empty properties list",
			properties: []string{},
			wantJoined: "",
			dataFile:   liteDataPath,
		},
		//{
		//	name:       "single property",
		//	properties: []string{"IpRangeStart"},
		//	wantJoined: "IpRangeStart",
		//	dataFile:   liteDataPath,
		//},
		{
			name:       "multiple properties",
			properties: []string{"IpRangeStart", "IpRangeEnd", "AccuracyRadius", "RegisteredCountry", "RegisteredName", "Longitude", "Latitude", "Areas"},
			wantJoined: "IpRangeStart,IpRangeEnd,AccuracyRadius,RegisteredCountry,RegisteredName,Longitude,Latitude,Areas",
			dataFile:   liteDataPath,
		},
		{
			name:       "nil properties",
			properties: nil,
			wantJoined: strings.Join(defaultProperties, ","), // Should keep default properties
			dataFile:   liteDataPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ipi_interop.NewConfigIpi(ipi_interop.InMemory)
			if config == nil {
				t.Fatal("NewConfigIpi returned nil")
			}

			// Create engine with test configuration
			engine, err := New(
				WithConfigIpi(config),
				WithAutoUpdate(false),
				WithDataFile(tt.dataFile),
				WithProperties(tt.properties),
			)

			if err != nil {
				t.Fatalf("Failed to create engine: %v", err)
			}
			defer engine.Stop()

			// Verify the properties are set correctly
			if tt.properties == nil {
				// For nil properties, should keep default properties
				if engine.managerProperties != tt.wantJoined {
					t.Errorf("WithProperties(nil) = %v, want %v",
						engine.managerProperties,
						tt.wantJoined)
				}
			} else {
				if engine.managerProperties != tt.wantJoined {
					t.Errorf("WithProperties(%v) = %v, want %v",
						tt.properties,
						engine.managerProperties,
						tt.wantJoined)
				}
			}
		})
	}
}
