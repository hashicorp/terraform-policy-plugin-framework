// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package types

import "github.com/hashicorp/terraform-policy-core/policy/types"

func FromPolicyEvaluateResult(result types.EvaluateResult) EvaluateResult {
	switch result {
	case types.EvaluateResultUnknown:
		return EvaluateResult_UNKNOWN_EVALUATE_RESULT
	case types.EvaluateResultError:
		return EvaluateResult_ERROR_EVALUATE_RESULT
	case types.EvaluateResultAllow:
		return EvaluateResult_ALLOW_EVALUATE_RESULT
	case types.EvaluateResultDeny:
		return EvaluateResult_DENY_EVALUATE_RESULT
	default:
		// for backwards and forwards compatibility, we set any unknown result
		// to invalid.
		return EvaluateResult_INVALID_EVALUATE_RESULT
	}
}

func (result EvaluateResult) ToPolicyEvaluateResult() types.EvaluateResult {
	switch result {
	case EvaluateResult_UNKNOWN_EVALUATE_RESULT:
		return types.EvaluateResultUnknown
	case EvaluateResult_ERROR_EVALUATE_RESULT:
		return types.EvaluateResultError
	case EvaluateResult_ALLOW_EVALUATE_RESULT:
		return types.EvaluateResultAllow
	case EvaluateResult_DENY_EVALUATE_RESULT:
		return types.EvaluateResultDeny
	default:
		// for backwards and forwards compatibility, we set any unknown results
		// to an error.
		return types.EvaluateResultError
	}
}

func FromPolicyFetchResult(result types.FetchResult) FetchResult {
	switch result {
	case types.FetchResultInvalid:
		return FetchResult_ERROR_FETCH_RESULT
	case types.FetchResultValid:
		return FetchResult_VALID_FETCH_RESULT
	default:
		// for backwards and forwards compatibility, we set any unknown result
		// to invalid.
		return FetchResult_INVALID_FETCH_RESULT
	}
}

func (result FetchResult) ToPolicyFetchResult() types.FetchResult {
	switch result {
	case FetchResult_ERROR_FETCH_RESULT:
		return types.FetchResultInvalid
	case FetchResult_VALID_FETCH_RESULT:
		return types.FetchResultValid
	default:
		// for backwards and forwards compatibility, we set any unknown results
		// to invalid.
		return types.FetchResultInvalid
	}
}
