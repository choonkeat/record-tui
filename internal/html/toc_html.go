package html

import (
	"encoding/json"
)

// tocCSS returns the CSS for the floating navigation indicator.
func tocCSS() string {
	return `
    #nav-indicator {
      position: fixed;
      top: 12px;
      right: 12px;
      z-index: 1000;
      background: rgba(30, 30, 30, 0.9);
      border: 1px solid rgba(212, 212, 212, 0.2);
      color: #d4d4d4;
      padding: 6px 14px;
      font-size: 13px;
      font-family: inherit;
      border-radius: 4px;
      user-select: none;
      display: none;
      backdrop-filter: blur(8px);
    }
    #nav-indicator span {
      vertical-align: middle;
    }
    .nav-btn {
      cursor: pointer;
      padding: 2px 6px;
      color: #888;
      transition: color 0.15s;
      font-size: 16px;
    }
    .nav-btn:hover {
      color: #fff;
    }
    .nav-pos {
      color: #666;
      font-size: 12px;
      margin: 0 4px;
    }
    .nav-label {
      color: #a0a0a0;
      font-size: 12px;
      max-width: 200px;
      display: inline-block;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      margin: 0 4px;
    }
    #nav-indicator.expanded {
      padding: 6px 0;
      max-height: 60vh;
      overflow-y: auto;
    }
    #nav-indicator.expanded .nav-btn,
    #nav-indicator.expanded .nav-pos,
    #nav-indicator.expanded .nav-label {
      display: none;
    }
    .nav-compact {
      cursor: pointer;
    }
    .nav-list {
      display: none;
    }
    #nav-indicator.expanded .nav-list {
      display: block;
    }
    .nav-list-item {
      padding: 4px 14px;
      cursor: pointer;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 12px;
      color: #a0a0a0;
      max-width: 300px;
    }
    .nav-list-item:hover {
      background: rgba(255, 255, 255, 0.1);
      color: #fff;
    }
    .nav-list-item.active {
      color: #fff;
      background: rgba(255, 200, 50, 0.1);
      border-left: 2px solid rgba(255, 200, 50, 0.6);
    }
    #nav-highlight {
      position: absolute;
      left: 0;
      right: 0;
      background: rgba(255, 200, 50, 0.12);
      border-left: 3px solid rgba(255, 200, 50, 0.6);
      pointer-events: none;
      z-index: 10;
      transition: top 0.15s ease;
      display: none;
    }
`
}

// tocHTML returns the HTML markup for the navigation indicator.
// Returns empty string if there are no TOC entries.
func tocHTML(entries []TOCEntry) string {
	if len(entries) == 0 {
		return ""
	}
	return `
  <div id="nav-indicator">
    <span class="nav-compact" id="nav-compact">
      <span class="nav-btn" id="nav-prev" title="Previous command (&lt;)">&lt;</span>
      <span class="nav-pos" id="nav-pos"></span>
      <span class="nav-label" id="nav-label"></span>
      <span class="nav-btn" id="nav-next" title="Next command (&gt;)">&gt;</span>
    </span>
    <div class="nav-list" id="nav-list"></div>
  </div>
`
}

