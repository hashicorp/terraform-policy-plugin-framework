// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cty

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

// ToCtyPath converts a Path to a cty.Path.
func (path *Path) ToCtyPath() cty.Path {
	var steps []cty.PathStep
	for _, step := range path.Steps {
		steps = append(steps, step.ToCtyPathStep())
	}
	return steps
}

// FromCtyPath converts a cty.Path to a Path.
func FromCtyPath(path cty.Path) *Path {
	var steps []*Step
	for _, step := range path {
		steps = append(steps, FromCtyPathStep(step))
	}
	return &Path{
		Steps: steps,
	}
}

// ToCtyPathStep converts a Step to a cty.PathStep.
func (step *Step) ToCtyPathStep() cty.PathStep {
	switch step := step.Step.(type) {
	case *Step_AttributeStep:
		return cty.GetAttrStep{
			Name: step.AttributeStep,
		}
	case *Step_IndexStep:
		return cty.IndexStep{
			Key: step.IndexStep.ToCtyValue(cty.DynamicPseudoType),
		}
	default:
		panic(fmt.Errorf("unsupported Step type: %T", step))
	}
}

// FromCtyPathStep converts a cty.PathStep to a Step.
func FromCtyPathStep(step cty.PathStep) *Step {
	switch step := step.(type) {
	case cty.GetAttrStep:
		return &Step{
			Step: &Step_AttributeStep{
				AttributeStep: step.Name,
			},
		}
	case cty.IndexStep:
		return &Step{
			Step: &Step_IndexStep{
				IndexStep: FromCtyValue(step.Key, cty.DynamicPseudoType),
			},
		}
	default:
		panic(fmt.Errorf("unsupported cty.PathStep type: %T", step))
	}
}
