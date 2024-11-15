// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package diagnostics

import (
	sentinel_types "github.com/hashicorp/go-s2/sentinel/types"
	"github.com/hashicorp/hcl/v2"

	proto_types "github.com/hashicorp/go-s2-plugin/sentinel-plugin/proto/types"
)

func FromHclDiagnostics(diagnostics hcl.Diagnostics) []*Diagnostic {
	var diags []*Diagnostic
	for _, diag := range diagnostics {
		diags = append(diags, FromHclDiagnostic(diag))
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

func FromHclDiagnostic(diagnostic *hcl.Diagnostic) *Diagnostic {
	diag := &Diagnostic{
		Severity: FromHclSeverity(diagnostic.Severity),
		Summary:  diagnostic.Summary,
		Detail:   diagnostic.Detail,
	}

	if diagnostic.Extra != nil {
		switch extra := diagnostic.Extra.(type) {
		case sentinel_types.EvaluateResult:
			diag.Result = &Diagnostic_EvaluateResult{
				EvaluateResult: proto_types.FromSentinelEvaluateResult(extra),
			}
		case sentinel_types.FetchResult:
			diag.Result = &Diagnostic_FetchResult{
				FetchResult: proto_types.FromSentinelFetchResult(extra),
			}
		}
	}

	return diag
}

func (diagnostic *Diagnostic) ToHclDiagnostic() *hcl.Diagnostic {
	diag := &hcl.Diagnostic{
		Severity: diagnostic.Severity.ToHclSeverity(),
		Summary:  diagnostic.Summary,
		Detail:   diagnostic.Detail,
	}

	if diagnostic.Result != nil {
		switch result := diagnostic.Result.(type) {
		case *Diagnostic_EvaluateResult:
			diag.Extra = result.EvaluateResult.ToSentinelEvaluateResult()
		case *Diagnostic_FetchResult:
			diag.Extra = result.FetchResult.ToSentinelFetchResult()
		}
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
