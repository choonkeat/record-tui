package html

// PlaybackFrame represents a single frame of terminal content at a specific timestamp
type PlaybackFrame struct {
	Timestamp float64 `json:"timestamp"` // Time in seconds (cumulative from start)
	Content   string  `json:"content"`   // Terminal content (with ANSI codes preserved)
}
