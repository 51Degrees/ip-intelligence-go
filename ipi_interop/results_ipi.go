/* *********************************************************************
 * This Original Work is copyright of 51 Degrees Mobile Experts Limited.
 * Copyright 2019 51 Degrees Mobile Experts Limited, 5 Charlotte Close,
 * Caversham, Reading, Berkshire, United Kingdom RG4 7BY.
 *
 * This Original Work is licensed under the European Union Public Licence (EUPL)
 * v.1.2 and is subject to its terms as set out below.
 *
 * If a copy of the EUPL was not distributed with this file, You can obtain
 * one at https://opensource.org/licenses/EUPL-1.2.
 *
 * The 'Compatible Licences' set out in the Appendix to the EUPL (as may be
 * amended by the European Commission) shall be deemed incompatible for
 * the purposes of the Work and the provisions of the compatibility
 * clause in Article 5 of the EUPL shall not apply.
 *
 * If using the Work as, or as part of, a network application, by
 * including the attribution notice(s) required under Article 5 of the EUPL
 * in the end user terms of the application under an appropriate heading,
 * such notice(s) shall fulfill the requirements of that article.
 * ********************************************************************* */
package ipi_interop

//#include <string.h>
//#include "ip-intelligence-cxx.h"
import "C"
import (
	"fmt"
	"math"
	"runtime"
	"unsafe"
)

// uint16Max represents the maximum value of a uint16 converted to a float64.
var uint16Max = float64(C.UINT16_MAX)

// ResultsIpi represents a structure to manage IP-related results in the C library.
// It contains a pointer to the C.ResultsIpi structure and a dynamic C result slice.
type ResultsIpi struct {
	CPtr     *C.ResultsIpi
	CResults *interface{} // Pointer to a slice holding C results
}

// NewResultsIpi creates a new ResultsIpi instance using the provided ResourceManager.
// The instance handles C.ResultsIpi creation and associated memory management.
// A finalizer is set to ensure resources are explicitly freed.
func NewResultsIpi(manager *ResourceManager) *ResultsIpi {
	r := C.ResultsIpiCreate(manager.CPtr)

	var cResults interface{} = (*[math.MaxInt32 / int(C.sizeof_ResultIpi)]C.ResultIpi)(unsafe.Pointer(r.items))[:r.capacity:r.capacity]

	res := &ResultsIpi{r, &cResults}
	runtime.SetFinalizer(res, resultsFinalizer)

	return res
}

// ResultsIpiFromIpAddress processes the given IP address and populates the ResultsIpi instance with related data.
// Returns an error if the operation fails.
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

// HasValues checks if the ResultsIpi instance contains valid results by verifying whether the C pointer is non-nil and count > 0.
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

// GetPropertyIndexByName retrieves the index of a property by its name from the associated dataset in the C library.
func (r *ResultsIpi) GetPropertyIndexByName(propertyName string) int {
	dataSet := (*C.DataSetIpi)(r.CPtr.b.dataSet)

	cName := C.CString(propertyName)
	defer C.free(unsafe.Pointer(cName))

	i := C.PropertiesGetRequiredPropertyIndexFromName(dataSet.b.b.available, cName)

	return int(i)
}

// getRequiredPropertyIndexFromName retrieves the index of a required property by its name from the associated dataset.
func (r *ResultsIpi) getRequiredPropertyIndexFromName(propertyName string) int {
	dataSet := (*C.DataSetIpi)(r.CPtr.b.dataSet)

	cName := C.CString(propertyName)
	defer C.free(unsafe.Pointer(cName))

	i := C.PropertiesGetRequiredPropertyIndexFromName(dataSet.b.b.available, cName)

	return int(i)
}

// getPropertiesIndexes returns a slice of indexes for the given property names by mapping each to its required property index.
func (r *ResultsIpi) getPropertiesIndexes(properties []string) []int {
	indexes := make([]int, 0, len(properties))
	for _, property := range properties {
		indexes = append(indexes, r.getRequiredPropertyIndexFromName(property))
	}

	return indexes
}

