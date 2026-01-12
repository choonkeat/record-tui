package html

// PlaybackFrame represents a single frame of terminal content at a specific timestamp
type PlaybackFrame struct {
	Timestamp float64 `json:"timestamp"` // Time in seconds (cumulative from start)
	Content   string  `json:"content"`   // Terminal content (with ANSI codes preserved)
}

// FooterLink represents a co-branding link in the footer
type FooterLink struct {
	Text string // Display text (e.g., "swe-swe")
	URL  string // Link URL (e.g., "https://github.com/choonkeat/swe-swe")
}
