package ipi_onpremise

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	ipiDataPath   = ""
	licenseKey    = "<KEY>"
	dataUpdateUrl = "https://example.com/ipi/data/51Degrees-LiteIpiV41.ipi"
)

func TestWithDataFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-data-dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	validFile := filepath.Join(tempDir, "valid.data")
	if err := os.WriteFile(validFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid file path",
			path:        validFile,
			expectError: false,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: true,
			errorMsg:    "failed to get file path",
		},
		{
			name:        "non-existent file",
			path:        filepath.Join(tempDir, "nonexistent.data"),
			expectError: true,
			errorMsg:    "failed to get file path",
		},
		{
			name:        "directory instead of file",
			path:        tempDir,
			expectError: false, // Note: The function doesn't validate if it's a file or directory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithDataFile option
			err := WithDataFile(tt.path)(engine)

			// Check error conditions
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error, but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}

				// Verify the path was set correctly
				expectedPath := filepath.Join(tt.path)
				if engine.dataFile != expectedPath {
					t.Errorf("Expected dataFile to be %q, got %q", expectedPath, engine.dataFile)
				}
			}
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
		name        string
		url         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid http URL",
			url:         "http://example.com/data/file.dat",
			expectError: false,
		},
		{
			name:        "valid https URL",
			url:         "https://example.com/data/file.dat",
			expectError: false,
		},
		{
			name:        "valid URL with query parameters",
			url:         "https://example.com/data/file.dat?key=value&other=param",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
			errorMsg:    "parse \"\": empty url",
		},
		{
			name:        "invalid URL format",
			url:         "not-a-url",
			expectError: true,
			errorMsg:    "invalid URI for request",
		},
		{
			name:        "missing protocol",
			url:         "example.com/file",
			expectError: true,
			errorMsg:    "invalid URI for request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithDataUpdateUrl option
			err := WithDataUpdateUrl(tt.url)(engine)

			// Check error conditions
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error, but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if engine.dataFileUrl != tt.url {
					t.Errorf("Expected dataFileUrl to be %q, got %q", tt.url, engine.dataFileUrl)
				}
			}
		})
	}
}

func TestWithTempDataDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-temp-dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file in the temp directory to test the "not a directory" case
	tempFile := filepath.Join(tempDir, "testfile")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	tests := []struct {
		name        string
		dir         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid directory",
			dir:         tempDir,
			expectError: false,
		},
		{
			name:        "non-existent directory",
			dir:         filepath.Join(tempDir, "nonexistent"),
			expectError: true,
			errorMsg:    "failed to get file path",
		},
		{
			name:        "file instead of directory",
			dir:         tempFile,
			expectError: true,
			errorMsg:    "path is not a directory",
		},
		{
			name:        "empty path",
			dir:         "",
			expectError: true,
			errorMsg:    "failed to get file path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance to test the option
			engine := &Engine{}

			// Apply the WithTempDataDir option
			err := WithTempDataDir(tt.dir)(engine)

			// Check error conditions
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error, but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if engine.tempDataDir != tt.dir {
					t.Errorf("Expected tempDataDir to be %q, got %q", tt.dir, engine.tempDataDir)
				}
			}
		})
	}
}

func TestWithPollingInterval(t *testing.T) {
	tests := []struct {
		name       string
		seconds    int
		expectedMs int
	}{
		{
			name:       "zero seconds",
			seconds:    0,
			expectedMs: 0,
		},
		{
			name:       "one second",
			seconds:    1,
			expectedMs: 1000,
		},
		{
			name:       "sixty seconds",
			seconds:    60,
			expectedMs: 60000,
		},
		{
			name:       "negative value",
			seconds:    -1,
			expectedMs: -1000,
		},
		{
			name:       "large value",
			seconds:    86400, // 24 hours
			expectedMs: 86400000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithPollingInterval option
			err := WithPollingInterval(tt.seconds)(engine)

			// Polling interval should never return an error
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check if the polling interval was set correctly
			if engine.dataFilePullEveryMs != tt.expectedMs {
				t.Errorf("Expected polling interval to be %d ms, got %d ms",
					tt.expectedMs, engine.dataFilePullEveryMs)
			}
		})
	}
}

func TestWithAutoUpdate(t *testing.T) {
	// Test the direct option function
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable auto update",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable auto update",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &Engine{}

			err := WithAutoUpdate(tt.enabled)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if engine.isAutoUpdateEnabled != tt.expected {
				t.Errorf("Expected isAutoUpdateEnabled to be %v, got %v",
					tt.expected, engine.isAutoUpdateEnabled)
			}
		})
	}
}

func TestWithLicenseKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectError bool
	}{
		{
			name:        "valid license key",
			key:         "valid-license-key",
			expectError: false,
		},
		{
			name:        "empty license key",
			key:         "",
			expectError: false,
		},
		{
			name:        "special characters in key",
			key:         "key@123!#$%^",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithLicenseKey option
			err := WithLicenseKey(tt.key)(engine)

			// Check error conditions
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
			}
		})
	}
}

func TestWithProduct(t *testing.T) {
	tests := []struct {
		name        string
		product     string
		expectError bool
	}{
		{
			name:        "valid product name",
			product:     "TestProduct",
			expectError: false,
		},
		{
			name:        "empty product name",
			product:     "",
			expectError: false,
		},
		{
			name:        "product with special characters",
			product:     "Test-Product_123",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithProduct option
			err := WithProduct(tt.product)(engine)

			// Check error conditions
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
			}
		})
	}
}

func TestWithMaxRetries(t *testing.T) {
	tests := []struct {
		name    string
		retries int
		want    int
	}{
		{
			name:    "positive retries",
			retries: 5,
			want:    5,
		},
		{
			name:    "zero retries",
			retries: 0,
			want:    0,
		},
		{
			name:    "negative retries",
			retries: -1,
			want:    -1,
		},
		{
			name:    "large number of retries",
			retries: 1000,
			want:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithMaxRetries option
			err := WithMaxRetries(tt.retries)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("WithMaxRetries() error = %v, want no error", err)
			}

			// Check if retries was set correctly
			if engine.maxRetries != tt.want {
				t.Errorf("WithMaxRetries() got = %v, want %v",
					engine.maxRetries, tt.want)
			}
		})
	}
}

func TestWithLogging(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable logging",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable logging",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{
				logger: logWrapper{
					enabled: !tt.enabled, // Set initial state opposite to test value
				},
			}

			// Apply the WithLogging option
			err := WithLogging(tt.enabled)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("WithLogging() error = %v, want no error", err)
			}

			// Check if logging was set correctly
			if engine.logger.enabled != tt.expected {
				t.Errorf("WithLogging() got = %v, want %v",
					engine.logger.enabled, tt.expected)
			}
		})
	}
}

// Define test logger types
type testLogger struct {
	messages []string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(format, v...))
}

type nilLogger struct{}

func (l *nilLogger) Printf(format string, v ...interface{}) {}

func TestWithCustomLogger(t *testing.T) {
	tests := []struct {
		name          string
		logger        LogWriter
		expectEnabled bool
		verifyLogger  bool
	}{
		{
			name:          "custom logger",
			logger:        &testLogger{},
			expectEnabled: true,
			verifyLogger:  true,
		},
		{
			name:          "nil logger",
			logger:        &nilLogger{},
			expectEnabled: true,
			verifyLogger:  true,
		},
		{
			name:          "nil value",
			logger:        nil,
			expectEnabled: true,
			verifyLogger:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithCustomLogger option
			err := WithCustomLogger(tt.logger)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("WithCustomLogger() error = %v, want no error", err)
			}

			// Verify logger is enabled by default
			if engine.logger.enabled != tt.expectEnabled {
				t.Errorf("WithCustomLogger() enabled = %v, want %v",
					engine.logger.enabled, tt.expectEnabled)
			}

			// Verify logger was set correctly
			if tt.verifyLogger && engine.logger.logger != tt.logger {
				t.Errorf("WithCustomLogger() logger not set correctly, got %v, want %v",
					engine.logger.logger, tt.logger)
			}

			// Test logger functionality if it's a testLogger
			if testLog, ok := tt.logger.(*testLogger); ok {
				testMessage := "test message"
				engine.logger.Printf(testMessage)
				if len(testLog.messages) != 1 {
					t.Errorf("Expected 1 message, got %d", len(testLog.messages))
				}
				if len(testLog.messages) > 0 && testLog.messages[0] != testMessage {
					t.Errorf("Expected message %q, got %q", testMessage, testLog.messages[0])
				}
			}
		})
	}
}

func TestWithFileWatch(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable file watching",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable file watching",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{
				isFileWatcherEnabled: !tt.enabled, // Set initial state opposite to test value
			}

			// Apply the WithFileWatch option
			err := WithFileWatch(tt.enabled)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("WithFileWatch() error = %v, want no error", err)
			}

			// Check if file watching was set correctly
			if engine.isFileWatcherEnabled != tt.expected {
				t.Errorf("WithFileWatch() got = %v, want %v",
					engine.isFileWatcherEnabled, tt.expected)
			}
		})
	}
}

func TestWithUpdateOnStart(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable update on start",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable update on start",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{
				isUpdateOnStartEnabled: !tt.enabled, // Set initial state opposite to test value
			}

			// Apply the WithUpdateOnStart option
			err := WithUpdateOnStart(tt.enabled)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("WithUpdateOnStart() error = %v, want no error", err)
			}

			// Check if update on start was set correctly
			if engine.isUpdateOnStartEnabled != tt.expected {
				t.Errorf("WithUpdateOnStart() got = %v, want %v",
					engine.isUpdateOnStartEnabled, tt.expected)
			}
		})
	}
}

