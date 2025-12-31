# ADR-001: Language Choice - Go

**Status**: Accepted

**Context**
Originally prototyped in Node.js/TypeScript. Need a language that produces portable, standalone binaries for distribution to ~/bin without runtime dependencies.

**Decision**
Implement record-tui in Go. Single compiled binary handles all recording and HTML conversion tasks.

**Consequences**
- ✅ 2.7MB standalone binary (vs 100MB+ Node.js runtime)
- ✅ No npm/Node.js dependency required
- ✅ Instant startup, direct system calls
- ✅ Easy distribution as single executable
- ❌ Smaller ecosystem for HTML/terminal tools
- ❌ More complex testing (Go testing vs npm test)
