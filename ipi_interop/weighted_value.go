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

// maxRawWeighting is the maximum raw weighting a value can carry; the raw
// weightings for a property sum to this value. It matches C.UINT16_MAX.
const maxRawWeighting = 65535.0

// WeightedValue represents a value together with its weighting. RawWeight is the
// raw 16-bit weighting from the data file (0..65535) and Weight is the derived
// probability RawWeight/65535 in (0, 1]. Properties whose values are not
// weighted carry a zero weighting.
type WeightedValue struct {
	Value     interface{}
	Weight    float64
	RawWeight uint16
}

// Values is a map where each key is a string representing a property, and the value is a slice of WeightedValue pointers.
type Values map[string][]*WeightedValue

// GetValueByProperty retrieves the first value for the specified property (ignoring weight).
// This is the recommended method for all properties except MCC.
// Returns the value and a boolean indicating success or failure.
func (v Values) GetValueByProperty(property string) (interface{}, bool) {
	if val, ok := v[property]; ok && len(val) > 0 {
		return val[0].Value, true
	}
	return "", false
}

// GetValueWeightByProperty retrieves the first value and its weight for the specified property.
// Returns the value, weight, and a boolean indicating success or failure.
func (v Values) GetValueWeightByProperty(property string) (interface{}, float64, bool) {
	if val, ok := v[property]; ok && len(val) > 0 {
		return val[0].Value, val[0].Weight, true
	}
	return "", 0, false
}

// Append adds a value without weight (Weight will be set to 0.0).
// Use this for all properties except MCC.
func (v Values) Append(property string, value interface{}) {
	v[property] = append(v[property], &WeightedValue{
		Value:  value,
		Weight: 0.0,
	})
}

// AppendWithWeight adds a value with a specific normalized weight.
func (v Values) AppendWithWeight(property string, value interface{}, weight float64) {
	v[property] = append(v[property], &WeightedValue{
		Value:  value,
		Weight: weight,
	})
}

// AppendWeighted adds a value with its raw weighting, deriving the normalized
// Weight. Use this for properties whose values carry a weighting, such as the
// weighted country-code lists and MCC.
func (v Values) AppendWeighted(property string, value interface{}, rawWeight uint16) {
	v[property] = append(v[property], &WeightedValue{
		Value:     value,
		Weight:    float64(rawWeight) / maxRawWeighting,
		RawWeight: rawWeight,
	})
}

func (v Values) InitProperty(property string) {
	if _, ok := v[property]; !ok {
		v[property] = make([]*WeightedValue, 0, 0)
	}
}
