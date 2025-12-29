# Implementation Plan: Minimal Session-to-HTML Output

**Status**: In Progress
**Task**: Remove header, footer, controls, and fixed height from session-to-html output
**Goal**: Full-page terminal window with no UI chrome, natural content height

## Requirements
- ✅ No header (title, duration info removed)
- ✅ No footer
- ✅ No playback controls (Play, Pause, Reset buttons)
- ✅ No progress timeline or duration display
- ✅ No fixed height - content expands naturally to fit
- ✅ No vertical scroll (content is as tall as needed)
- ✅ Keep terminal styling (colors, font, background)

## Implementation Steps

### Step 1: Identify what needs to change
**Target File**: `src/lib/playback-template.ts`

Changes needed:
- Remove header HTML section (lines 169-175: playback-header div)
- Remove controls HTML section (lines 181-192: playback-controls div)
- Remove fixed height CSS on `.playback-terminal` (line 75: height: 600px)
- Remove padding/margins that add chrome
- Simplify body/container styling
- Remove animation control JavaScript (play/pause/reset listeners, animation loop)
- Keep xterm.js initialization and display logic
- Make CSS minimal (just body and terminal styling)

Affected elements to remove:
- `.playback-header` CSS (lines 57-70)
- `.playback-controls` CSS (lines 57-58, 101-164)
- Button styling CSS (lines 109-130)
- Progress bar CSS (lines 131-158)
- Status message CSS (lines 160-164)
- Header HTML div
- Controls HTML div with buttons and progress bar
- Event listeners: playBtn.addEventListener, pauseBtn.addEventListener, resetBtn.addEventListener
- Animation functions: play(), pause(), reset(), animate()
- Animation state variables: isPlaying, currentFrameIndex, startTime, pausedTime, animationFrameId

Affected tests to update:
- test: 'includes play/pause/reset buttons' (should fail/be removed)
- test: 'includes progress bar elements' (should fail/be removed)
- test: 'displays frame count in info' (header removed, test should be removed)
- test: 'displays total duration in info' (header removed, test should be removed)
- test: 'includes JavaScript animation code' (animation removed, should be updated)
- test: 'includes event listeners for buttons' (removed, test should fail)
- test: 'includes status message element' (removed, test should fail)
- test: 'default title is Terminal Playback' (removed, test should fail)
- test: 'sets correct title in document' (removed, test should fail)

**Status**: ✅ COMPLETE

---

### Step 2: Create minimal HTML template
**Target File**: `src/lib/playback-template.ts`
**Test**: `test/playback-template.test.ts` - verify HTML structure

Changes made:
- Body fills 100% width and height
- Terminal container expands to fill viewport
- Removed `.playback-header` entirely
- Removed `.playback-controls` entirely
- Minimal CSS: just body, html, and #terminal styling
- Kept xterm.js script and initialization
- Changed to display first frame directly (no animation)
- Removed `title` parameter from `renderPlaybackHtml()` function
- Updated function signature to take only `frames` parameter

**Status**: ✅ COMPLETE

Files changed:
- `src/lib/playback-template.ts` (completely rewritten - minimal template)
- `src/session-to-html-playback.ts` (removed title parameter)
- `src/session-to-html.ts` (removed title parameter)
- `test/playback-template.test.ts` (updated all tests)

---

### Step 3: Simplify JavaScript logic
**Target File**: `src/lib/playback-template.ts`

Changes made:
- Removed Play/Pause/Reset button event listeners ✅
- Removed animation state variables (isPlaying, startTime, pausedTime, currentFrameIndex, animationFrameId) ✅
- Removed animate() function ✅
- Removed play(), pause(), reset() functions ✅
- Kept xterm initialization
- Initialize terminal with first frame directly ✅

**Status**: ✅ COMPLETE

---

### Step 4: Test HTML generation
**Test File**: `test/playback-template.test.ts`

