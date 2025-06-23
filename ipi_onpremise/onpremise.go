package ipi_onpremise

import (
	"errors"
	"fmt"
	"github.com/51Degrees/ip-intelligence-go/ipi_interop"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	ErrNoDataFileProvided = errors.New("no data file provided")
	ErrTooManyRetries     = errors.New("too many retries to pull data file")
	ErrFileNotModified    = errors.New("data file not modified")
	ErrLicenseKeyRequired = errors.New("auto update set to true, no custom URL specified, license key is required, set it using WithLicenseKey")
)

// Engine is an implementation of the on-premise (based on a local data file) device detection. It encapsulates
// the automatic data file updates feature - to periodically fetch and reload the new data file.
// File system watcher feature allows to monitor for changes to the local data file and reload it when it changes.
// Custom URL can be used to fetch data files, the polling interval is configurable
// The 51degrees distributor service can also be used with a licenseKey
// For more information see With... options and examples
type Engine struct {
	logger                      logWrapper
	fileWatcher                 fileWatcher
	dataFile                    string
	licenseKey                  string
	dataFileUrl                 string
	dataFilePullEveryMs         int
	isAutoUpdateEnabled         bool
	loggerEnabled               bool
	manager                     *ipi_interop.ResourceManager
	config                      *ipi_interop.ConfigIpi
	totalFilePulls              int
	stopCh                      chan *sync.WaitGroup
	fileSynced                  bool
	product                     string
	maxRetries                  int
	lastModificationTimestamp   *time.Time
	isFileWatcherEnabled        bool
	isUpdateOnStartEnabled      bool
	isCreateTempDataCopyEnabled bool
	tempDataFile                string
	tempDataDir                 string
	dataFileLastUsedByManager   string
	isCopyingFile               bool
	randomization               int
	isStopped                   bool
	fileExternallyChangedCount  int
	filePullerStarted           bool
	fileWatcherStarted          bool
	managerProperties           []string
	resultsPool                 chan *ipi_interop.ResultsIpi // Pool of pre-allocated ResultsIpi objects
	propertyIndexCache          map[string]int
	propertyIndexes             []int
}

const (
	defaultDataFileUrl = "" // TODO: Fix url for update
)

var (
	defaultProperties = []string{
		"IpRangeStart", "IpRangeEnd", "AccuracyRadius", "RegisteredCountry", "RegisteredName", "Longitude", "Latitude", "Areas", "Mcc",
	}
)

// New creates an instance of the on-premise device detection engine.  WithDataFile must be provided
// to specify the path to the data file, otherwise initialization will fail
func New(opts ...EngineOptions) (*Engine, error) {
	engine := &Engine{
		logger: logWrapper{
			logger:  DefaultLogger,
			enabled: true,
		},
		config:                      nil,
		stopCh:                      make(chan *sync.WaitGroup),
		fileSynced:                  false,
		dataFileUrl:                 defaultDataFileUrl,
		dataFilePullEveryMs:         30 * 60 * 1000, // default 30 minutes
		isFileWatcherEnabled:        true,
		isUpdateOnStartEnabled:      false,
		isAutoUpdateEnabled:         true,
		isCreateTempDataCopyEnabled: true,
		tempDataDir:                 "",
		randomization:               10 * 60 * 1000, // default 10 minutes
		managerProperties:           defaultProperties,
		propertyIndexCache:          make(map[string]int),
	}

	for _, opt := range opts {
		err := opt(engine)
		if err != nil {
			engine.Stop()
			return nil, err
		}
	}

	if engine.dataFile == "" {
		return nil, ErrNoDataFileProvided
	}

	if engine.isCreateTempDataCopyEnabled && engine.tempDataDir == "" {
		path, err := os.MkdirTemp("", "51degrees-on-premise")
		if err != nil {
			return nil, err
		}
		engine.tempDataDir = path
	}

	err := engine.run()
	if err != nil {
		engine.Stop()
		return nil, err
	}

	// Pre-compute property indexes using a temporary results object
	engine.initPropertyIndexes()

	// Initialize pool of ResultsIpi objects
	engine.initResultsPool()

	// if file watcher is enabled, start the watcher
	if engine.isFileWatcherEnabled {
		engine.fileWatcher, err = newFileWatcher(engine.logger, engine.dataFile, engine.stopCh)
		if err != nil {
			return nil, err
		}
		// this will watch the data file, if it changes, it will reload the data file in the manager
		err = engine.fileWatcher.watch(engine.handleFileExternallyChanged)
		if err != nil {
			return nil, err
		}
		engine.fileWatcherStarted = true
		go engine.fileWatcher.run()
	}

	return engine, nil
}

