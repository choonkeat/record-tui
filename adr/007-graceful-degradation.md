# ADR-007: Graceful Degradation - Recording Over Conversion

**Status**: Accepted

**Context**
Recording and HTML conversion are two separate steps. If conversion fails (e.g., template bug), should the entire operation fail?

**Decision**
Recording is the primary goal. If conversion fails, exit with warning but preserve session.log and exit code 0 (success). User can retry conversion manually.

**Consequences**
- ✅ User data never lost (session.log is always kept)
- ✅ Clear feedback (warning message explains what happened)
- ✅ Conversion bugs don't break recording workflow
- ✅ User can manually run session-to-html later
- ❌ Could hide conversion problems (user might not notice warning)
- ❌ HTML file won't exist, but session.log will
