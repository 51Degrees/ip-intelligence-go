package ipi_interop

//#include <string.h>
//#include "ip-intelligence-cxx.h"
import "C"

// Performance Profile
type PerformanceProfile int

const (
	Default PerformanceProfile = iota
	LowMemory
	BalancedTemp
	Balanced
	HighPerformance
	InMemory
	SingleLoaded
)

// ConfigIpi wraps around pointer to a value of C ConfigIpi structure
type ConfigIpi struct {
	CPtr *C.ConfigIpi
	perf PerformanceProfile
}

/* Constructor and Destructor */

// NewConfigIpi creates a new ConfigIpi object. Target performance profile
// is required as initial setup and any following adjustments will be done
// on top of this initial performance profile. Any invalid input will result
// in default config to be used. In C the default performance profile is the
// same as balanced profile.
func NewConfigIpi(perf PerformanceProfile) *ConfigIpi {
	var config C.ConfigIpi
	profile := perf
	switch perf {
	case InMemory:
		config = C.IpiInMemoryConfig
	case HighPerformance:
		config = C.IpiHighPerformanceConfig
	case LowMemory:
		config = C.IpiLowMemoryConfig
	case Balanced:
		config = C.IpiBalancedConfig
	case BalancedTemp:
		config = C.IpiBalancedTempConfig
	case SingleLoaded:
		config = C.fiftyoneDegreesIpiSingleLoadedConfig
	default:
		config = C.IpiDefaultConfig
		profile = Default
	}
	return &ConfigIpi{&config, profile}
}

// SetConcurrency sets the expected concurrent requests.
func (config *ConfigIpi) SetConcurrency(concurrency uint16) {
	config.CPtr.strings.concurrency = C.ushort(concurrency)
	config.CPtr.properties.concurrency = C.ushort(concurrency)
	config.CPtr.values.concurrency = C.ushort(concurrency)
	config.CPtr.profiles.concurrency = C.ushort(concurrency)
	config.CPtr.nodes.concurrency = C.ushort(concurrency)
	config.CPtr.profileOffsets.concurrency = C.ushort(concurrency)
	config.CPtr.maps.concurrency = C.ushort(concurrency)
	config.CPtr.components.concurrency = C.ushort(concurrency)
}

// SetUsePredictiveGraph sets whether predictive optimized graph should be
// used for processing.
func (config *ConfigIpi) SetUsePredictiveGraph(use bool) {
	if use {
		config.CPtr.usePredictiveGraph = C.IntToBool(1)
	} else {
		config.CPtr.usePredictiveGraph = C.IntToBool(0)
	}
}

// SetUsePerformanceGraph sets whether performance optimized graph should be
// used for processing.
func (config *ConfigIpi) SetUsePerformanceGraph(use bool) {
	if use {
		config.CPtr.usePerformanceGraph = C.IntToBool(1)
	} else {
		config.CPtr.usePerformanceGraph = C.IntToBool(0)
	}
}

// SetUseUpperPrefixHeaders set whether or not the HTTP header might be
// prefixed with 'HTTP_'
func (config *ConfigIpi) SetUseUpperPrefixHeaders(use bool) {
	if use {
		config.CPtr.b.b.usesUpperPrefixedHeaders = C.IntToBool(1)
	} else {
		config.CPtr.b.b.usesUpperPrefixedHeaders = C.IntToBool(0)
	}
}
