package main

import (
	"bytes"
	"regexp"
	"unicode/utf8"
)

// function UndergoBankruptcy clears all lines in sections deeper than maxLevel
//
// Since the concept of "today" is a rolling todo list which gets forwarded from
// day to day, we can think of "declaring task-bankruptcy" as us discarding
// all tasks from the previous entry.
//
// maxLevel lets us not declare _total_ bankruptcy, but rather just ignore
// content that is too deep in the hierarchy.
//
// NOTE: This is about as naive of a "parser" (if you can even call it that)
// as you can get. It processes the markdown in a line-oriented fashion,
// and doesn't even try to parse the headers correctly. It just counts
// the number of leading `#` characters. TBH this get's the job done.
func UndergoBankruptcy(prevContent []byte, maxLevel int) []byte {
	// level = 0 means no header found yet, as a result, all "top-level"
	// lines (lines before any header) are kept
	level := 0
	lines := make([][]byte, 0)
	headerRegex := regexp.MustCompile(`^(#+)\s`)
	for _, line := range bytes.Split(prevContent, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			// empty line - keep no more than 1 sequential empty line
			if len(lines) > 0 && len(bytes.TrimSpace(lines[len(lines)-1])) == 0 {
				continue
			}
			lines = append(lines, line)
			continue
		}

		if m := headerRegex.FindSubmatch(line); m != nil {
			// header line - update the level
			level = utf8.RuneCount(m[1])
			lines = append(lines, line)
			continue
		}

		if level <= maxLevel {
			// non-header line that's not too deep - keep it
			lines = append(lines, line)
		}
	}

	return bytes.Join(lines, []byte("\n"))
}
