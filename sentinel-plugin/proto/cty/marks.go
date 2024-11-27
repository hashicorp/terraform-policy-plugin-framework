// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cty

import (
	"encoding/json"
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

func ToMarks(marks [][]byte) cty.ValueMarks {
	if len(marks) == 0 {
		return nil
	}

	ms := make(cty.ValueMarks, len(marks))
	for _, bytes := range marks {
		var mark interface{}
		if err := json.Unmarshal(bytes, &mark); err != nil {
			panic(fmt.Errorf("failed to unmarshal mark: %w", err))
		}
		ms[mark] = struct{}{}
	}
	return ms
}

func FromMarks(marks cty.ValueMarks) [][]byte {
	if len(marks) == 0 {
		return nil
	}

	var ms [][]byte
	for mark := range marks {
		bytes, err := json.Marshal(mark)
		if err != nil {
			panic(fmt.Errorf("failed to marshal mark: %w", err))
		}
		ms = append(ms, bytes)
	}
	return ms
}