// getCollectionByProperties retrieves a weighted values collection from the C library based on specified property names.
// Returns the collection and an error if the operation fails.
func (r *ResultsIpi) getCollectionByProperties(properties []string) (C.fiftyoneDegreesWeightedValuesCollection, error) {
	exception := NewException()

	indexes := r.getPropertiesIndexes(properties)

	var cIndexes *C.int
	var cIndexesCount C.uint

	if len(indexes) > 0 {
		// Allocate C memory for the array
		cIndexes = (*C.int)(C.malloc(C.size_t(len(indexes)) * C.size_t(unsafe.Sizeof(C.int(0)))))
		defer C.free(unsafe.Pointer(cIndexes))

		// Copy Go slice elements to C array
		cSlice := unsafe.Slice(cIndexes, len(indexes))
		for i, idx := range indexes {
			cSlice[i] = C.int(idx)
		}

		cIndexesCount = C.uint(len(indexes))
	}

	collection := C.fiftyoneDegreesResultsIpiGetValuesCollection(
		r.CPtr,
		cIndexes,
		cIndexesCount,
		nil, exception.CPtr,
	)
	if !exception.IsOkay() {
		return collection, fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
	}

	return collection, nil
}

// getPropertyNameSafe retrieves the property name associated with a required index from the given dataset safely.
// Returns an empty string if the required index is invalid or out of bounds.
func (r *ResultsIpi) getPropertyNameSafe(dataSet *C.DataSetIpi, requiredIndex C.int) string {
	if requiredIndex < 0 || requiredIndex >= C.int(dataSet.b.b.available.count) {
		return ""
	}

	res := C.fiftyoneDegreesPropertiesGetNameFromRequiredIndex(dataSet.b.b.available, requiredIndex)
	if res != nil {
		return C.GoString(&res.value)
	}

	return ""
}

// header represents the structure holding type, required property index, and raw weighting for a weighted value.
type header struct {
	valueType             C.fiftyoneDegreesPropertyValueType
	requiredPropertyIndex C.int
	rawWeighting          C.uint16_t
}

// GetWeightedValues retrieves weighted values for the specified properties from the associated dataset.
// Returns the constructed Values map and an error if the operation fails.
func (r *ResultsIpi) GetWeightedValues(properties []string) (Values, error) {
	dataSet := (*C.DataSetIpi)(r.CPtr.b.dataSet)

	collection, err := r.getCollectionByProperties(properties)
	if err != nil {
		return nil, err
	}
	// Release the collection
	defer C.fiftyoneDegreesWeightedValuesCollectionRelease(&collection)

	values := make(Values, collection.itemsCount)

	// Create a Go slice from the C array
	headers := unsafe.Slice(collection.items, collection.itemsCount)
	for _, h := range headers {
		nextHeader := (*header)(unsafe.Pointer(h))

		requiredPropertyIndex := nextHeader.requiredPropertyIndex

		propName := r.getPropertyNameSafe(dataSet, requiredPropertyIndex)

		values.InitProperty(propName)

		var val interface{}

		// Process based on value type
		switch PropertyValueType(nextHeader.valueType) {
		case IntegerValueType:
			// Cast to weighted integer and get value
			weightedInt := (*C.fiftyoneDegreesWeightedInt)(unsafe.Pointer(nextHeader))
			val = int(weightedInt.value)

		case FloatValueType:
		case DoubleValueType:
			// Cast to weighted double and get value
			weightedDouble := (*C.fiftyoneDegreesWeightedDouble)(unsafe.Pointer(nextHeader))
			val = float64(weightedDouble.value)

		case BooleanValueType:
			// Cast to weighted boolean and get value
			weightedBool := (*C.fiftyoneDegreesWeightedBool)(unsafe.Pointer(nextHeader))
			val = weightedBool.value

		case ByteValueType:
			// Cast to weighted byte and get value
			weightedByte := (*C.fiftyoneDegreesWeightedByte)(unsafe.Pointer(nextHeader))
			val = int(weightedByte.value)

		case StringValueType:
			fallthrough
		default:
			// Cast to weighted string and get value
			weightedString := (*C.fiftyoneDegreesWeightedString)(unsafe.Pointer(nextHeader))
			val = C.GoString(weightedString.value)
		}

		// Calculate weight
		weight := float64(nextHeader.rawWeighting) / uint16Max

		// append values to the map
		values.Append(propName, val, weight)
	}

	return values, nil
}