func TestWithTempDataCopy(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable temp data copy",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable temp data copy",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{
				isCreateTempDataCopyEnabled: !tt.enabled, // Set initial state opposite to test value
			}

			// Apply the WithTempDataCopy option
			err := WithTempDataCopy(tt.enabled)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("WithTempDataCopy() error = %v, want no error", err)
			}

			// Check if temp data copy setting was set correctly
			if engine.isCreateTempDataCopyEnabled != tt.expected {
				t.Errorf("WithTempDataCopy() got = %v, want %v",
					engine.isCreateTempDataCopyEnabled, tt.expected)
			}
		})
	}
}

func TestWithRandomization(t *testing.T) {
	tests := []struct {
		name           string
		seconds        int
		expectedMillis int
	}{
		{
			name:           "zero seconds",
			seconds:        0,
			expectedMillis: 0,
		},
		{
			name:           "one second",
			seconds:        1,
			expectedMillis: 1000,
		},
		{
			name:           "ten seconds",
			seconds:        10,
			expectedMillis: 10000,
		},
		{
			name:           "negative value",
			seconds:        -1,
			expectedMillis: -1000,
		},
		{
			name:           "large value",
			seconds:        3600, // 1 hour
			expectedMillis: 3600000,
		},
		{
			name:           "default value (10 minutes)",
			seconds:        600,
			expectedMillis: 600000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithRandomization option
			err := WithRandomization(tt.seconds)(engine)

			// Should never return error
			if err != nil {
				t.Errorf("WithRandomization() error = %v, want no error", err)
			}

			// Check if randomization was set correctly
			if engine.randomization != tt.expectedMillis {
				t.Errorf("WithRandomization() got = %v ms, want %v ms",
					engine.randomization, tt.expectedMillis)
			}
		})
	}
}

func TestWithProperties(t *testing.T) {
	tests := []struct {
		name               string
		properties         []string
		expectError        bool
		expectedProperties []string
	}{
		{
			name:               "valid properties slice",
			properties:         []string{"Property1", "Property2", "Property3"},
			expectError:        false,
			expectedProperties: []string{"Property1", "Property2", "Property3"},
		},
		{
			name:               "empty properties slice",
			properties:         []string{},
			expectError:        false,
			expectedProperties: []string{},
		},
		{
			name:               "single property",
			properties:         []string{"SingleProperty"},
			expectError:        false,
			expectedProperties: []string{"SingleProperty"},
		},
		{
			name:               "properties with special characters",
			properties:         []string{"Property-1", "Property_2", "Property@3"},
			expectError:        false,
			expectedProperties: []string{"Property-1", "Property_2", "Property@3"},
		},
		{
			name:               "properties with spaces",
			properties:         []string{"Property Name", "Another Property"},
			expectError:        false,
			expectedProperties: []string{"Property Name", "Another Property"},
		},
		{
			name:               "properties with empty strings",
			properties:         []string{"", "ValidProperty", ""},
			expectError:        false,
			expectedProperties: []string{"", "ValidProperty", ""},
		},
		{
			name:               "nil properties slice",
			properties:         nil,
			expectError:        false,
			expectedProperties: nil,
		},
		{
			name:               "large number of properties",
			properties:         generateLargePropertySlice(100),
			expectError:        false,
			expectedProperties: generateLargePropertySlice(100),
		},
		{
			name:               "duplicate properties",
			properties:         []string{"Duplicate", "Property", "Duplicate"},
			expectError:        false,
			expectedProperties: []string{"Duplicate", "Property", "Duplicate"},
		},
		{
			name:               "numeric string properties",
			properties:         []string{"123", "456", "789"},
			expectError:        false,
			expectedProperties: []string{"123", "456", "789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an Engine instance
			engine := &Engine{}

			// Apply the WithProperties option
			err := WithProperties(tt.properties)(engine)

			// Check error conditions
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}

				// Check if properties were set correctly
				if tt.expectedProperties == nil {
					if engine.managerProperties != nil {
						t.Errorf("Expected managerProperties to be nil, got %v", engine.managerProperties)
					}
				} else {
					if !slicesEqual(engine.managerProperties, tt.expectedProperties) {
						t.Errorf("Expected managerProperties to be %v, got %v",
							tt.expectedProperties, engine.managerProperties)
					}
				}
			}
		})
	}
}

// Helper function to compare slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper function to generate a large slice of properties for testing
func generateLargePropertySlice(size int) []string {
	properties := make([]string, size)
	for i := 0; i < size; i++ {
		properties[i] = fmt.Sprintf("Property%d", i)
	}
	return properties
}
