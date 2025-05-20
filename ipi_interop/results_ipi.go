package ipi_interop

//#include <string.h>
//#include "ip-intelligence-cxx.h"
import "C"
import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

type ResultsIpi struct {
	CPtr     *C.ResultsIpi
	CResults *interface{} // Pointer to a slice holding C results
}

const defaultSize = 4096
const separator = "|"
const regexPatter = "^\"([^\"]+)\":(([0-9]*[.])?[0-9]+)$"

func NewResultsIpi(manager *ResourceManager) *ResultsIpi {
	r := C.ResultsIpiCreate(manager.CPtr)

	var cResults interface{} = (*[math.MaxInt32 / int(C.sizeof_ResultIpi)]C.ResultIpi)(unsafe.Pointer(r.items))[:r.capacity:r.capacity]

	res := &ResultsIpi{r, &cResults}
	runtime.SetFinalizer(res, resultsFinalizer)

	return res
}

func (r *ResultsIpi) ResultsIpiFromIpAddress(ipAddress string) error {
	exception := NewException()

	char := C.CString(ipAddress)
	defer C.free(unsafe.Pointer(char))

	C.ResultsIpiFromIpAddressString(
		r.CPtr,
		char,
		C.strlen(char),
		exception.CPtr,
	)

	if !exception.IsOkay() {
		return fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
	}

	return nil
}

func (r *ResultsIpi) HasValues() bool {
	return r.CPtr != nil && r.CPtr.count > 0
}

// Free free the resource allocated in the C layer.
func (results *ResultsIpi) Free() {
	if results.CPtr != nil {
		C.ResultsIpiFree(results.CPtr)
		results.CPtr = nil
	}
}

// resultsFinalizer check if C resource has been explicitly
// freed by Free method. Panic if it was not.
func resultsFinalizer(res *ResultsIpi) {
	if res.CPtr != nil {
		panic("ERROR: ResultsIpi should be freed explicitly by its Free method.")
	}
}

func GetPropertyValueAsRaw(result *C.ResultsIpi, property string) (string, error) {
	var buffer []C.char

	buffer = make([]C.char, defaultSize)

	propertyName := C.CString(property)
	defer C.free(unsafe.Pointer(propertyName))

	cSeparator := C.CString(separator)
	defer C.free(unsafe.Pointer(cSeparator))

	exception := NewException()

	actualSize := uint64(C.ResultsIpiGetValuesString(result, propertyName, &buffer[0], C.size_t(defaultSize), cSeparator, exception.CPtr))
	if !exception.IsOkay() {
		return "", fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
	}

	var ds uint64 = defaultSize

	if actualSize > ds {
		// Add 1 for the null terminator
		ds = actualSize + 1
	}

	return C.GoString(&buffer[0]), nil
}

func GetPropertyValueAsStringWeightValue(result *C.ResultsIpi, property string) (string, float64, error) {
	var buffer []C.char

	buffer = make([]C.char, defaultSize)

	propertyName := C.CString(property)
	defer C.free(unsafe.Pointer(propertyName))

	cSeparator := C.CString(separator)
	defer C.free(unsafe.Pointer(cSeparator))

	exception := NewException()

	actualSize := uint64(C.ResultsIpiGetValuesString(result, propertyName, &buffer[0], C.size_t(defaultSize), cSeparator, exception.CPtr))
	if !exception.IsOkay() {
		return "", 0, fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
	}

	var ds uint64 = defaultSize

	if actualSize > ds {
		// Add 1 for the null terminator
		ds = actualSize + 1
	}

	r, err := regexp.Compile(regexPatter)
	if err != nil {
		return "", 0, err
	}

	match := r.FindStringSubmatch(C.GoString(&buffer[0]))

	if len(match) < 3 {
		return "", 0, errors.New("Invalid regex pattern")
	}

	weight, err := strconv.ParseFloat(strings.TrimSpace(match[2]), 64)
	if err != nil {
		return "", 0, err
	}

	return match[1], weight, nil
}
