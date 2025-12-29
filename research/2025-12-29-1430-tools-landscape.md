# Tools & Libraries Landscape for convo-as-html

Research date: 2025-12-29

## Executive Summary

This document surveys the landscape of tools and libraries for converting Claude conversations to styled HTML with iMessage-like appearance (user messages right-aligned in blue, AI responses left-aligned). The research covers ANSI-to-HTML converters, markdown renderers, code syntax highlighters, and existing chat exporters.

---

## 1. ANSI Escape Code to HTML Converters

ANSI escape codes are commonly used in Claude's terminal output for colors and formatting. Converting these to HTML is essential for preserving styled terminal output.

### JavaScript/Node.js Solutions

#### **ansi-to-html** (JavaScript)
- **Repository**: [https://github.com/rburns/ansi-to-html](https://github.com/rburns/ansi-to-html)
- **NPM**: [https://www.npmjs.com/package/ansi-to-html](https://www.npmjs.com/package/ansi-to-html)
- **Features**:
  - Originally ported from bcat's ANSI to HTML converter
  - Configurable output: inline styles or CSS classes
  - Handles ANSI codes for foreground/background colors
  - Drops "erase in line" escape code (\x1b[K)
  - Supports custom color overrides and palette modification
  - Stream mode for saving state across invocations
- **Usage**: `new Convert().toHtml(ansiText)`
- **Status**: Actively maintained, well-established

#### **ansi_up** (JavaScript)
- **Repository**: [https://github.com/drudru/ansi_up](https://github.com/drudru/ansi_up)
- **NPM**: [https://www.npmjs.com/package/ansi_up](https://www.npmjs.com/package/ansi_up)
- **Features**:
  - Zero dependencies
  - Works in browser and Node.js (isomorphic)
  - Single ES6 JavaScript file
  - Transforms ANSI color codes to colorful HTML
  - Easy to use API
- **Advantage**: No dependencies makes it lightweight
- **Status**: Well-maintained

#### **ansi-html** (JavaScript)
- **Repository**: [https://github.com/Tjatse/ansi-html](https://github.com/Tjatse/ansi-html)
- **NPM**: [https://www.npmjs.com/package/ansi-html](https://www.npmjs.com/package/ansi-html)
- **Features**:
  - Elegant library for converting chalked (ANSI) text to HTML
  - Support for all chalk styles and colors
- **Status**: Community-maintained

#### **stream-ansi2html** (JavaScript)
- **Repository**: [https://github.com/timmarinin/ansi2html](https://github.com/timmarinin/ansi2html)
- **NPM**: [https://www.npmjs.com/package/ansi2html](https://www.npmjs.com/package/ansi2html)
- **Features**:
  - Node.js stream-based converter
  - Convertible via piping with other streams

### Go Solutions

#### **ansihtml** (Go)
- **Package**: [https://pkg.go.dev/github.com/robert-nix/ansihtml](https://pkg.go.dev/github.com/robert-nix/ansihtml)
- **GitHub**: github.com/robert-nix/ansihtml
- **Features**:
  - Parses ANSI escape sequences to HTML
  - Outputs text suitable for `<pre>` tags
  - `ConvertToHTML()` - inline styles
  - `ConvertToHTMLWithClasses()` - CSS class approach
  - Encodes text effects as `<span>` tags
- **Use Case**: Server-side Go applications

#### **terminal-to-html** (Go)
- **Repository**: [https://github.com/buildkite/terminal-to-html](https://github.com/buildkite/terminal-to-html)
- **Package**: [https://pkg.go.dev/github.com/buildkite/terminal-to-html/v3](https://pkg.go.dev/github.com/buildkite/terminal-to-html/v3)
- **Features**:
  - Converts arbitrary shell output with ANSI to beautifully rendered HTML
  - Includes CLI tool and web server support
  - Support for iTerm2 inline images
  - More feature-rich than simple ANSI converters
- **Use Case**: Terminal output visualization

### Python Solutions

#### **ansi2html** (Python)
- **Repository**: [https://github.com/pycontribs/ansi2html](https://github.com/pycontribs/ansi2html)
- **PyPI**: [https://pypi.org/project/ansi2html/](https://pypi.org/project/ansi2html/)
- **Features**:
  - Converts text with ANSI color codes to HTML or LaTeX
  - Widely used and established
  - Can output complete HTML documents or fragments
- **Usage**: `Ansi2HTMLConverter().convert(text)`

#### **ansiconv** (Python)
- **Documentation**: [https://pythonhosted.org/ansiconv/](https://pythonhosted.org/ansiconv/)
- **Features**:
  - Converts ANSI coded text to plain text or HTML

#### **ansiparser** (Python)
- **PyPI**: [https://pypi.org/project/ansiparser/](https://pypi.org/project/ansiparser/)
- **Features**:
  - Convenient library for converting ANSI escape sequences
  - Implements state-machine parser like a terminal
  - Outputs formatted text or HTML

---

## 2. Markdown to HTML Renderers

Claude conversations typically include markdown-formatted content. These libraries convert markdown to structured HTML while preserving formatting.

### **marked** (JavaScript)
- **Repository**: [https://github.com/markedjs/marked](https://github.com/markedjs/marked)
- **Website**: [https://marked.js.org/](https://marked.js.org/)
- **NPM**: [https://www.npmjs.com/package/marked](https://www.npmjs.com/package/marked)
- **Features**:
  - Lightweight and fast markdown parser
  - GitHub Flavored Markdown (GFM) support
  - Supports code fencing with language specifiers
  - Configurable tokens and extensions
  - Built for speed
- **Advantage**: Fast, minimal overhead
- **Status**: Actively maintained

### **markdown-it** (JavaScript)
- **Repository**: [https://github.com/markdown-it/markdown-it](https://github.com/markdown-it/markdown-it)
- **Website**: [https://markdown-it.github.io/](https://markdown-it.github.io/)
- **NPM**: [https://www.npmjs.com/package/markdown-it](https://www.npmjs.com/package/markdown-it)
- **Features**:
  - Follows CommonMark specification
  - Extensible plugin system
  - Support for tables, strikethrough, task lists
  - Plugin architecture for custom rendering
- **Advantage**: Highly configurable with plugins
- **Status**: Actively maintained

### **Showdown** (JavaScript)
- **Repository**: [https://github.com/showdownjs/showdown](https://github.com/showdownjs/showdown)
- **Website**: [https://showdownjs.com/](https://showdownjs.com/)
- **NPM**: [https://www.npmjs.com/package/showdown](https://www.npmjs.com/package/showdown)
- **Features**:
  - Bidirectional markdown to HTML and HTML to markdown conversion
  - Works client-side (browser) or server-side (Node.js)
  - GitHub flavored markdown support
- **Use Case**: Projects needing bidirectional conversion

### **remark + rehype** (JavaScript - Unified Ecosystem)
- **Remark**: [https://github.com/remarkjs/remark](https://github.com/remarkjs/remark)
  - **Website**: [https://remark.js.org/](https://remark.js.org/)
  - **NPM**: [https://www.npmjs.com/package/remark](https://www.npmjs.com/package/remark)
- **Rehype**: [https://github.com/rehypejs/rehype](https://github.com/rehypejs/rehype)
  - **Website**: [https://rehype.js.org/](https://rehype.js.org/)
- **Bridge**: [https://github.com/remarkjs/remark-rehype](https://github.com/remarkjs/remark-rehype)
- **Features**:
  - Part of the unified ecosystem for AST-based content transformation
  - Remark = markdown AST (mdast)
  - Rehype = HTML AST (hast)
  - Extensive plugin ecosystem
  - Allows manipulation of AST for custom transformations
- **Advantage**: Decoupled markdown and HTML processing, extensive plugin ecosystem
- **Use Case**: Complex transformation pipelines, custom styling requirements
- **Status**: Actively maintained, industry standard

---

## 3. Code Syntax Highlighting

Claude output often includes code blocks. These libraries provide syntax highlighting for code in HTML.

### **highlight.js** (JavaScript)
- **Website**: [https://highlightjs.org/](https://highlightjs.org/)
- **Repository**: [https://github.com/highlightjs/highlight.js](https://github.com/highlightjs/highlight.js)
- **NPM**: [https://www.npmjs.com/package/highlight.js](https://www.npmjs.com/package/highlight.js)
- **Features**:
  - Supports 192 languages
  - 512+ color themes
  - Automatic language detection
  - Zero dependencies
  - Works with any markup
  - Browser, Node.js, Deno support
  - Works with vanilla JS, Vue, React, etc.
- **Usage**: `hljs.highlightAll()` or `hljs.highlight(code, {language: 'javascript'})`
- **Advantage**: Extensive language support, zero dependencies
- **Status**: Industry standard, actively maintained

### **Prism.js** (JavaScript)
- **Website**: [https://prismjs.com/](https://prismjs.com/)
- **Repository**: [https://github.com/PrismJS/prism](https://github.com/PrismJS/prism)
- **NPM**: [https://www.npmjs.com/package/prismjs](https://www.npmjs.com/package/prismjs)
- **Features**:
  - Lightweight and elegant
  - Supports multiple themes and plugins
  - Line numbers, copy-to-clipboard, and more plugins
  - Language defined via `class="language-xxxx"`
  - Extensible without modifying code
  - Plugin architecture
- **Advantage**: Lightweight, excellent plugin system
- **Status**: Widely used, actively maintained

### **highlight.js vs Prism.js**
- **highlight.js**: Better for auto-detection, more languages
- **Prism.js**: Lighter weight, more elegant plugins

---

## 4. Git Diff Syntax Highlighting

Claude conversations may include git diffs. These tools render diffs with syntax highlighting.

### **diff2html** (JavaScript)
- **Website**: [https://diff2html.xyz/](https://diff2html.xyz/)
- **Repository**: [https://github.com/rtfpessoa/diff2html](https://github.com/rtfpessoa/diff2html)
- **NPM**: [https://www.npmjs.com/package/diff2html](https://www.npmjs.com/package/diff2html)
- **Features**:
  - Converts unified diff or git diff to syntax-highlighted HTML
  - GitHub-style diff rendering
  - Uses highlight.js for syntax highlighting of code within diffs
  - Diff2HtmlUI wrapper for easy DOM injection
  - Side-by-side or line-by-line view options
  - Can be used as library or CLI tool
- **Advantage**: Complete solution for diff visualization
- **Status**: Actively maintained

### **delta** (Rust)
- **Repository**: [https://github.com/dandavison/delta](https://github.com/dandavison/delta)
- **Features**:
  - Syntax-aware diff pager
  - More themes and configuration options
  - Terminal and HTML output
- **Use Case**: Terminal-first approach

---

## 5. Existing Chat/Conversation Exporters

These projects export chat conversations to HTML, providing reference implementations for styling and structure.

### **Claude Export** (Browser Extension)
- **Chrome Web Store**: [https://chromewebstore.google.com/detail/claude-export-tool/gogjdkpnnlnckhenijamkacidklkhgkk](https://chromewebstore.google.com/detail/claude-export-tool/gogjdkpnnlnckhenijamkacidklkhgkk)
- **Website**: [https://www.claudexporter.com/](https://www.claudexporter.com/)
- **Features**:
  - Captures JSON page content during download
  - Converts to HTML mimicking Claude's browser interface
  - Generates index.html with table of contents
  - Multiple export formats: PDF, Markdown, TXT, JSON, CSV, HTML
  - Browser extension for Chrome/Firefox
- **Advantage**: Direct Claude interface styling reference

### **ClaudeExport** (GitHub)
- **Repository**: [https://github.com/Llaves/ClaudeExport](https://github.com/Llaves/ClaudeExport)
- **Features**:
  - Open-source Claude conversation exporter
  - HTML output
- **Status**: Community project

### **claude-export** (Browser Script)
- **Repository**: [https://github.com/ryanschiang/claude-export](https://github.com/ryanschiang/claude-export)
- **Features**:
  - Browser script to export conversations
  - Supports: Markdown, JSON, PNG (image)
  - Runs entirely locally from browser console
  - No external dependencies
- **Advantage**: Privacy-focused, no backend required

### **claude-chat-exporter** (JavaScript)
- **Repository**: [https://github.com/agarwalvishal/claude-chat-exporter](https://github.com/agarwalvishal/claude-chat-exporter)
- **Features**:
  - JavaScript tool for exporting Claude conversations
  - Outputs well-formatted Markdown
  - Browser-based

### **claude-conversation-extractor** (Python)
- **Repository**: [https://github.com/ZeroSumQuant/claude-conversation-extractor](https://github.com/ZeroSumQuant/claude-conversation-extractor)
- **PyPI**: [https://pypi.org/project/claude-conversation-extractor/](https://pypi.org/project/claude-conversation-extractor/)
- **Features**:
  - Extracts conversation logs from Claude Code
  - Supports HTML output with modern styling
  - Python-based tool
- **Status**: Community project

### **iMessage-Export** (macOS Messages)
- **Repository**: [https://github.com/aaronpk/iMessage-Export](https://github.com/aaronpk/iMessage-Export)
- **Features**:
  - Exports iMessage history to HTML, CSV, or SQL
  - iMessage style reference for conversation styling
  - Includes photos in exports
- **Use Case**: Reference for right-aligned message styling

### **imessage-exporter** (Rust)
- **Repository**: [https://github.com/ReagentX/imessage-exporter](https://github.com/ReagentX/imessage-exporter)
- **Crates**: [https://crates.io/crates/imessage-exporter](https://crates.io/crates/imessage-exporter)
- **Features**:
  - Export iMessage data to TXT/HTML
  - Mimics Messages app 1:1
  - Cross-platform
- **Status**: Actively maintained

### **OSX-Messages-Exporter** (macOS Messages)
- **Repository**: [https://github.com/cfinke/OSX-Messages-Exporter](https://github.com/cfinke/OSX-Messages-Exporter)
- **Features**:
  - Exports iMessages and SMS to HTML
  - Custom CSS styling support
  - Message threading

---

## 6. HTML/CSS Patterns for Chat UI

Styling for conversation interfaces with right-aligned messages.

### **Flexbox Approach**
- **Resources**:
  - [CSS-Tricks - Styling Comment Threads](https://css-tricks.com/styling-comment-threads/)
  - [MDN - Flexbox](https://developer.mozilla.org/en-US/docs/Learn_web_development/Core/CSS/layout/Flexbox)
  - [CSS-Tricks - A Complete Guide to Flexbox](https://css-tricks.com/snippets/css/a-guide-to-flexbox/)
- **Pattern**:
  - Use classes: `.chat-message-left` (AI) and `.chat-message-right` (User)
  - Use flexbox with `justify-content: flex-start` or `flex-end`
  - Color differently (blue for user, gray for AI)
  - Optional: Message bubbles with border-radius

### **Bulma CSS Framework**
- **Website**: [https://bulma.io/](https://bulma.io/)
- **Message Component**: [https://bulma.io/documentation/components/message/](https://bulma.io/documentation/components/message/)
- **Features**:
  - Pre-built message component
  - Built on Flexbox
  - message-header and message-body parts
  - Customizable with color modifiers

### **CodePen Examples**
- [Flexbox Chat](https://codepen.io/thatdevguy/pen/ANYebj)
- [iOS CSS Chat Message Bubbles](https://codepen.io/swards/pen/gxQmbj)
- [Simple Chat Conversation](https://codepen.io/MariamMassadeh/pen/nNaBEz)

---

## 7. Recommended Tech Stack for convo-as-html

Based on the research, here's a recommended approach:

### **Markdown Rendering**
- **Primary**: `marked` (fast, lightweight) or `remark + rehype` (extensible)
- **Advantage**: `marked` for simplicity, `remark+rehype` for complex transformations

### **ANSI to HTML**
- **JavaScript**: `ansi-to-html` (configurable) or `ansi_up` (zero deps)
- **Go**: `terminal-to-html` (feature-rich) or `ansihtml` (simple)
- **Python**: `ansi2html` (established)

### **Code Syntax Highlighting**
- **Primary**: `highlight.js` (extensive language support)
- **Alternative**: `Prism.js` (lighter weight, elegant)

### **Git Diff Highlighting**
- **Primary**: `diff2html` (complete solution)

### **Chat UI Styling**
- **CSS**: Custom flexbox-based layout (no framework needed per CLAUDE.md preference)
- **Scoping**: Wrap all styles in `.convo-as-html` class to avoid conflicts

### **Browser/Node.js Stack** (Preferred approach)
- **Language**: JavaScript/Node.js (per CLAUDE.md: "use Typescript with node dependencies in package.json")
- **Core Libraries**:
  - `marked` - markdown parsing
  - `ansi-to-html` - ANSI color codes
  - `highlight.js` - code syntax highlighting
  - `diff2html` - git diff rendering
- **Build**: Makefile with targets for fmt, test, lint, build
- **Output**: Static HTML files with embedded CSS (no framework dependencies)

### **Alternative: Go Stack** (More minimal)
- **Language**: Go (per CLAUDE.md: "prefer stdlib")
- **Libraries**:
  - Use standard library for HTML generation
  - `terminal-to-html` for ANSI conversion
  - Embed syntax highlighting or call highlight.js via template
  - Minimal external dependencies

---

## 8. Key Design Considerations

1. **Custom CSS Scoping**: Wrap all styles in `.convo-as-html` class to prevent conflicts with other libraries
2. **Message Structure**:
   - User messages: right-aligned, blue background
   - AI messages: left-aligned, light gray background
   - Use semantic HTML: `<article>`, `<section>`, or `<div>` containers
3. **Code Blocks**:
   - Preserve language detection for syntax highlighting
   - Consider code copy button (highlight.js plugins)
4. **Mobile Responsiveness**: Flexbox makes responsive design natural
5. **Dark Mode Support**: Consider CSS custom properties for theming
6. **Performance**: Choose lightweight libraries (ansi_up, marked, highlight.js)
7. **Accessibility**: Semantic HTML, proper contrast ratios

---

## 9. IMPORTANT FINDING: ANSI Color Code Sources

**Claude Code JSONL Format Does NOT Contain ANSI Codes**

After analyzing the sample conversation file (`560eea06-2609-4466-a53c-5934d1e0486a.jsonl`), the stored JSONL contains **plain text** without ANSI escape sequences. However, as shown in the terminal screenshots, the original live output DOES have rich ANSI coloring including:
- Cyan/turquoise (for command output, file paths)
- Red (for errors)
- Yellow (for warnings)
- Green (for success messages)
- Gray/dim (for metadata)

### Challenge: Getting ANSI Codes into HTML

**Option 1: Capture Raw Terminal Output (Recommended)**
- Pipe Claude Code output to a file that captures ANSI codes: `command | tee output.txt`
- Or use `script` command to record terminal session with colors intact
- Then parse that file with `ansi-to-html` for conversion

**Option 2: Reconstruct Colors from Context**
- Analyze tool_result content to infer colors based on type:
  - Error messages → red
  - Success messages → green
  - File paths/code → cyan
  - Warnings → yellow
- Limited but works for predictable output patterns

**Option 3: Use Claude Code Export Extensions**
- The existing Claude exporters (Chrome extension, claude-export) might preserve colors
- Worth testing their export format

### ANSI Code Converter Status

Given that **Claude Code JSONL strips ANSI codes**, the key is:
1. Capture raw output with ANSI codes (separate from JSONL export)
2. Use `ansi-to-html` or `ansi_up` to convert captured terminal output
3. Combine with JSONL metadata (user/AI distinction, timestamps)

### Recommendation for convo-as-html

**Two-file approach:**
1. **JSONL file**: Structured conversation data from Claude Code
2. **Raw output file**: Captured terminal output with ANSI codes (captured separately)

Then the converter can:
- Use JSONL for message structure (who said what, timestamps)
- Use raw output for colors (ANSI → HTML)
- Combine both for final styled HTML

---

## Sources

### ANSI to HTML
- [ansi-to-html npm](https://www.npmjs.com/package/ansi-to-html)
- [ansi-to-html GitHub](https://github.com/rburns/ansi-to-html)
- [ansi_up GitHub](https://github.com/drudru/ansi_up)
- [ansi-html GitHub](https://github.com/Tjatse/ansi-html)
- [ansihtml Go Package](https://pkg.go.dev/github.com/robert-nix/ansihtml)
- [terminal-to-html GitHub](https://github.com/buildkite/terminal-to-html)
- [ansi2html Python](https://github.com/pycontribs/ansi2html)

### Markdown to HTML
- [marked GitHub](https://github.com/markedjs/marked)
- [markdown-it GitHub](https://github.com/markdown-it/markdown-it)
- [Showdown GitHub](https://github.com/showdownjs/showdown)
- [remark GitHub](https://github.com/remarkjs/remark)
- [rehype GitHub](https://github.com/rehypejs/rehype)
- [remark-rehype GitHub](https://github.com/remarkjs/remark-rehype)

### Code Syntax Highlighting
- [highlight.js](https://highlightjs.org/)
- [highlight.js GitHub](https://github.com/highlightjs/highlight.js)
- [Prism.js](https://prismjs.com/)
- [Prism GitHub](https://github.com/PrismJS/prism)

### Git Diff
- [diff2html](https://diff2html.xyz/)
- [diff2html GitHub](https://github.com/rtfpessoa/diff2html)

### Chat Exporters
- [Claude Export Tool](https://www.claudexporter.com/)
- [ClaudeExport GitHub](https://github.com/Llaves/ClaudeExport)
- [claude-export GitHub](https://github.com/ryanschiang/claude-export)
- [claude-conversation-extractor](https://pypi.org/project/claude-conversation-extractor/)
- [iMessage-Export GitHub](https://github.com/aaronpk/iMessage-Export)
- [imessage-exporter](https://crates.io/crates/imessage-exporter)
- [OSX-Messages-Exporter GitHub](https://github.com/cfinke/OSX-Messages-Exporter)

### HTML/CSS Patterns
- [CSS-Tricks - Styling Comment Threads](https://css-tricks.com/styling-comment-threads/)
- [MDN - Flexbox](https://developer.mozilla.org/en-US/docs/Learn_web_development/Core/CSS/layout/Flexbox)
- [Bulma CSS](https://bulma.io/)
