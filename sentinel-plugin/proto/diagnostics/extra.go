// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package diagnostics

import "github.com/hashicorp/go-s2/sentinel/diagnostics"

// SnippetExtra is an extra containing a code snippet. As source information
// is lost when the diagnostic is translated to a protocol buffer, this extra
// captures the relevant parts of the source code.
type SnippetExtra struct {
	diagnostics.Unwrapper

	Snippet *Snippet
}

// ExpressionValuesExtra is an extra containing expression values. As HCL
// evaluation contexts are lost when the diagnostic is translated to a protocol
// buffer, this extra captures the expression values from the context.
type ExpressionValuesExtra struct {
	diagnostics.Unwrapper

	ExpressionValues []*ExpressionValue
}

// FunctionCallExtra is an extra containing a function call. As HCL evaluation
// contexts are lost when the diagnostic is translated to a protocol buffer,
// this extra captures the function call from the context.
type FunctionCallExtra struct {
	diagnostics.Unwrapper

	FunctionCall *FunctionCall
}
