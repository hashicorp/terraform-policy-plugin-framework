// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package policy

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

func NewLogger() hclog.Logger {
	logger := hclog.New(&hclog.LoggerOptions{
		Level: logLevel(),
	})
	return logger
}

func logLevel() hclog.Level {
	level := hclog.LevelFromString(os.Getenv("TF_POLICY_LOG_LEVEL"))
	if level == hclog.NoLevel {
		return hclog.Error
	}
	return level
}
