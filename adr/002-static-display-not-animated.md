# ADR-002: Static Display Instead of Animated Playback

**Status**: Accepted

**Context**
Could display session as animated frames showing command execution in real-time (with timing.log), or as single static snapshot of final output.

**Decision**
Use single PlaybackFrame with final session output. Renders instantly in browser without animation overhead.

**Consequences**
- ✅ Simpler implementation (fewer moving parts)
- ✅ Smaller HTML files (one frame vs many)
- ✅ Faster generation
- ✅ Works in all browsers with no JavaScript complexity
- ❌ Cannot see command execution flow in real-time
- ❌ Loses timing information from original session
- ⚠️ Future: Can extend to support animated playback by generating multiple frames from timing.log
