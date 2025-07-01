package ipi_onpremise

import (
	"fmt"
	common_go "github.com/51Degrees/common-go"
	"github.com/51Degrees/ip-intelligence-go/ipi_interop"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Engine is an implementation of the on-premise (based on a local data file) device detection. It encapsulates
// the automatic data file updates feature - to periodically fetch and reload the new data file.
// File system watcher feature allows to monitor for changes to the local data file and reload it when it changes.
// Custom URL can be used to fetch data files, the polling interval is configurable
// The 51degrees distributor service can also be used with a licenseKey
// For more information see With... options and examples
type Engine struct {
	*common_go.FileUpdater
	logger *common_go.LogWrapper

	manager *ipi_interop.ResourceManager
	config  *ipi_interop.ConfigIpi

	stopCh           chan *sync.WaitGroup
	reloadFileEvents chan struct{}

	product                   string
	dataFileLastUsedByManager string
	licenseKey                string

	maxRetries int

	isStopped bool

	managerProperties  []string
	propertyIndexCache map[string]int // name → index mapping
	propertyNameCache  map[int]string // index → name mapping (readonly after init)
	propertyIndexes    []int
}

const (
	defaultDataFileUrl = "" // TODO: set default file path url (when it will be available)
)

var (
	defaultProperties = []string{
		"IpRangeStart", "IpRangeEnd", "AccuracyRadius", "RegisteredCountry", "RegisteredName", "Longitude", "Latitude", "Areas", "Mcc",
	}
)

// New creates an instance of the on-premise device detection engine.  WithDataFile must be provided
// to specify the path to the data file, otherwise initialization will fail
func New(opts ...EngineOptions) (*Engine, error) {
	fileUpdater := common_go.NewFileUpdater(defaultDataFileUrl)
	logger := fileUpdater.GetLogger()

	engine := &Engine{
		FileUpdater: fileUpdater,
		logger:      logger,

		config:             nil,
		stopCh:             make(chan *sync.WaitGroup),
		reloadFileEvents:   make(chan struct{}),
		managerProperties:  defaultProperties,
		propertyIndexCache: make(map[string]int),
		propertyNameCache:  make(map[int]string),
	}

	for _, opt := range opts {
		err := opt(engine)
		if err != nil {
			engine.Stop()
			return nil, err
		}
	}

	if !engine.IsDataFileProvided() {
		return nil, common_go.ErrNoDataFileProvided
	}

	if err := engine.InitCreateTempDataCopy(); err != nil {
		return nil, err
	}
	err := engine.run()
	if err != nil {
		engine.Stop()
		return nil, err
	}

	// Pre-compute property indexes using a temporary results object
	engine.initPropertyIndexes()

	// if file watcher is enabled, start the watcher
	if engine.IsFileWatcherEnabled() {
		if err := engine.InitFileWatcher(engine.logger, engine.stopCh); err != nil {
			return nil, err
		}

		if err := engine.Watch(engine.handleFileExternallyChanged); err != nil {
			return nil, err
		}

		engine.SetFileWatcherStarted(true)
		engine.RunWatcher()
	}

	return engine, nil
}

func (e *Engine) handleFileExternallyChanged() {
	if err := e.processFileExternallyChanged(); err != nil {
		e.logger.Printf("failed to handle file externally changed: %v", err)
	}

	e.IncreaseFileExternallyChangedCount()
}

// run starts the engine
func (e *Engine) run() error {
	e.recoverEngine()

	go e.reloadFileEvent()

	if err := e.processFileExternallyChanged(); err != nil {
		return err
	}

	if err := e.validateAndAppendUrlParams(); err != nil {
		return err
	}

	if e.IsAutoUpdateEnabled() {
		e.SetFilePullerStarted(true)
		go e.ScheduleFilePulling(e.stopCh, e.reloadFileEvents)
	}

	return nil
}

// Stop has to be called to free all the resources of the engine
// before the instance goes out of scope
func (e *Engine) Stop() {
	num := 0
	if e.IsAutoUpdateEnabled() && e.IsFilePullerStarted() {
		num++ // file puller is enabled and started
	}
	if e.IsFileWatcherEnabled() && e.IsFileWatcherStarted() {
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
	close(e.reloadFileEvents)

	if e.manager != nil {
		e.manager.Free()
	} else {
		e.logger.Printf("stopping engine, manager is nil")
	}

	if e.IsCreateTempDataCopyEnabled() {
		dir := filepath.Dir(e.dataFileLastUsedByManager)
		os.RemoveAll(dir)
	}
}

func (e *Engine) recoverEngine() {
	// recover from panic
	// if panic occurs, we will log the error and restart the file pulling
	defer func() {
		if r := recover(); r != nil {
			e.logger.Printf("error occurred when pulling data: %v", r)
			if !e.isStopped {
				go e.ScheduleFilePulling(e.stopCh, e.reloadFileEvents)
			}
		}
	}()
}

func (e *Engine) reloadFileEvent() {
	for range e.reloadFileEvents {
		if err := e.processFileExternallyChanged(); err != nil {
			return
		}
	}
}

func (e *Engine) validateAndAppendUrlParams() error {
	if e.isDefaultDataFileUrl() && !e.hasDefaultDistributorParams() && e.IsAutoUpdateEnabled() {
		return common_go.ErrLicenseKeyRequired
	}

	if e.isDefaultDataFileUrl() && e.IsAutoUpdateEnabled() {
		if err := e.appendLicenceKey(); err != nil {
			return err
		}

		if err := e.appendProduct(); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) appendProduct() error {
	urlParsed, err := url.Parse(e.GetDataFileUrl())
	if err != nil {
		return fmt.Errorf("failed to parse data file url: %w", err)
	}
	query := urlParsed.Query()
	query.Set("Product", e.product)
	urlParsed.RawQuery = query.Encode()

	e.SetDataFileUrl(urlParsed.String())

	return nil
}

func (e *Engine) isDefaultDataFileUrl() bool {
	return e.GetDataFileUrl() == defaultDataFileUrl
}

func (e *Engine) hasDefaultDistributorParams() bool {
	return len(e.licenseKey) > 0
}

func (e *Engine) processFileExternallyChanged() error {
	reloadFilePath, err := e.GetReloadFilePath()
	if err != nil {
		return err
	}

	if err := e.reloadManager(reloadFilePath); err != nil {
		return err
	}

	return nil
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
	} else if !e.IsCreateTempDataCopyEnabled() {
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

// initPropertyIndexes pre-computes and caches bidirectional property index↔name mappings
func (e *Engine) initPropertyIndexes() {
	e.propertyIndexes = make([]int, len(e.managerProperties))

	// Create a temporary results object to get property indexes
	tempResults := ipi_interop.NewResultsIpi(e.manager)
	defer tempResults.Free()

	for i, prop := range e.managerProperties {
		idx := e.getPropertyIndex(tempResults, prop)
		e.propertyIndexes[i] = idx
		// Cache bidirectional mappings: name ↔ index
		e.propertyIndexCache[prop] = idx
		e.propertyNameCache[idx] = prop
	}
}

// getPropertyIndex gets the property index from results
func (e *Engine) getPropertyIndex(results *ipi_interop.ResultsIpi, propertyName string) int {
	// Use the existing method from results_ipi.go
	return results.GetPropertyIndexByName(propertyName)
}

// GetPropertyNameByIndex retrieves the property name for a given index from the engine's cache
// This is thread-safe as the cache is readonly after initialization
func (e *Engine) GetPropertyNameByIndex(index int) string {
	if name, exists := e.propertyNameCache[index]; exists {
		return name
	}
	return "" // Unknown index
}

// NewResultsIpi creates a new ResultsIpi object using this engine's manager
// Caller is responsible for calling Free() on the returned object
func (e *Engine) NewResultsIpi() *ipi_interop.ResultsIpi {
	return ipi_interop.NewResultsIpi(e.manager)
}

// Process processes the given IP address and retrieves associated values using the default properties.
// If results is nil, creates a new ResultsIpi object for this call (per-call mode).
// If results is provided, reuses the existing object (reuse mode for better performance).
func (e *Engine) Process(ipAddress string) (ipi_interop.Values, error) {
	return e.ProcessWithResults(ipAddress, nil)
}

// ProcessWithResults processes the given IP address with an optional reusable ResultsIpi object.
// If results is nil, creates a new ResultsIpi object for this call.
// If results is provided, reuses it for better performance (caller manages lifecycle).
func (e *Engine) ProcessWithResults(ipAddress string, results *ipi_interop.ResultsIpi) (ipi_interop.Values, error) {
	var shouldFree bool

	if results == nil {
		// Create a new ResultsIpi object for this call
		results = ipi_interop.NewResultsIpi(e.manager)
		shouldFree = true // We created it, so we should free it
	}

	if shouldFree {
		defer results.Free() // Ensure proper cleanup only if we created it
	}

	if err := results.ResultsIpiFromIpAddress(ipAddress); err != nil {
		return nil, err
	}

	var values ipi_interop.Values
	var err error

	if results.HasValues() {
		// OPTIMIZATION: Use pre-computed indexes with Engine's bidirectional property mapping
		// This eliminates expensive index→name CGO calls by using Engine's readonly cache
		values, err = results.GetWeightedValuesByIndexes(e.propertyIndexes, e.GetPropertyNameByIndex)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

// appendLicenceKey appends the license key as a query parameter to the data file URL in the Engine instance.
func (e *Engine) appendLicenceKey() error {
	urlParsed, err := url.Parse(e.GetDataFileUrl())
	if err != nil {
		return err
	}
	query := urlParsed.Query()
	query.Set("LicenseKeys", e.licenseKey)
	urlParsed.RawQuery = query.Encode()

	e.SetDataFileUrl(urlParsed.String())

	return nil
}

// getPublishedDate retrieves the published date of the data file being used by the engine.
func (e *Engine) getPublishedDate() time.Time {
	return ipi_interop.GetPublishedDate(e.manager)
}
