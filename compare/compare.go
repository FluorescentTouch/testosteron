package compare

import (
	"errors"
	"fmt"

	"github.com/json-iterator/go"
)

// Error messages for different comparison failures
const (
	errReasonTypesMismatch  = "different types want: %T, have: %T"
	errReasonNilValue       = "want does not equal nil"
	errReasonValuesNotEqual = "want: %v, have: %v"
	errReasonLengthMismatch = "different len want: %d, have: %d"
	errReasonKeyNotFound    = "key '%s' not found"
	errReasonKeyNotEqual    = "key '%s' does not match: %s"
	errReasonValueNotFound  = "value not found in slice: %v"
)

// JSONCompare compares two JSON byte slices for semantic equality.
// It considers:
// - Maps: all keys must exist and values must match recursively
// - Slices: all elements from 'want' must exist in 'have' (order independent)
// - Primitive types: strict equality comparison
//
// Returns:
//   - eq: true if the JSON structures are semantically equal
//   - err: error describing the first difference found, nil if equal
func JSONCompare(want, have []byte) (eq bool, err error) {
	if want == nil || have == nil {
		err = fmt.Errorf(errReasonNilValue)
		return
	}

	var wantValue interface{}
	var haveValue interface{}

	err = jsoniter.Unmarshal(want, &wantValue)
	if err != nil {
		return
	}

	err = jsoniter.Unmarshal(have, &haveValue)
	if err != nil {
		return
	}

	eq, reason := compareValues(wantValue, haveValue)
	if !eq {
		err = errors.New(reason)
	}
	return
}

// compareValues recursively compares two values of any type.
// It handles:
// - Primitive types: string, float64, bool
// - Complex types: map[string]interface{}, []interface{}
// - nil values
func compareValues(want, have interface{}) (eq bool, reason string) {
	eq = false

	switch wantTyped := want.(type) {
	case string:
		return compareString(wantTyped, have)
	case float64:
		return compareFloat(wantTyped, have)
	case bool:
		return compareBool(wantTyped, have)
	case map[string]interface{}:
		return compareMap(wantTyped, have)
	case []interface{}:
		return compareSlice(wantTyped, have)
	}

	// Handle nil values
	if want == nil && have == nil {
		eq = true
	}

	return
}

// compareString compares a string value with another value.
func compareString(want string, have interface{}) (eq bool, reason string) {
	haveString, ok := have.(string)
	if !ok {
		reason = fmt.Sprintf(errReasonTypesMismatch, want, have)
		return
	}
	eq = want == haveString
	if !eq {
		reason = fmt.Sprintf(errReasonValuesNotEqual, want, haveString)
	}
	return
}

// compareFloat compares a float64 value with another value.
func compareFloat(want float64, have interface{}) (eq bool, reason string) {
	haveFloat, ok := have.(float64)
	if !ok {
		reason = fmt.Sprintf(errReasonTypesMismatch, want, have)
		return
	}
	eq = want == haveFloat
	if !eq {
		reason = fmt.Sprintf(errReasonValuesNotEqual, want, haveFloat)
	}
	return
}

// compareBool compares a bool value with another value.
func compareBool(want bool, have interface{}) (eq bool, reason string) {
	haveBool, ok := have.(bool)
	if !ok {
		reason = fmt.Sprintf(errReasonTypesMismatch, want, have)
		return
	}
	eq = want == haveBool
	if !eq {
		reason = fmt.Sprintf(errReasonValuesNotEqual, want, haveBool)
	}
	return
}

// compareMap compares a map with another value.
// All keys from 'want' must exist in 'have' with matching values.
func compareMap(want map[string]interface{}, have interface{}) (eq bool, reason string) {
	haveMap, ok := have.(map[string]interface{})
	if !ok {
		reason = fmt.Sprintf(errReasonTypesMismatch, want, have)
		return
	}

	for key, wantValue := range want {
		haveValue, ok := haveMap[key]
		if !ok {
			eq = false
			reason = fmt.Sprintf(errReasonKeyNotFound, key)
			return
		}

		eq, reason = compareValues(wantValue, haveValue)
		if !eq {
			reason = fmt.Sprintf(errReasonKeyNotEqual, key, reason)
			return
		}
	}

	eq = true
	return
}

// compareSlice compares a slice with another value.
// All elements from 'want' must exist in 'have' (order independent).
// The slice can have additional elements in 'have'.
func compareSlice(want []interface{}, have interface{}) (eq bool, reason string) {
	haveSlice, ok := have.([]interface{})
	if !ok {
		reason = fmt.Sprintf(errReasonTypesMismatch, want, have)
		return
	}

	if len(want) > len(haveSlice) {
		reason = fmt.Sprintf(errReasonLengthMismatch, len(want), len(haveSlice))
		return
	}

	for _, wantValue := range want {
		eq = findInSlice(wantValue, haveSlice)
		if !eq {
			reason = fmt.Sprintf(errReasonValueNotFound, wantValue)
			return
		}
	}
	eq = true
	return
}

// findInSlice searches for a value in a slice using recursive comparison.
func findInSlice(want interface{}, have []interface{}) (found bool) {
	for _, haveValue := range have {
		found, _ = compareValues(want, haveValue)
		if found {
			return
		}
	}
	return
}
