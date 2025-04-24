// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package convert

import (
	"fmt"
	"strings"
)

type Path []*Step

type Step struct {
	Key     string
	Indices []string
}

func (p Path) String() string {
	var path []string
	for _, step := range p {
		if len(step.Indices) == 0 {
			path = append(path, step.Key)
		} else {
			path = append(path, fmt.Sprintf("%s[%s]", step.Key, strings.Join(step.Indices, "][")))
		}
	}
	return strings.Join(path, ".")
}

func (p Path) Append(key string) Path {
	return append(p, &Step{Key: key})
}

func (p Path) WithIndex(index string) Path {
	if len(p) == 0 {
		return Path{
			&Step{
				Indices: []string{index},
			},
		}
	}

	last := p[len(p)-1]
	last.Indices = append(last.Indices, index)
	return p
}

var (
	_ error = (*PathError)(nil)
)

type PathError struct {
	Err  error
	Path Path
}

func (p *PathError) Error() string {
	return fmt.Sprintf("error at %s: %v", p.Path, p.Err)
}

func withPath(path Path, err error) error {
	if err == nil {
		return nil
	}
	return &PathError{
		Err:  err,
		Path: path,
	}
}
