/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package versioner

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiVersionString(t *testing.T) {
	require.Equal(t, ApiVersion+"42", ApiVersionString("42"))
}
