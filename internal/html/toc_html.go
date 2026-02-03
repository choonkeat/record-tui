package html

import (
	"encoding/json"
	htmlpkg "html"
)

// tocCSS returns the CSS for the floating TOC panel.
func tocCSS() string {
	return `
    #userinput-toggle {
      position: fixed;
      top: 12px;
      right: 12px;
      z-index: 1000;
      background: rgba(30, 30, 30, 0.9);
      border: 1px solid rgba(212, 212, 212, 0.2);
      color: #d4d4d4;
      padding: 6px 12px;
      font-size: 13px;
      font-family: inherit;
      cursor: pointer;
      border-radius: 4px;
      transition: background 0.2s, border-color 0.2s;
    }
    #userinput-toggle:hover {
      background: rgba(50, 50, 50, 0.95);
      border-color: rgba(212, 212, 212, 0.4);
    }

    #userinput-panel {
      position: fixed;
      top: 44px;
      right: 12px;
      z-index: 999;
      background: rgba(30, 30, 30, 0.95);
      border: 1px solid rgba(212, 212, 212, 0.2);
      border-radius: 6px;
      max-height: calc(100vh - 60px);
      max-width: 320px;
      min-width: 180px;
      overflow-y: auto;
      display: none;
      backdrop-filter: blur(8px);
    }
    #userinput-panel.open {
      display: block;
    }

    #userinput-panel ul {
      list-style: none;
      margin: 0;
      padding: 8px 0;
    }
    #userinput-panel li {
      margin: 0;
    }
    #userinput-panel a {
      display: block;
      padding: 6px 16px;
      color: #a0a0a0;
      text-decoration: none;
      font-size: 12px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      transition: background 0.15s, color 0.15s;
      border-left: 2px solid transparent;
    }
    #userinput-panel a:hover {
      background: rgba(255, 255, 255, 0.05);
      color: #e0e0e0;
    }
    #userinput-panel a.active {
      color: #ffffff;
      border-left-color: #569cd6;
      background: rgba(86, 156, 214, 0.1);
    }
`
}

// tocHTML returns the HTML markup for the TOC toggle button and panel.
// Returns empty string if there are no TOC entries.
func tocHTML(entries []TOCEntry) string {
	if len(entries) == 0 {
		return ""
	}

	result := `
  <button id="userinput-toggle" title="User Inputs">User Inputs</button>
  <div id="userinput-panel">
    <ul>
`
	for i, e := range entries {
		escapedLabel := htmlpkg.EscapeString(e.Label)
		result += `      <li><a href="javascript:void(0)" data-toc-index="` + itoa(i) + `" data-toc-line="` + itoa(e.Line) + `" title="` + escapedLabel + `">` + escapedLabel + `</a></li>
`
	}
	result += `    </ul>
  </div>
`
	return result
}

