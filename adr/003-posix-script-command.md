# ADR-003: Use POSIX `script` Command for Recording

**Status**: Accepted

**Context**
Multiple approaches exist for capturing terminal sessions: custom PTY wrapper (like asciinema), system-specific recording, or built-in `script` command.

**Decision**
Use POSIX `script` command via exec.Command() to record sessions. Captures terminal output with ANSI codes preserved.

**Consequences**
- ✅ Zero custom code for PTY handling (OS does it)
- ✅ Built into macOS/Linux, battle-tested since 1979
- ✅ Trivial integration (single exec call)
- ✅ Captures full ANSI color codes
- ❌ Limited to POSIX systems (not native Windows)
- ❌ Cannot customize capture behavior
- ❌ Depends on `script` command version/behavior
