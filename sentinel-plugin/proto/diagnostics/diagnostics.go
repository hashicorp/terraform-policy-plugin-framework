// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package diagnostics

import (
	"strings"

	"github.com/hashicorp/go-s2/sentinel/diagnostics"
	"github.com/hashicorp/go-s2/sentinel/diagnostics/snippet"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	proto_cty "github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/cty"
	proto_types "github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/types"
)

func FromHclDiagnostics(diagnostics hcl.Diagnostics, sources map[string]*hcl.File) []*Diagnostic {
	var diags []*Diagnostic
	for _, diag := range diagnostics {
		diags = append(diags, FromHclDiagnostic(diag, sources))
	}
	return diags
}

func ToHclDiagnostics(diagnostics []*Diagnostic) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for _, diag := range diagnostics {
		diags = append(diags, diag.ToHclDiagnostic())
	}
	return diags
}

func FromHclDiagnostic(diagnostic *hcl.Diagnostic, sources map[string]*hcl.File) *Diagnostic {
	diag := &Diagnostic{
		Severity: FromHclSeverity(diagnostic.Severity),
		Summary:  diagnostic.Summary,
		Detail:   diagnostic.Detail,
	}

	if diagnostic.Subject != nil {
		diag.Subject = FromHclRange(*diagnostic.Subject)
	}
	if diagnostic.Context != nil {
		diag.Context = FromHclRange(*diagnostic.Context)
	}

	result, ok := hcl.DiagnosticExtra[*diagnostics.Result](diagnostic)
	if ok {
		switch {
		case result.EvaluateResult != nil:
			diag.Result = &Diagnostic_EvaluateResult{
				EvaluateResult: proto_types.FromSentinelEvaluateResult(*result.EvaluateResult),
			}
		case result.FetchResult != nil:
			diag.Result = &Diagnostic_FetchResult{
				FetchResult: proto_types.FromSentinelFetchResult(*result.FetchResult),
			}
		}
	}

	attr, ok := hcl.DiagnosticExtra[*diagnostics.Attribute](diagnostic)
	if ok {
		diag.Attribute = proto_cty.FromCtyPath(attr.Path)
	}

	if extra, ok := hcl.DiagnosticExtra[*SnippetExtra](diagnostic); ok {
		diag.Snippet = extra.Snippet
	} else {
		diag.Snippet = FromSentinelSnippet(snippet.Snippet(diagnostic, sources))
	}

	if extra, ok := hcl.DiagnosticExtra[*ExpressionValuesExtra](diagnostic); ok {
		diag.ExpressionValues = extra.ExpressionValues
	} else {
		diag.ExpressionValues = FromSentinelValues(snippet.BuildExpressionValues(diagnostic))
	}

	if extra, ok := hcl.DiagnosticExtra[*FunctionCallExtra](diagnostic); ok {
		diag.FunctionCall = extra.FunctionCall
	} else {
		fn, ok := hcl.DiagnosticExtra[hclsyntax.FunctionCallDiagExtra](diagnostic)
		if ok && diagnostic.EvalContext != nil {
			absoluteName := fn.CalledFunctionName()
			if fn, ok := diagnostic.EvalContext.Functions[absoluteName]; ok {
				baseName := absoluteName
				if idx := strings.LastIndex(baseName, "::"); idx >= 0 {
					baseName = baseName[idx+2:]
				}

				diag.FunctionCall = &FunctionCall{
					AbsoluteName: absoluteName,
					Function:     proto_cty.FromCtyFunction(baseName, fn),
				}
			}
		}
	}

	return diag
}