// tocJS returns the JavaScript for < > keyboard navigation between user inputs.
// Requires `xterm` variable to be in scope (the Terminal instance).
// Returns empty string if there are no TOC entries.
func tocJS(entries []TOCEntry) string {
	if len(entries) == 0 {
		return ""
	}

	entriesJSON, _ := json.Marshal(entries)

	return `
    // User input < > navigation
    (function() {
      var tocEntries = ` + string(entriesJSON) + `;
      var currentIndex = -1;
      var indicator = document.getElementById('nav-indicator');
      var posEl = document.getElementById('nav-pos');
      var labelEl = document.getElementById('nav-label');
      var highlight = document.createElement('div');
      highlight.id = 'nav-highlight';
      var navList = document.getElementById('nav-list');
      var expanded = false;

      // Resolve actual rendered row for each entry by searching the xterm buffer.
      var resolvedRows = null;

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
          var found = -1;
          var lengths = [30, 20, 10, 5];
          for (var li = 0; li < lengths.length && found < 0; li++) {
            var len = Math.min(lengths[li], label.length);
            if (len < 2) continue;
            found = findInBuffer(buffer, label.substring(0, len), searchFrom);
          }
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
        if (rows.length > 0) {
          indicator.style.display = 'block';
          buildList();
          updateIndicator();
          // Check URL hash on load
          var match = location.hash.match(/^#input-(\d+)$/);
          if (match) {
            navigateTo(parseInt(match[1], 10), false);
          }
        }
      }
      document.addEventListener('xterm-ready', resolveRows);

      function getCellHeight() {
        var terminalDiv = document.getElementById('terminal');
        var xtermScreen = terminalDiv.querySelector('.xterm-screen');
        if (xtermScreen && xterm.buffer.active) {
          var totalRows = xterm.rows;
          if (totalRows > 0) {
            return xtermScreen.getBoundingClientRect().height / totalRows;
          }
        }
        return 17;
      }

      function scrollToRow(row) {
        var terminalDiv = document.getElementById('terminal');
        var termRect = terminalDiv.getBoundingClientRect();
        var termTop = termRect.top + window.pageYOffset;
        var cellHeight = getCellHeight();
        var targetY = termTop + (row * cellHeight);
        window.scrollTo(0, Math.max(0, targetY - 20));
      }

      function highlightRow(row) {
        var terminalDiv = document.getElementById('terminal');
        if (!highlight.parentNode) {
          terminalDiv.style.position = 'relative';
          terminalDiv.appendChild(highlight);
        }
        var cellHeight = getCellHeight();
        highlight.style.top = (row * cellHeight) + 'px';
        highlight.style.height = cellHeight + 'px';
        highlight.style.display = 'block';
      }

      function buildList() {
        navList.innerHTML = '';
        for (var i = 0; i < tocEntries.length; i++) {
          var item = document.createElement('div');
          item.className = 'nav-list-item';
          item.textContent = (i + 1) + '. ' + (tocEntries[i].label || '(empty)');
          item.setAttribute('data-index', i);
          item.addEventListener('click', function(e) {
            e.stopPropagation();
            var idx = parseInt(this.getAttribute('data-index'), 10);
            collapseList();
            navigateTo(idx);
          });
          navList.appendChild(item);
        }
      }

      function updateListActive() {
        var items = navList.querySelectorAll('.nav-list-item');
        for (var i = 0; i < items.length; i++) {
          if (i === currentIndex) {
            items[i].classList.add('active');
          } else {
            items[i].classList.remove('active');
          }
        }
      }

      function toggleExpand() {
        expanded = !expanded;
        if (expanded) {
          indicator.classList.add('expanded');
          updateListActive();
        } else {
          indicator.classList.remove('expanded');
        }
      }

      function collapseList() {
        expanded = false;
        indicator.classList.remove('expanded');
      }

      function updateIndicator() {
        if (!resolvedRows || resolvedRows.length === 0) return;
        if (currentIndex < 0) {
          posEl.textContent = '-/' + resolvedRows.length;
          labelEl.textContent = '';
        } else {
          posEl.textContent = (currentIndex + 1) + '/' + resolvedRows.length;
          labelEl.textContent = tocEntries[currentIndex].label || '';
        }
        if (expanded) updateListActive();
      }

      function navigateTo(index, pushState) {
        if (!resolvedRows) resolveRows();
        if (!resolvedRows || resolvedRows.length === 0) return;
        if (index < 0) index = 0;
        if (index >= resolvedRows.length) index = resolvedRows.length - 1;
        currentIndex = index;
        scrollToRow(resolvedRows[currentIndex]);
        highlightRow(resolvedRows[currentIndex]);
        updateIndicator();
        if (pushState !== false) {
          history.pushState(null, '', '#input-' + currentIndex);
        }
      }

      function goNext() {
        navigateTo(currentIndex + 1);
      }

      function goPrev() {
        navigateTo(currentIndex - 1);
      }

      document.getElementById('nav-prev').addEventListener('click', function(e) { e.stopPropagation(); goPrev(); });
      document.getElementById('nav-next').addEventListener('click', function(e) { e.stopPropagation(); goNext(); });
      document.getElementById('nav-compact').addEventListener('click', toggleExpand);

      document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && expanded) {
          collapseList();
          return;
        }
        if (e.key === '<') {
          e.preventDefault();
          collapseList();
          goPrev();
        } else if (e.key === '>') {
          e.preventDefault();
          collapseList();
          goNext();
        } else if (e.key === 'Tab') {
          e.preventDefault();
          collapseList();
          if (e.shiftKey) {
            goPrev();
          } else {
            goNext();
          }
        }
      });

      // Browser back/forward support
      window.addEventListener('popstate', function() {
        var match = location.hash.match(/^#input-(\d+)$/);
        if (match) {
          navigateTo(parseInt(match[1], 10), false);
        }
      });

      // Track scroll position to update current index
      window.addEventListener('scroll', function() {
        if (!resolvedRows || resolvedRows.length === 0) return;
        var terminalDiv = document.getElementById('terminal');
        var termTop = terminalDiv.getBoundingClientRect().top + window.pageYOffset;
        var cellHeight = getCellHeight();
        var scrollTop = window.pageYOffset + 40;

        var idx = -1;
        for (var i = resolvedRows.length - 1; i >= 0; i--) {
          var entryY = termTop + (resolvedRows[i] * cellHeight);
          if (scrollTop >= entryY) {
            idx = i;
            break;
          }
        }
        if (idx !== currentIndex) {
          currentIndex = idx;
          updateIndicator();
          if (currentIndex >= 0) {
            highlightRow(resolvedRows[currentIndex]);
          } else {
            highlight.style.display = 'none';
          }
        }
      }, { passive: true });
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