// tocJS returns the JavaScript for TOC panel interactivity.
// Requires `xterm` variable to be in scope (the Terminal instance).
// Returns empty string if there are no TOC entries.
func tocJS(entries []TOCEntry) string {
	if len(entries) == 0 {
		return ""
	}

	entriesJSON, _ := json.Marshal(entries)

	return `
    // User input navigation logic
    (function() {
      var tocEntries = ` + string(entriesJSON) + `;
      var tocToggle = document.getElementById('userinput-toggle');
      var tocPanel = document.getElementById('userinput-panel');
      var tocLinks = tocPanel.querySelectorAll('a[data-toc-line]');

      // Resolve actual rendered row for each entry by searching the xterm buffer.
      // Raw line numbers from Go cannot account for cursor-movement escape sequences
      // that xterm.js processes (overwriting rows), so we search the buffer directly.
      var resolvedRows = null;
      // Search buffer for a needle starting from a given row
      function findInBuffer(buffer, needle, from) {
        for (var row = from; row < buffer.length; row++) {
          var line = buffer.getLine(row);
          if (line) {
            var text = line.translateToString(true);
            if (text.indexOf(needle) !== -1) return row;
          }
        }
        return -1;
      }

      function resolveRows() {
        var rows = [];
        var buffer = xterm.buffer.active;
        if (!buffer || buffer.length === 0) return;
        var searchFrom = 0;
        for (var i = 0; i < tocEntries.length; i++) {
          var label = tocEntries[i].label;
          if (!label || label.length < 2) {
            rows.push(searchFrom);
            continue;
          }
          // Try progressively shorter prefix needles: 30, 20, 10, 5 chars
          var found = -1;
          var lengths = [30, 20, 10, 5];
          for (var li = 0; li < lengths.length && found < 0; li++) {
            var len = Math.min(lengths[li], label.length);
            if (len < 2) continue;
            found = findInBuffer(buffer, label.substring(0, len), searchFrom);
          }
          // Fallback: try substrings starting from various offsets
          // This handles tab-completion where prefix was replaced by shell
          if (found < 0) {
            var offsets = [];
            for (var si = 1; si < label.length; si++) {
              var ch = label[si];
              if (ch === ' ' || ch === "'" || ch === '"' || ch === '/' || ch === '-') {
                offsets.push(si);
                offsets.push(si + 1);
              }
            }
            for (var oi = 0; oi < offsets.length && found < 0; oi++) {
              var off = offsets[oi];
              if (off >= label.length) continue;
              var sub = label.substring(off);
              if (sub.length >= 5) {
                var needle = sub.substring(0, Math.min(20, sub.length));
                found = findInBuffer(buffer, needle, searchFrom);
              }
            }
          }
          if (found >= 0) {
            rows.push(found);
            searchFrom = found + 1;
          } else {
            rows.push(searchFrom);
          }
        }
        resolvedRows = rows;
      }
      // Defer resolution until after xterm rendering and resize complete
      setTimeout(resolveRows, 200);

      // Compute cell height once from the xterm DOM
      function getCellHeight() {
        var terminalDiv = document.getElementById('terminal');
        var xtermScreen = terminalDiv.querySelector('.xterm-screen');
        if (xtermScreen && xterm.buffer.active) {
          // total pixel height / total rows = actual cell height
          var totalRows = xterm.rows;
          if (totalRows > 0) {
            return xtermScreen.getBoundingClientRect().height / totalRows;
          }
        }
        return 17; // fallback
      }

      tocToggle.addEventListener('click', function(e) {
        e.stopPropagation();
        tocPanel.classList.toggle('open');
      });

      // Close panel when clicking outside
      document.addEventListener('click', function(e) {
        if (!tocPanel.contains(e.target) && e.target !== tocToggle) {
          tocPanel.classList.remove('open');
        }
      });

      // Click handler for TOC links
      tocPanel.addEventListener('click', function(e) {
        var link = e.target.closest('a[data-toc-line]');
        if (!link) return;
        e.preventDefault();
        e.stopPropagation();

        // Ensure rows are resolved
        if (!resolvedRows) resolveRows();
        if (!resolvedRows) return;

        var idx = parseInt(link.getAttribute('data-toc-index'), 10);
        if (idx >= 0 && idx < resolvedRows.length) {
          tocPanel.classList.remove('open');
          // Use requestAnimationFrame to scroll after panel closes
          requestAnimationFrame(function() {
            scrollToRow(resolvedRows[idx]);
          });
        }
      });

      function scrollToRow(row) {
        var terminalDiv = document.getElementById('terminal');
        var termRect = terminalDiv.getBoundingClientRect();
        var termTop = termRect.top + window.pageYOffset;
        var cellHeight = getCellHeight();
        var targetY = termTop + (row * cellHeight);
        window.scrollTo(0, Math.max(0, targetY - 20));
      }

      // Highlight current section based on scroll position
      function updateActiveLink() {
        if (!resolvedRows) return;
        var terminalDiv = document.getElementById('terminal');
        var termTop = terminalDiv.getBoundingClientRect().top + window.pageYOffset;
        var cellHeight = getCellHeight();
        var scrollTop = window.pageYOffset + 40;

        var activeIndex = 0;
        for (var i = resolvedRows.length - 1; i >= 0; i--) {
          var entryY = termTop + (resolvedRows[i] * cellHeight);
          if (scrollTop >= entryY) {
            activeIndex = i;
            break;
          }
        }

        tocLinks.forEach(function(link, idx) {
          link.classList.toggle('active', idx === activeIndex);
        });
      }

      window.addEventListener('scroll', updateActiveLink, { passive: true });
      // Initial highlight
      setTimeout(updateActiveLink, 300);
    })();
`
}

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