func (diagnostic *Diagnostic) ToHclDiagnostic() *hcl.Diagnostic {
	builder := diagnostics.ForSeverity(diagnostic.Severity.ToHclSeverity(), diagnostic.Summary).
		WithDetail(diagnostic.Detail)

	if diagnostic.Subject != nil {
		builder = builder.WithSubject(diagnostic.Subject.ToHclRange())
	}

	if diagnostic.Context != nil {
		builder = builder.WithContext(diagnostic.Context.ToHclRange())
	}

	if diagnostic.Result != nil {
		switch result := diagnostic.Result.(type) {
		case *Diagnostic_EvaluateResult:
			builder = builder.WithEvaluateResult(result.EvaluateResult.ToSentinelEvaluateResult())
		case *Diagnostic_FetchResult:
			builder = builder.WithFetchResult(result.FetchResult.ToSentinelFetchResult())
		}
	}

	if diagnostic.Attribute != nil {
		builder = builder.WithAttribute(diagnostic.Attribute.ToCtyPath())
	}

	diag := builder.Build()

	if diagnostic.Snippet != nil {
		diag = diagnostics.AddExtra(diag, &SnippetExtra{
			Snippet: diagnostic.Snippet,
		})
	}

	if len(diagnostic.ExpressionValues) > 0 {
		diag = diagnostics.AddExtra(diag, &ExpressionValuesExtra{
			ExpressionValues: diagnostic.ExpressionValues,
		})
	}

	if diagnostic.FunctionCall != nil {
		diag = diagnostics.AddExtra(diag, &FunctionCallExtra{
			FunctionCall: diagnostic.FunctionCall,
		})
	}

	return diag
}

func FromHclSeverity(severity hcl.DiagnosticSeverity) Severity {
	switch severity {
	case hcl.DiagError:
		return Severity_ERROR
	case hcl.DiagWarning:
		return Severity_WARNING
	default:
		return Severity_INVALID
	}
}

func (severity Severity) ToHclSeverity() hcl.DiagnosticSeverity {
	switch severity {
	case Severity_ERROR:
		return hcl.DiagError
	case Severity_WARNING:
		return hcl.DiagWarning
	default:
		return hcl.DiagInvalid
	}
}

func (rng *Range) ToHclRange() hcl.Range {
	return hcl.Range{
		Filename: rng.Filename,
		Start:    rng.Start.ToHclPos(),
		End:      rng.End.ToHclPos(),
	}
}

func FromHclRange(rng hcl.Range) *Range {
	return &Range{
		Filename: rng.Filename,
		Start:    FromHclPos(rng.Start),
		End:      FromHclPos(rng.End),
	}
}

func (pos *Position) ToHclPos() hcl.Pos {
	return hcl.Pos{
		Byte:   int(pos.Byte),
		Line:   int(pos.Line),
		Column: int(pos.Column),
	}
}

func FromHclPos(pos hcl.Pos) *Position {
	return &Position{
		Byte:   int64(pos.Byte),
		Line:   int64(pos.Line),
		Column: int64(pos.Column),
	}
}

func (s *Snippet) ToSentinelSnippet() *snippet.DiagnosticSnippet {
	if s == nil {
		return nil
	}

	snippet := &snippet.DiagnosticSnippet{
		Code:                 s.Code,
		StartLine:            int(s.StartLine),
		HighlightStartOffset: int(s.HighlightStartOffset),
		HighlightEndOffset:   int(s.HighlightEndOffset),
	}
	if s.Context != nil {
		snippet.Context = &s.Context.Context
	}
	return snippet
}

func FromSentinelSnippet(s *snippet.DiagnosticSnippet) *Snippet {
	if s == nil {
		return nil
	}

	snippet := &Snippet{
		Code:                 s.Code,
		StartLine:            int64(s.StartLine),
		HighlightStartOffset: int64(s.HighlightStartOffset),
		HighlightEndOffset:   int64(s.HighlightEndOffset),
	}
	if s.Context != nil {
		snippet.Context = &Snippet_Context{Context: *s.Context}
	}
	return snippet
}

func FromSentinelValues(evs []snippet.ExpressionValue) []*ExpressionValue {
	var values []*ExpressionValue
	for _, ev := range evs {
		values = append(values, FromSentinelValue(ev))
	}
	return values
}

func FromSentinelValue(ev snippet.ExpressionValue) *ExpressionValue {
	return &ExpressionValue{
		Path:  proto_cty.FromCtyPath(ev.Path),
		Value: proto_cty.FromCtyValue(ev.Value, cty.DynamicPseudoType),
	}
}
