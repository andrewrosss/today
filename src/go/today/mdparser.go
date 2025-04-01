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
	// the lines we're going to output
	lines := make([][]byte, 0)
	// regex to match the header lines and capture the header level
	headerRegex := regexp.MustCompile(`^(#+)\s`)
	// level = 0 means no header found yet, as a result, all "top-level"
	// lines (lines before any header) are kept
	level := 0
	// inside code fences we don't want to do any header processing,
	// e.g. comments in python start with `#`, but we don't want to
	// treat them as headers. So let's keep track of the current code
	// fence delimiter and skip header processing if we're inside a code fence.
	var currentCodeFenceDelim []byte = nil
	codeFenceRegex := regexp.MustCompile("^(\\s*(`{3,}|~{3,})\\s*)")

	for _, line := range bytes.Split(prevContent, []byte("\n")) {
		trimmed := bytes.TrimSpace(line)

		if len(trimmed) == 0 {
			// empty line - keep no more than 1 sequential empty line
			if len(lines) > 0 && len(bytes.TrimSpace(lines[len(lines)-1])) == 0 {
				continue
			}
			lines = append(lines, line)
			continue
		}

		if currentCodeFenceDelim != nil {
			// we're in a code fence - check if we're at the end
			if bytes.Equal(currentCodeFenceDelim, trimmed) {
				// end of code fence - reset the delimiter
				currentCodeFenceDelim = nil
			}
		} else {
			// we're not in a code fence - check if we're at the start of one
			// or a header
			if m := codeFenceRegex.FindSubmatch(line); m != nil {
				// start of code fence AND we're not already in one - set the
				// current code fence delimiter
				currentCodeFenceDelim = bytes.TrimSpace(m[1])
			} else if m := headerRegex.FindSubmatch(line); m != nil {
				// header line - update the level
				level = utf8.RuneCount(m[1])

				// since we're keeping all headers (and only dropping the
				// _content_) in sections deeper than maxLevel, we append
				// the header line to the output here instead of falling
				// through to the next case
				lines = append(lines, line)
				continue
			}
		}

		if level <= maxLevel {
			// non-header line that's not too deep - keep it
			lines = append(lines, line)
		}
	}

	return bytes.Join(lines, []byte("\n"))
}
