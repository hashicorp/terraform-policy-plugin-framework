// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugins

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
)

func NewLogger(plugin string) hclog.Logger {
	logger := hclog.New(&hclog.LoggerOptions{
		Level: logLevel(plugin),
	})
	return logger
}

func logLevel(plugin string) hclog.Level {
	level := hclog.LevelFromString(os.Getenv(fmt.Sprintf("TF_POLICY_LOG_LEVEL_%s", plugin)))
	if level == hclog.NoLevel {
		return hclog.Error
	}
	return level
}
