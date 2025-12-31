# ADR-006: Automatically Set FORCE_COLOR and COLORTERM

**Status**: Accepted

**Context**
Terminal applications (cargo, npm, etc.) detect PTY mode and disable colored output to avoid escape codes in pipes. Recording with `script` runs in PTY mode, losing colors.

**Decision**
Automatically set FORCE_COLOR=1 and COLORTERM=truecolor before executing `script` command. Forces color output despite PTY detection.

**Consequences**
- ✅ Colors always captured (when app supports them)
- ✅ Transparent to user (automatic, no action needed)
- ✅ Matches user expectations (running record-tui should capture colors)
- ✅ Standard practice (Node.js tools follow same pattern)
- ❌ Modifies global environment for duration of recording
- ❌ Could theoretically affect other child processes if needed
