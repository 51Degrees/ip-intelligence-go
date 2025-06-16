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

import (
	"reflect"
	"testing"
)

func TestValues_GetValueWeightByProperty(t *testing.T) {
	tests := []struct {
		name        string
		values      Values
		property    string
		wantValue   interface{}
		wantWeight  float64
		wantSuccess bool
	}{
		{
			name: "existing property with single value",
			values: Values{
				"browser": []*WeightedValue{
					{Value: "Chrome", Weight: 0.9},
				},
			},
			property:    "browser",
			wantValue:   "Chrome",
			wantWeight:  0.9,
			wantSuccess: true,
		},
		{
			name: "existing property with multiple values",
			values: Values{
				"platform": []*WeightedValue{
					{Value: "Windows", Weight: 0.8},
					{Value: "Linux", Weight: 0.2},
				},
			},
			property:    "platform",
			wantValue:   "Windows",
			wantWeight:  0.8,
			wantSuccess: true,
		},
		{
			name: "non-existing property",
			values: Values{
				"os": []*WeightedValue{
					{Value: "Android", Weight: 1.0},
				},
			},
			property:    "browser",
			wantValue:   "",
			wantWeight:  0,
			wantSuccess: false,
		},
		{
			name:        "empty values map",
			values:      Values{},
			property:    "anything",
			wantValue:   "",
			wantWeight:  0,
			wantSuccess: false,
		},
		{
			name: "property with empty value slice",
			values: Values{
				"empty": []*WeightedValue{},
			},
			property:    "empty",
			wantValue:   "",
			wantWeight:  0,
			wantSuccess: false,
		},
		{
			name: "different value types",
			values: Values{
				"string":  []*WeightedValue{{Value: "text", Weight: 0.5}},
				"int":     []*WeightedValue{{Value: 42, Weight: 0.7}},
				"bool":    []*WeightedValue{{Value: true, Weight: 0.9}},
				"float64": []*WeightedValue{{Value: 3.14, Weight: 0.6}},
			},
			property:    "int",
			wantValue:   42,
			wantWeight:  0.7,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotWeight, gotSuccess := tt.values.GetValueWeightByProperty(tt.property)

			if gotSuccess != tt.wantSuccess {
				t.Errorf("GetValueWeightByProperty() success = %v, want %v", gotSuccess, tt.wantSuccess)
			}

			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("GetValueWeightByProperty() value = %v, want %v", gotValue, tt.wantValue)
			}

			if gotWeight != tt.wantWeight {
				t.Errorf("GetValueWeightByProperty() weight = %v, want %v", gotWeight, tt.wantWeight)
			}
		})
	}
}

func TestValues_MultipleProperties(t *testing.T) {
	values := Values{
		"browser": []*WeightedValue{
			{Value: "Chrome", Weight: 0.9},
			{Value: "Firefox", Weight: 0.1},
		},
		"os": []*WeightedValue{
			{Value: "Windows", Weight: 0.7},
			{Value: "MacOS", Weight: 0.2},
			{Value: "Linux", Weight: 0.1},
		},
	}

	tests := []struct {
		name        string
		property    string
		wantValue   interface{}
		wantWeight  float64
		wantSuccess bool
	}{
		{
			name:        "get first browser",
			property:    "browser",
			wantValue:   "Chrome",
			wantWeight:  0.9,
			wantSuccess: true,
		},
		{
			name:        "get first os",
			property:    "os",
			wantValue:   "Windows",
			wantWeight:  0.7,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotWeight, gotSuccess := values.GetValueWeightByProperty(tt.property)

			if gotSuccess != tt.wantSuccess {
				t.Errorf("GetValueWeightByProperty() success = %v, want %v", gotSuccess, tt.wantSuccess)
			}

			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("GetValueWeightByProperty() value = %v, want %v", gotValue, tt.wantValue)
			}

			if gotWeight != tt.wantWeight {
				t.Errorf("GetValueWeightByProperty() weight = %v, want %v", gotWeight, tt.wantWeight)
			}
		})
	}
}

func TestValues_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		values      Values
		property    string
		wantValue   interface{}
		wantWeight  float64
		wantSuccess bool
	}{
		{
			name:        "nil map",
			values:      nil,
			property:    "anything",
			wantValue:   "",
			wantWeight:  0,
			wantSuccess: false,
		},
		{
			name: "empty string property",
			values: Values{
				"": []*WeightedValue{{Value: "value", Weight: 1.0}},
			},
			property:    "",
			wantValue:   "value",
			wantWeight:  1.0,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotWeight, gotSuccess := tt.values.GetValueWeightByProperty(tt.property)

			if gotSuccess != tt.wantSuccess {
				t.Errorf("GetValueWeightByProperty() success = %v, want %v", gotSuccess, tt.wantSuccess)
			}

			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("GetValueWeightByProperty() value = %v, want %v", gotValue, tt.wantValue)
			}

			if gotWeight != tt.wantWeight {
				t.Errorf("GetValueWeightByProperty() weight = %v, want %v", gotWeight, tt.wantWeight)
			}
		})
	}
}
