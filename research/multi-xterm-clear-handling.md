# Multi-xterm Clear Handling

## Problem

When terminal recordings contain clear sequences (`\x1b[2J`, `\x1b[3J`), the current approach replaces them with a text separator:

```
──────── terminal cleared ────────
```

However, TUI applications (vim, htop, interactive prompts) often use cursor positioning sequences (`\x1b[row;colH`) immediately after clearing, which overwrite the separator text. This makes the separator invisible in TUI-heavy recordings.

## Current Approach

Replace clear sequences with text separator using regex:

```javascript
const clearPattern = /\x1b\[H\x1b\[[23]J|\x1b\[[23]J\x1b\[H|\x1b\[[23]J/g;
text.replace(clearPattern, CLEAR_SEPARATOR);
```

**Pros:**
- Simple implementation
- Single xterm instance
- Low memory usage

**Cons:**
- Separator can be overwritten by cursor positioning
- No guarantee of visibility

## Proposed Approach: Multi-xterm

Create a new xterm instance for each clear sequence, leaving previous terminals untouched.

### Implementation

```javascript
let terminals = [];
let currentTerminal = null;
const container = document.getElementById('terminals');

function createNewTerminal() {
  // Visual separator between terminals
  if (currentTerminal) {
    const separator = document.createElement('div');
    separator.className = 'clear-separator';
    separator.textContent = '──────── terminal cleared ────────';
    container.appendChild(separator);
  }

  const termDiv = document.createElement('div');
  termDiv.className = 'terminal-instance';
  container.appendChild(termDiv);

  currentTerminal = new Terminal({
    cols: 120,
    rows: 50,
    fontSize: 15,
    cursorBlink: false,
    disableStdin: true,
    theme: { background: '#1e1e1e', foreground: '#d4d4d4' },
    allowProposedApi: true,
  });
  currentTerminal.open(termDiv);
  terminals.push(currentTerminal);

  return currentTerminal;
}

function streamWithMultiTerminal(text) {
  // Split on clear sequences
  const parts = text.split(clearPattern);

  for (let i = 0; i < parts.length; i++) {
    if (i > 0) {
      // Clear sequence encountered, create new terminal
      createNewTerminal();
    }
    if (parts[i]) {
      currentTerminal.write(parts[i]);
    }
  }
}
```

### CSS for Separators

```css
.clear-separator {
  padding: 12px 24px;
  text-align: center;
  color: #888888;
  background: linear-gradient(to right,
    transparent,
    rgba(136, 136, 136, 0.3),
    transparent);
  margin: 8px 0;
  font-size: 12px;
}

.terminal-instance {
  margin-bottom: 4px;
}
```

### Memory Cost

- Each xterm.js instance: ~1-5MB depending on buffer size
- Recording with 10 clears: ~10-50MB total
- Recording with 100 clears: ~100-500MB total (problematic)

### Complexity

- Implementation: ~30-50 lines of JS changes
- Streaming integration: Need to track clear boundaries across chunks
- Resize handling: Each terminal needs individual resize-to-fit

### Edge Cases

1. **Many clears**: TUI apps may clear 100+ times. Memory grows linearly.
2. **Rapid clears**: Some apps clear/redraw frequently. Could create many tiny terminals.
3. **Scrollback**: Each terminal has independent scrollback buffer.

## Recommendation

**Don't implement multi-xterm for now.** The complexity and memory cost don't justify the benefit for the current use case.

Instead:
1. Keep the `\r\n` separator change (low cost, helps some cases)
2. Accept that TUI-heavy recordings will have fewer visible separators
3. Consider multi-xterm only if users specifically request it

## Alternative Ideas (Not Recommended)

1. **Screenshot before clear**: Capture canvas image before clear. High complexity.
2. **Accumulate clear count**: Show "terminal cleared (5 times)" at end of TUI section. Loses context.
3. **Detect TUI patterns**: Skip separators during cursor-heavy sections. Fragile heuristics.