func (e *Engine) handleFileExternallyChanged() {
	err := e.processFileExternallyChanged()
	if err != nil {
		e.logger.Printf("failed to handle file externally changed: %v", err)
	}
	e.fileExternallyChangedCount++
}

// run starts the engine
func (e *Engine) run() error {
	err := e.processFileExternallyChanged()
	if err != nil {
		return err
	}

	err = e.validateAndAppendUrlParams()
	if err != nil {
		return err
	}

	if e.isAutoUpdateEnabled {
		e.filePullerStarted = true
		go e.scheduleFilePulling()
	}

	return nil
}

// Stop has to be called to free all the resources of the engine
// before the instance goes out of scope
func (e *Engine) Stop() {
	num := 0
	if e.isAutoUpdateEnabled && e.filePullerStarted {
		num++ // file puller is enabled and started
	}
	if e.isFileWatcherEnabled && e.fileWatcherStarted {
		num++ // file watcher is enabled and started
	}

	if num > 0 {
		var wg sync.WaitGroup
		wg.Add(num)
		for i := 0; i < num; i++ {
			e.stopCh <- &wg
		}
		// make sure that all routines finished processing current work, only after that free the manager
		wg.Wait()
	}

	e.isStopped = true
	close(e.stopCh)

	// Free all ResultsIpi objects in the pool before freeing the manager
	if e.resultsPool != nil {
		close(e.resultsPool)
		for results := range e.resultsPool {
			results.Free()
		}
	}

	if e.manager != nil {
		e.manager.Free()
	} else {
		e.logger.Printf("stopping engine, manager is nil")
	}

	if e.isCreateTempDataCopyEnabled {
		dir := filepath.Dir(e.dataFileLastUsedByManager)
		os.RemoveAll(dir)
	}
}

