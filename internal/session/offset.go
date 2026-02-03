package session

import "strings"

// OffsetMapper maps byte offsets from source content to destination content.
// It tracks which regions of the source were preserved in the destination.
type OffsetMapper struct {
	regions []mappedRegion
	dstLen  int
}

type mappedRegion struct {
	srcStart int
	srcEnd   int // exclusive
	dstStart int
}

// Map converts a source byte offset to a destination byte offset.
// For offsets within preserved regions, returns the exact mapped position.
// For offsets in removed regions, returns the start of the next preserved region.
// For offsets beyond all regions, returns the destination length.
func (m *OffsetMapper) Map(srcOffset int) int {
	if len(m.regions) == 0 {
		return 0
	}

	for _, r := range m.regions {
		if srcOffset < r.srcStart {
			// In a removed region before this preserved region
			return r.dstStart
		}
		if srcOffset < r.srcEnd {
			// Inside this preserved region
			return r.dstStart + (srcOffset - r.srcStart)
		}
	}

	// After all regions
	return m.dstLen
}

// identityMapper returns a mapper that maps every offset to itself.
func identityMapper(contentLen int) *OffsetMapper {
	return &OffsetMapper{
		regions: []mappedRegion{{0, contentLen, 0}},
		dstLen:  contentLen,
	}
}

// NeutralizeAllWithOffsets applies both NeutralizeAltScreenSequences and
// NeutralizeClearSequences to content, returning the processed content
// and a function that maps byte offsets from the original content to
// positions in the processed content.
func NeutralizeAllWithOffsets(content string) (string, func(int) int) {
	// Step 1: Alt screen neutralization with offset tracking
	intermediate, mapper1 := neutralizeAltScreenWithOffsets(content)

	// Step 2: Clear neutralization with offset tracking
	final, mapper2 := neutralizeClearWithOffsets(intermediate)

	// Compose both mappers
	mapFn := func(rawOffset int) int {
		return mapper2.Map(mapper1.Map(rawOffset))
	}

	return final, mapFn
}

// neutralizeAltScreenWithOffsets is like NeutralizeAltScreenSequences but also
// returns an OffsetMapper tracking which source regions were preserved.
func neutralizeAltScreenWithOffsets(content string) (string, *OffsetMapper) {
	altMatches := altScreenPattern.FindAllStringSubmatchIndex(content, -1)
	if len(altMatches) == 0 {
		return content, identityMapper(len(content))
	}

	clearMatches := clearPattern.FindAllStringIndex(content, -1)

	var result strings.Builder
	var regions []mappedRegion
	lastEnd := 0
	inAltScreen := false

	for _, match := range altMatches {
		start, end := match[0], match[1]
		isEnter := content[end-1] == 'h'

		if isEnter && !inAltScreen {
			stripFrom := start
			for _, cm := range clearMatches {
				if cm[0] >= lastEnd && cm[1] <= start {
					stripFrom = cm[0]
					break
				}
			}

			before := content[lastEnd:stripFrom]
			if len(before) > 0 {
				regions = append(regions, mappedRegion{
					srcStart: lastEnd,
					srcEnd:   stripFrom,
					dstStart: result.Len(),
				})
			}
			result.WriteString(before)
			inAltScreen = true
		} else if !isEnter && inAltScreen {
			inAltScreen = false
			beforeContent := result.String()
			remaining := content[end:]
			if strings.TrimSpace(beforeContent) != "" && strings.TrimSpace(remaining) != "" {
				result.WriteString(AltScreenSeparator)
			}
		}

		lastEnd = end
	}

	if !inAltScreen {
		remaining := content[lastEnd:]
		if len(remaining) > 0 {
			regions = append(regions, mappedRegion{
				srcStart: lastEnd,
				srcEnd:   lastEnd + len(remaining),
				dstStart: result.Len(),
			})
		}
		result.WriteString(remaining)
	}

	return result.String(), &OffsetMapper{regions: regions, dstLen: result.Len()}
}

// neutralizeClearWithOffsets is like NeutralizeClearSequences but also
// returns an OffsetMapper tracking which source regions were preserved.
func neutralizeClearWithOffsets(content string) (string, *OffsetMapper) {
	matches := clearPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return content, identityMapper(len(content))
	}

	var result strings.Builder
	var regions []mappedRegion
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		before := content[lastEnd:start]

		if strings.TrimSpace(before) != "" {
			regions = append(regions, mappedRegion{
				srcStart: lastEnd,
				srcEnd:   start,
				dstStart: result.Len(),
			})
			result.WriteString(before)

			remaining := content[end:]
			if strings.TrimSpace(remaining) != "" {
				result.WriteString(ClearSeparator)
			}
		}

		lastEnd = end
	}

	remaining := content[lastEnd:]
	if strings.TrimSpace(remaining) != "" {
		regions = append(regions, mappedRegion{
			srcStart: lastEnd,
			srcEnd:   lastEnd + len(remaining),
			dstStart: result.Len(),
		})
		if result.Len() == 0 {
			result.WriteString(remaining)
		} else {
			result.WriteString(remaining)
		}
	}

	return result.String(), &OffsetMapper{regions: regions, dstLen: result.Len()}
}