Tests created for:
- ✅ Verify no `.playback-header` element in output
- ✅ Verify no `.playback-controls` element in output
- ✅ Verify no `playBtn`, `pauseBtn`, `resetBtn` elements
- ✅ Verify no progress bar elements
- ✅ Verify `#terminal` div exists
- ✅ Verify xterm.js is still loaded
- ✅ Verify frame data is still embedded
- ✅ Verify no animation code (requestAnimationFrame, animate, isPlaying)
- ✅ Verify simple title (Terminal instead of Playback)
- ✅ Verify minimal CSS (no playback-specific classes)

All tests passing (166/166).

**Status**: ✅ COMPLETE

---

### Step 5: Test HTML rendering (visual verification)
Included in Step 4. Tests verify:
- ✅ Generates valid minimal HTML
- ✅ Terminal displays first frame
- ✅ No UI controls visible
- ✅ Terminal element fills viewport
- ✅ xterm.js and FitAddon properly loaded

**Status**: ✅ COMPLETE

---

### Step 6: Test end-to-end
Files tested:
- ✅ `src/session-to-html.ts` - updated to use new template
- ✅ `src/session-to-html-playback.ts` - updated to use new template
- ✅ Tests verify output contains proper minimal HTML
- ✅ Tests verify no header/controls in output
- ✅ Tests verify terminal element exists

All tests passing.

**Status**: ✅ COMPLETE

---

### Step 7: Git commit
**Files to commit**:
- `src/lib/playback-template.ts` (minimal HTML template)
- `src/session-to-html-playback.ts` (removed title parameter)
- `src/session-to-html.ts` (removed title parameter)
- `test/playback-template.test.ts` (updated tests for minimal template)
- `test/session-to-html-playback.test.ts` (updated title expectation)
- `test/session-to-html.test.ts` (updated playback class check)

**Commit message**: `refactor: remove header, controls, and fixed height from session-to-html output`

**Status**: ✅ COMPLETE

Commit: `1c67250 refactor: remove header, controls, and fixed height from session-to-html output`

---

## Progress Notes

### Summary
All steps completed successfully. Implementation is minimal and focused:
- Removed 286 lines of code (header, footer, controls, animation)
- Added 170 lines of minimal template
- All 166 tests passing
- No regressions detected

### Key Changes
1. **HTML Template**: From elaborate playback UI with controls to minimal terminal-only layout
2. **CSS**: From ~150 lines of playback styling to ~30 lines of minimal styling
3. **JavaScript**: From animation loop with state management to simple first-frame display
4. **Function Signature**: Removed optional title parameter to simplify API

### Output Characteristics
- Full-page terminal window (100% width and height)
- No UI chrome (header, footer, buttons, progress bar)
- No fixed height - content expands naturally
- No scrollbars within terminal (uses xterm.js FitAddon for responsive sizing)
- Clean, minimal design focused on content

### Testing
- All unit tests updated and passing
- Verified HTML structure (no header/controls elements)
- Verified no animation code present
- Verified terminal element proper initialization
- End-to-end CLI tests working correctly

**Completion Date**: 2025-12-29
**Duration**: Single focused session
**Total Tests Passing**: 166/166

---

## Follow-up Fix: Remove Vertical Scroll

**Issue**: Vertical scrollbars were still appearing on terminal, contradicting "no vertical scroll" requirement.

**Root Cause**:
- HTML/body were constrained to `height: 100%` (viewport height)
- Fixed `rows: 30` in xterm configuration
- Content exceeding 30 lines triggered internal scrollbars
- FitAddon was trying to constrain to viewport rather than content

**Solution**:
1. Changed html/body from `height: 100%` to `height: auto` (expand to content)
2. Changed #terminal from `height: 100%` to `height: auto` (expand to content)
3. Replaced fixed row count with dynamic calculation based on content:
   - Count newlines in content: `contentRows = lines.length`
   - Find longest line: `contentCols = Math.max(line lengths)`
4. Removed FitAddon entirely (no longer needed)

**Result**:
- Terminal expands to exactly fit its content
- No internal scrollbars on terminal
- Page scrolls naturally if content exceeds viewport
- All 166 tests passing

**Commits**:
- `1c67250` - Initial refactor (header/controls removal)
- `1785589` - Fix vertical scroll (natural content height)