func (e *Engine) validateAndAppendUrlParams() error {
	if e.isDefaultDataFileUrl() && !e.hasDefaultDistributorParams() && e.isAutoUpdateEnabled {
		return ErrLicenseKeyRequired
	} else if e.isDefaultDataFileUrl() && e.isAutoUpdateEnabled {
		err := e.appendLicenceKey()
		if err != nil {
			return err
		}
		err = e.appendProduct()
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) appendProduct() error {
	urlParsed, err := url.Parse(e.dataFileUrl)
	if err != nil {
		return fmt.Errorf("failed to parse data file url: %w", err)
	}
	query := urlParsed.Query()
	query.Set("Product", e.product)
	urlParsed.RawQuery = query.Encode()

	e.dataFileUrl = urlParsed.String()

	return nil
}

func (e *Engine) isDefaultDataFileUrl() bool {
	return e.dataFileUrl == defaultDataFileUrl
}

func (e *Engine) hasDefaultDistributorParams() bool {
	return len(e.licenseKey) > 0
}

func (e *Engine) copyFileAndReloadManager() error {
	dirPath, tempFilepath, err := e.copyToTempFile()
	if err != nil {
		return err
	}
	fullPath := filepath.Join(dirPath, tempFilepath)
	err = e.reloadManager(fullPath)
	if err != nil {
		return err
	}
	e.tempDataFile = tempFilepath

	return nil
}

func (e *Engine) processFileExternallyChanged() error {
	if e.isCreateTempDataCopyEnabled {
		err := e.copyFileAndReloadManager()
		if err != nil {
			return err
		}
	} else {
		err := e.reloadManager(e.dataFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) copyToTempFile() (string, string, error) {
	data, err := os.ReadFile(e.dataFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read data file: %w", err)
	}
	originalFileName := filepath.Base(e.dataFile)

	f, err := os.CreateTemp(e.tempDataDir, originalFileName)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp data file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return "", "", fmt.Errorf("failed to write temp data file: %w", err)
	}

	tempFileName := filepath.Base(f.Name())
	return e.tempDataDir, tempFileName, nil
}

// this function will be called when the engine is started or the is new file available
// it will create and initialize a new manager from the new file if it does not exist
// if the manager exists, it will create a new manager from the new file and replace the existing manager thus freeing memory of the old manager
func (e *Engine) reloadManager(filePath string) error {
	if e.isStopped {
		return nil
	}
	// if manager is nil, create a new one
	defer func() {
		year, month, day := e.getPublishedDate().Date()
		e.logger.Printf("data file loaded from " + filePath + " published on: " + fmt.Sprintf("%d-%d-%d", year, month, day))
	}()

	if e.manager == nil {
		e.manager = ipi_interop.NewResourceManager()
		// init manager from file
		if e.config == nil {
			e.config = ipi_interop.NewConfigIpi(ipi_interop.Balanced)
		}

		if err := ipi_interop.InitManagerFromFile(e.manager, *e.config, strings.Join(e.managerProperties, ","), filePath); err != nil {
			return fmt.Errorf("failed to init manager from file: %+v", err)
		}
		e.dataFileLastUsedByManager = filePath
		// return nil is created for the first time
		return nil
	} else if !e.isCreateTempDataCopyEnabled {
		err := e.manager.ReloadFromOriginalFile()
		if err != nil {
			return fmt.Errorf("failed to reload manager from original file: %w", err)
		}
		return nil
	}

	err := e.manager.ReloadFromFile(*e.config, strings.Join(e.managerProperties, ","), filePath)
	if err != nil {
		return fmt.Errorf("failed to reload manager from file: %w", err)
	}

	err = os.Remove(e.dataFileLastUsedByManager)
	if err != nil {
		return err
	}

	e.dataFileLastUsedByManager = filePath

	return nil
}

// initPropertyIndexes pre-computes and caches property indexes
func (e *Engine) initPropertyIndexes() {
	e.propertyIndexes = make([]int, len(e.managerProperties))
	
	// Create a temporary results object to get property indexes
	tempResults := ipi_interop.NewResultsIpi(e.manager)
	defer tempResults.Free()
	
	for i, prop := range e.managerProperties {
		idx := e.getPropertyIndex(tempResults, prop)
		e.propertyIndexes[i] = idx
		e.propertyIndexCache[prop] = idx
	}
}

// initResultsPool creates a pool of pre-allocated ResultsIpi objects
func (e *Engine) initResultsPool() {
	// Create a pool with size based on CPU count * 2 for good concurrency
	poolSize := runtime.NumCPU() * 2
	e.resultsPool = make(chan *ipi_interop.ResultsIpi, poolSize)
	
	// Pre-allocate ResultsIpi objects
	for i := 0; i < poolSize; i++ {
		results := ipi_interop.NewResultsIpi(e.manager)
		e.resultsPool <- results
	}
}

// getPropertyIndex gets the property index from results
func (e *Engine) getPropertyIndex(results *ipi_interop.ResultsIpi, propertyName string) int {
	// Use the existing method from results_ipi.go
	return results.GetPropertyIndexByName(propertyName)
}

// Process processes the given IP address and retrieves associated values using the default properties.
func (e *Engine) Process(ipAddress string) (ipi_interop.Values, error) {
	// Get a ResultsIpi object from the pool
	results := <-e.resultsPool
	// Return it to the pool when done
	defer func() {
		e.resultsPool <- results
	}()
	
	if err := results.ResultsIpiFromIpAddress(ipAddress); err != nil {
		return nil, err
	}

	var values ipi_interop.Values
	var err error

	if results.HasValues() {
		// Use pre-computed indexes instead of property names
		values, err = results.GetWeightedValuesByIndexes(e.propertyIndexes)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

// appendLicenceKey appends the license key as a query parameter to the data file URL in the Engine instance.
func (e *Engine) appendLicenceKey() error {
	urlParsed, err := url.Parse(e.dataFileUrl)
	if err != nil {
		return err
	}
	query := urlParsed.Query()
	query.Set("LicenseKeys", e.licenseKey)
	urlParsed.RawQuery = query.Encode()

	e.dataFileUrl = urlParsed.String()

	return nil
}

// getFilePath returns the file path of the data file or its temporary copy depending on configuration settings.
func (e *Engine) getFilePath() string {
	if e.isCreateTempDataCopyEnabled {
		return filepath.Join(e.tempDataDir, e.tempDataFile)
	}

	return e.dataFile
}

// getPublishedDate retrieves the published date of the data file being used by the engine.
func (e *Engine) getPublishedDate() time.Time {
	return ipi_interop.GetPublishedDate(e.manager)
}
