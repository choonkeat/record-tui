# ADR-005: Line-Based Detection for Session Metadata Stripping

**Status**: Accepted

**Context**
`script` command adds header (Script started on, Command:) and footer (Script done on, Saving session) lines that need removal before displaying content.

**Decision**
Parse session.log line-by-line, removing lines matching known patterns: "Script started on", "Command:", "Script done on", "Saving session".

**Consequences**
- ✅ Simple, readable code (easy to understand and debug)
- ✅ Single pass through content (O(n) efficient)
- ✅ Easy to extend with new patterns as needed
- ✅ Works with real-world session.log files from ~/.record-tui/
- ❌ Not formally specified (depends on `script` output format)
- ❌ Fragile to future `script` command changes
- ⚠️ Mitigation: Test suite includes real session.log files
