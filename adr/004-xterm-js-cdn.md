# ADR-004: Load xterm.js from CDN

**Status**: Accepted

**Context**
Need proper terminal emulation in HTML to display ANSI colors and control sequences correctly. Could bundle as npm dependency or load from CDN.

**Decision**
Load xterm.js@5.3.0 from jsDelivr CDN. Generated HTML is completely standalone except for this single external fetch.

**Consequences**
- ✅ No bundling/build complexity (no webpack, npm)
- ✅ HTML is truly standalone (single file with CDN reference)
- ✅ Browser caching benefits (xterm.js cached across multiple users)
- ✅ Easy to update (just change URL)
- ❌ Requires internet to render HTML initially
- ❌ User IP visible to CDN provider (privacy)
- ❌ Offline viewing not possible without local copy
