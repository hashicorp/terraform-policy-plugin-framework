// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package values

import "github.com/zclconf/go-cty/cty"

const (
	Sensitive = "sensitive"
)

// SensitiveValue marks a value as sensitive.
func SensitiveValue(value cty.Value) cty.Value {
	return value.Mark(Sensitive)
}

// UnsensitiveValue returns the unsensitive value and a boolean indicating if
// the value was sensitive. This function returns any other marks that the
// value may have.
func UnsensitiveValue(value cty.Value) (cty.Value, bool) {
	v, marks := value.Unmark()
	sensitive := false

	for mark := range marks {
		if mark == Sensitive {
			sensitive = true
		}
	}
	delete(marks, Sensitive)

	return v.WithMarks(marks), sensitive
}
