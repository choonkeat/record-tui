package html

import (
	"encoding/json"
	htmlpkg "html"
)

// tocCSS returns the CSS for the floating TOC panel.
func tocCSS() string {
	return `
    #toc-toggle {
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
    #toc-toggle:hover {
      background: rgba(50, 50, 50, 0.95);
      border-color: rgba(212, 212, 212, 0.4);
    }

    #toc-panel {
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
    #toc-panel.open {
      display: block;
    }

    #toc-panel ul {
      list-style: none;
      margin: 0;
      padding: 8px 0;
    }
    #toc-panel li {
      margin: 0;
    }
    #toc-panel a {
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
    #toc-panel a:hover {
      background: rgba(255, 255, 255, 0.05);
      color: #e0e0e0;
    }
    #toc-panel a.active {
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
  <button id="toc-toggle" title="Table of Contents">TOC</button>
  <div id="toc-panel">
    <ul>
`
	for i, e := range entries {
		escapedLabel := htmlpkg.EscapeString(e.Label)
		result += `      <li><a href="#" data-toc-index="` + itoa(i) + `" data-toc-line="` + itoa(e.Line) + `" title="` + escapedLabel + `">` + escapedLabel + `</a></li>
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
    // TOC panel logic
    (function() {
      var tocEntries = ` + string(entriesJSON) + `;
      var tocToggle = document.getElementById('toc-toggle');
      var tocPanel = document.getElementById('toc-panel');
      var tocLinks = tocPanel.querySelectorAll('a[data-toc-line]');

      tocToggle.addEventListener('click', function() {
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

        var line = parseInt(link.getAttribute('data-toc-line'), 10);
        scrollToLine(line);
        tocPanel.classList.remove('open');
      });

      function scrollToLine(line) {
        var terminalDiv = document.getElementById('terminal');
        var termTop = terminalDiv.getBoundingClientRect().top + window.scrollY;
        // Get xterm.js cell dimensions from the actual rendered element
        var cellHeight = terminalDiv.querySelector('.xterm-rows') ?
          terminalDiv.querySelector('.xterm-rows').children[0].getBoundingClientRect().height : 17;
        var targetY = termTop + (line * cellHeight);
        window.scrollTo({ top: Math.max(0, targetY - 20), behavior: 'smooth' });
      }

      // Highlight current section based on scroll position
      function updateActiveLink() {
        var terminalDiv = document.getElementById('terminal');
        var termTop = terminalDiv.getBoundingClientRect().top + window.scrollY;
        var cellHeight = terminalDiv.querySelector('.xterm-rows') ?
          terminalDiv.querySelector('.xterm-rows').children[0].getBoundingClientRect().height : 17;
        var scrollTop = window.scrollY + 40;

        var activeIndex = 0;
        for (var i = tocEntries.length - 1; i >= 0; i--) {
          var entryY = termTop + (tocEntries[i].line * cellHeight);
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
      setTimeout(updateActiveLink, 200);
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
