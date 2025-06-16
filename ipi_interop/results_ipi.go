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
	"sort"
	"unsafe"
)

type ResultsIpi struct {
	CPtr     *C.ResultsIpi
	CResults *interface{} // Pointer to a slice holding C results
}

const defaultSize = 4096

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

// getPropertyIndexByName retrieves the index of a property by its name from the dataset associated with the ResultsIpi instance.
func (r *ResultsIpi) getPropertyIndexByName(propertyName string) int {
	dataSet := (*C.DataSetIpi)(r.CPtr.b.dataSet)

	cName := C.CString(propertyName)
	defer C.free(unsafe.Pointer(cName))

	i := C.PropertiesGetRequiredPropertyIndexFromName(dataSet.b.b.available, cName)

	return int(i)
}

// RequiredPropertyIndexFromName retrieves the index of a required property by its name from the associated dataset.
func (r *ResultsIpi) RequiredPropertyIndexFromName(propertyName string) int {
	dataSet := (*C.DataSetIpi)(r.CPtr.b.dataSet)

	cName := C.CString(propertyName)
	defer C.free(unsafe.Pointer(cName))

	i := C.PropertiesGetRequiredPropertyIndexFromName(dataSet.b.b.available, cName)

	return int(i)
}

// GetValuesByProperty retrieves a sorted list of WeightedValue by a specified property name from the ResultsIpi instance.
func (r *ResultsIpi) GetValuesByProperty(requiredProperty string) ([]*WeightedValue, error) {
	vv := make([]*WeightedValue, 0, 0)

	dataSet := (*C.DataSetIpi)(r.CPtr.b.dataSet)

	requiredPropertyIndex := r.getPropertyIndexByName(requiredProperty)

	propertyIndex := C.PropertiesGetPropertyIndexFromRequiredIndex(dataSet.b.b.available, C.int(requiredPropertyIndex))

	if propertyIndex >= 0 {
		exception := NewException()

		storedValueType := C.PropertyGetStoredTypeByIndex(dataSet.propertyTypes, C.uint(propertyIndex), exception.CPtr)
		if !exception.IsOkay() {
			return vv, fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
		}

		exception.Clear()

		exposedValueType := C.PropertyGetValueType(dataSet.properties, C.uint(propertyIndex), exception.CPtr)
		if !exception.IsOkay() {
			return vv, fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
		}

		exception.Clear()

		valuesItems := C.ResultsIpiGetValues(r.CPtr, C.int(requiredPropertyIndex), exception.CPtr)

		if valuesItems == nil {
			return vv, nil
		}

		if !exception.IsOkay() {
			return vv, fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
		}

		goSlice := unsafe.Slice(valuesItems, r.CPtr.values.count)
		for _, v := range goSlice {
			var err error
			var val interface{}

			storedBinaryValue := (*C.StoredBinaryValue)(unsafe.Pointer(v.item.data.ptr))

			valueType := PropertyValueType(exposedValueType)

			switch valueType {
			case IntegerValueType: // int
				val = valueType.GetIntegerValue(storedBinaryValue, storedValueType)
			case DoubleValueType: // float
			case FloatValueType:
				val = valueType.GetFloatValue(storedBinaryValue, storedValueType)
			case BooleanValueType: // bool
				val = valueType.GetBooleanValue(storedBinaryValue, storedValueType)
			case StringValueType: // string
				val = valueType.GetStringValue(storedBinaryValue)
			default:
				if val, err = valueType.GetIpAddressValue(storedBinaryValue, storedValueType); err != nil {
					return vv, err
				}
			}

			vv = append(vv, &WeightedValue{
				Value:  val,
				Weight: float64(v.rawWeighting) / 65535.0,
			})
		}
	}

	sort.Slice(vv, func(i, j int) bool {
		return vv[i].Weight > vv[j].Weight
	})

	return vv, nil
}
