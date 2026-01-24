// Package js provides the embedded cleaner-core.js for use by other packages.
package js

import (
	_ "embed"
)

// CleanerCoreJS contains the browser-compatible streaming cleaner JavaScript.
// This is the single source of truth used by both:
// - Node.js test harness (via cleaner.js which requires cleaner-core.js)
// - Browser streaming HTML (embedded in template via this variable)
//
//go:embed cleaner-core.js
var CleanerCoreJS string
