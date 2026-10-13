package utils

import (
	"strings"
)

// SliceContent slices message content into multiple snippets respecting maxLength.
func SliceContent(content string, maxLength int, separator string) []string {
	if maxLength <= 0 || content == "" {
		return []string{}
	}
	if separator == "" {
		separator = "\n"
	}

	var snippets []string
	current := ""

	flushCurrent := func() {
		trimmed := strings.TrimSpace(current)
		if trimmed != "" {
			snippets = append(snippets, trimmed)
		}
		current = ""
	}

	splitIntoChunks := func(str string) {
		runes := []rune(str)
		for i := 0; i < len(runes); i += maxLength {
			end := i + maxLength
			if end > len(runes) {
				end = len(runes)
			}
			part := strings.TrimSpace(string(runes[i:end]))
			if part != "" {
				snippets = append(snippets, part)
			}
		}
	}

	appendLine := func(lineWithSep string) {
		if current == "" || len([]rune(current))+len([]rune(lineWithSep)) <= maxLength {
			current += lineWithSep
			return
		}

		flushCurrent()
		if len([]rune(lineWithSep)) <= maxLength {
			current = lineWithSep
		} else {
			splitIntoChunks(lineWithSep)
		}
	}

	lines := strings.Split(content, separator)
	for i, line := range lines {
		isLast := i == len(lines)-1
		lineWithSep := line
		if !isLast {
			lineWithSep += separator
		}

		if current == "" && len([]rune(lineWithSep)) > maxLength {
			splitIntoChunks(lineWithSep)
			continue
		}

		appendLine(lineWithSep)
	}

	flushCurrent()
	return snippets
}
