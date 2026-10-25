/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package versioner

// ApiVersion defines the current code version in ISO 8601 format; SetConfigParams() creates the application version.
const (
	ApiVersion = "2026_10_25_" // ISO 8601 format string; the application concatenates the application revision number.
)

func ApiVersionString(appRevision string) string {
	return ApiVersion + appRevision
}