// GetWeightedValuesByIndexes retrieves weighted values using pre-computed property indexes.
// This is more efficient than GetWeightedValues as it avoids repeated string conversions and lookups.
func (r *ResultsIpi) GetWeightedValuesByIndexes(indexes []int) (Values, error) {
	dataSet := (*C.DataSetIpi)(r.CPtr.b.dataSet)
	exception := NewException()

	var cIndexes *C.int
	var cIndexesCount C.uint

	if len(indexes) > 0 {
		// Allocate C memory for the array
		cIndexes = (*C.int)(C.malloc(C.size_t(len(indexes)) * C.size_t(unsafe.Sizeof(C.int(0)))))
		defer C.free(unsafe.Pointer(cIndexes))

		// Copy Go slice elements to C array
		cSlice := unsafe.Slice(cIndexes, len(indexes))
		for i, idx := range indexes {
			cSlice[i] = C.int(idx)
		}

		cIndexesCount = C.uint(len(indexes))
	}

	collection := C.fiftyoneDegreesResultsIpiGetValuesCollection(
		r.CPtr,
		cIndexes,
		cIndexesCount,
		nil, exception.CPtr,
	)
	if !exception.IsOkay() {
		return nil, fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
	}
	
	// Release the collection
	defer C.fiftyoneDegreesWeightedValuesCollectionRelease(&collection)

	values := make(Values, collection.itemsCount)

	// Create a Go slice from the C array
	headers := unsafe.Slice(collection.items, collection.itemsCount)
	for _, h := range headers {
		nextHeader := (*header)(unsafe.Pointer(h))

		requiredPropertyIndex := nextHeader.requiredPropertyIndex

		propName := r.getPropertyNameSafe(dataSet, requiredPropertyIndex)

		values.InitProperty(propName)

		var val interface{}

		// Process based on value type
		switch PropertyValueType(nextHeader.valueType) {
		case IntegerValueType:
			// Cast to weighted integer and get value
			weightedInt := (*C.fiftyoneDegreesWeightedInt)(unsafe.Pointer(nextHeader))
			val = int(weightedInt.value)

		case FloatValueType:
		case DoubleValueType:
			// Cast to weighted double and get value
			weightedDouble := (*C.fiftyoneDegreesWeightedDouble)(unsafe.Pointer(nextHeader))
			val = float64(weightedDouble.value)

		case BooleanValueType:
			// Cast to weighted boolean and get value
			weightedBool := (*C.fiftyoneDegreesWeightedBool)(unsafe.Pointer(nextHeader))
			val = weightedBool.value

		case ByteValueType:
			// Cast to weighted byte and get value
			weightedByte := (*C.fiftyoneDegreesWeightedByte)(unsafe.Pointer(nextHeader))
			val = int(weightedByte.value)

		case StringValueType:
			fallthrough
		default:
			// Cast to weighted string and get value
			weightedString := (*C.fiftyoneDegreesWeightedString)(unsafe.Pointer(nextHeader))
			val = C.GoString(weightedString.value)
		}

		// Calculate weight
		weight := float64(nextHeader.rawWeighting) / uint16Max

		// append values to the map
		values.Append(propName, val, weight)
	}

	return values, nil
}
