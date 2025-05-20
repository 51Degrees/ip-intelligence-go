package ipi_onpremise

import (
	"errors"
	"fmt"
	"github.com/51Degrees/ip-intelligence-go/ipi_interop"
	"net/url"
	"os"
	"path/filepath"
	"sync"
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
	logger logWrapper
	//fileWatcher                 fileWatcher
	dataFile            string
	licenseKey          string
	dataFileUrl         string
	dataFilePullEveryMs int
	isAutoUpdateEnabled bool
	loggerEnabled       bool
	manager             *ipi_interop.ResourceManager
	config              *ipi_interop.ConfigIpi
	//totalFilePulls              int
	stopCh     chan *sync.WaitGroup
	fileSynced bool
	//product                     string
	maxRetries int
	//lastModificationTimestamp   *time.Time
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
	managerProperties           string
}

// New creates an instance of the on-premise device detection engine.  WithDataFile must be provided
// to specify the path to the data file, otherwise initialization will fail
func New(opts ...EngineOptions) (*Engine, error) {
	engine := &Engine{
		logger: logWrapper{
			logger:  DefaultLogger,
			enabled: true,
		},
		config:     nil,
		stopCh:     make(chan *sync.WaitGroup),
		fileSynced: false,
		//dataFileUrl:                 defaultDataFileUrl,
		dataFilePullEveryMs:         30 * 60 * 1000, // default 30 minutes
		isFileWatcherEnabled:        true,
		isUpdateOnStartEnabled:      false,
		isAutoUpdateEnabled:         true,
		isCreateTempDataCopyEnabled: true,
		tempDataDir:                 "",
		randomization:               10 * 60 * 1000, // default 10 minutes
		managerProperties:           "IpRangeStart,IpRangeEnd,AccuracyRadius,RegisteredCountry,RegisteredName,Longitude,Latitude,Areas",
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

	// if file watcher is enabled, start the watcher
	if engine.isFileWatcherEnabled {
		// TODO: check and uncomment
		//engine.fileWatcher, err = newFileWatcher(engine.logger, engine.dataFile, engine.stopCh)
		//if err != nil {
		//	return nil, err
		//}
		//// this will watch the data file, if it changes, it will reload the data file in the manager
		//err = engine.fileWatcher.watch(engine.handleFileExternallyChanged)
		//if err != nil {
		//	return nil, err
		//}
		//engine.fileWatcherStarted = true
		//go engine.fileWatcher.run()
	}

	return engine, nil
}

// run starts the engine
func (e *Engine) run() error {
	err := e.processFileExternallyChanged()
	if err != nil {
		return err
	}

	//err = e.validateAndAppendUrlParams()
	//if err != nil {
	//	return err
	//}

	if e.isAutoUpdateEnabled {
		e.filePullerStarted = true
		//go e.scheduleFilePulling()
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
		//year, month, day := e.getPublishedDate().Date()
		//e.logger.Printf("data file loaded from " + filePath + " published on: " + fmt.Sprintf("%d-%d-%d", year, month, day))
	}()

	if e.manager == nil {
		e.manager = ipi_interop.NewResourceManager()
		// init manager from file
		if e.config == nil {
			e.config = ipi_interop.NewConfigIpi(ipi_interop.Balanced)
		}

		err := ipi_interop.InitManagerFromFile(e.manager, *e.config, e.managerProperties, filePath)

		if err != nil {
			return fmt.Errorf("failed to init manager from file: %w", err)
		}
		e.dataFileLastUsedByManager = filePath
		// return nil is created for the first time
		return nil
	} else if !e.isCreateTempDataCopyEnabled {
		//err := e.manager.ReloadFromOriginalFile()
		//if err != nil {
		//	return fmt.Errorf("failed to reload manager from original file: %w", err)
		//}
		return nil
	}

	err := e.manager.ReloadFromFile(*e.config, e.managerProperties, filePath)
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

// Process detects the device from the provided evidence list
// returns the dd.ResultsIpi object from which various device properties
// are retrieved
func (e *Engine) Process(ipAddress string) (*ipi_interop.ResultsIpi, error) {
	results := ipi_interop.NewResultsIpi(e.manager)

	if err := results.ResultsIpiFromIpAddress(ipAddress); err != nil {
		return nil, err
	}

	return results, nil
}

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
