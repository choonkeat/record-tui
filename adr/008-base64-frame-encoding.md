# ADR-008: Base64 Encode Frame Data in HTML

**Status**: Accepted

**Context**
Session content with ANSI codes needs to be embedded in HTML and decoded in JavaScript. Could use direct string, JSON, or binary-safe encoding.

**Decision**
Encode frames as base64 JSON in the HTML template. JavaScript decodes with atob()/Uint8Array before parsing.

**Consequences**
- ✅ Binary-safe (handles all characters without escaping)
- ✅ No quote/backslash escaping issues
- ✅ ANSI codes preserved exactly (no accidental interpretation)
- ✅ Simple JavaScript decoding (standard atob function)
- ❌ 33% size overhead (base64 encoding expansion)
- ❌ Extra decode step on page load (negligible performance impact)
