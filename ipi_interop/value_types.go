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
	"unsafe"
)

type PropertyValueType int

// https://github.com/51Degrees/common-cxx/blob/version/4.5/propertyValueType.h#L53
const (
	StringValueType      PropertyValueType = iota // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_STRING
	IntegerValueType                              //FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_INTEGER
	DoubleValueType                               //FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_DOUBLE
	BooleanValueType                              // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_BOOLEAN
	JavascriptValueType                           // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_JAVASCRIPT
	FloatValueType                                // FIFTYONE_DEGREES_PROPERTY_VALUE_SINGLE_PRECISION_FLOAT
	ByteValueType                                 // FIFTYONE_DEGREES_PROPERTY_VALUE_SINGLE_BYTE
	CoordinateValueType                           // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_COORDINATE
	IpAddressValueType                            // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_IP_ADDRESS
	WkbValueType                                  // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_WKB
	ObjectValueType                               // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_OBJECT
	DeclinationValueType                          // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_DECLINATION
	AzimuthValueType                              // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_AZIMUTH
	WkbRValueType                                 // FIFTYONE_DEGREES_PROPERTY_VALUE_TYPE_WKB_R

)

// GetIntegerValue retrieves an integer value from the provided storedBinaryValue using the specified storedValueType.
func (p PropertyValueType) GetIntegerValue(storedBinaryValue *C.StoredBinaryValue, storedValueType C.PropertyValueType) int {
	intVal := C.fiftyoneDegreesStoredBinaryValueToIntOrDefault(
		storedBinaryValue,
		storedValueType,
		0)

	return int(intVal)
}

// GetFloatValue retrieves a float value from the provided storedBinaryValue using the specified storedValueType.
func (p PropertyValueType) GetFloatValue(storedBinaryValue *C.StoredBinaryValue, storedValueType C.PropertyValueType) float64 {
	floatVal := C.fiftyoneDegreesStoredBinaryValueToDoubleOrDefault(
		storedBinaryValue,
		storedValueType,
		0)

	return float64(floatVal)
}

// GetBooleanValue retrieves a boolean value from the provided storedBinaryValue using the specified storedValueType.
func (p PropertyValueType) GetBooleanValue(storedBinaryValue *C.StoredBinaryValue, storedValueType C.PropertyValueType) bool {
	boolVal := C.fiftyoneDegreesStoredBinaryValueToBoolOrDefault(
		storedBinaryValue,
		storedValueType,
		false)

	return bool(boolVal)
}

// GetStringValue retrieves a string value from the provided storedBinaryValue or returns an empty string if nil or invalid.
func (p PropertyValueType) GetStringValue(storedBinaryValue *C.StoredBinaryValue) string {
	if storedBinaryValue == nil {
		return ""
	}
	stringMember := (*C.fiftyoneDegreesString)(unsafe.Pointer(storedBinaryValue))
	if stringMember == nil {
		return ""
	}

	dataPtr := (*C.char)(unsafe.Pointer(&stringMember.value))
	if dataPtr == nil {
		return ""
	}

	return string(C.GoString(dataPtr))
}

// GetIpAddressValue retrieves an IP address value from the provided storedBinaryValue using the specified storedValueType.
func (p PropertyValueType) GetIpAddressValue(storedBinaryValue *C.StoredBinaryValue, storedValueType C.PropertyValueType) (string, error) {
	buf := (*C.char)(C.malloc(C.size_t(defaultSize)))
	defer C.free((unsafe.Pointer)(buf))

	builder := &C.StringBuilder{
		ptr:    buf,
		length: defaultSize,
	}
	C.StringBuilderInit(builder)

	exception := NewException()
	C.StringBuilderAddStringValue(
		builder,
		storedBinaryValue,
		storedValueType,
		16,
		exception.CPtr,
	)
	if !exception.IsOkay() {
		return "", fmt.Errorf(C.GoString(C.ExceptionGetMessage(exception.CPtr)))
	}
	C.StringBuilderComplete(builder)

	return C.GoString(buf), nil
}
