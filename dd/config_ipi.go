package dd

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
