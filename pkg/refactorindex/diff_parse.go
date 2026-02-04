package refactorindex

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type DiffFileEntry struct {
	Status  string
	OldPath string
	NewPath string
}

type DiffLine struct {
	Kind    string
	OldLine *int
	NewLine *int
	Text    string
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []DiffLine
}

type FilePatch struct {
	OldPath string
	NewPath string
	Hunks   []DiffHunk
}

func (d DiffFileEntry) PrimaryPath() string {
	if d.NewPath != "" {
		return d.NewPath
	}
	return d.OldPath
}

func ParseNameStatus(data []byte) ([]DiffFileEntry, error) {
	fields := bytes.Split(data, []byte{0})
	entries := make([]DiffFileEntry, 0)
	for i := 0; i < len(fields); {
		status := string(fields[i])
		i++
		if status == "" {
			break
		}
		statusType := status[:1]
		switch statusType {
		case "R", "C":
			if i+1 >= len(fields) {
				return nil, errors.New("invalid name-status output for rename/copy")
			}
			oldPath := string(fields[i])
			newPath := string(fields[i+1])
			i += 2
			entries = append(entries, DiffFileEntry{Status: status, OldPath: normalizeDiffPath(oldPath), NewPath: normalizeDiffPath(newPath)})
		case "A":
			if i >= len(fields) {
				return nil, errors.New("invalid name-status output for add")
			}
			path := string(fields[i])
			i++
			entries = append(entries, DiffFileEntry{Status: status, NewPath: normalizeDiffPath(path)})
		case "D":
			if i >= len(fields) {
				return nil, errors.New("invalid name-status output for delete")
			}
			path := string(fields[i])
			i++
			entries = append(entries, DiffFileEntry{Status: status, OldPath: normalizeDiffPath(path)})
		default:
			if i >= len(fields) {
				return nil, errors.New("invalid name-status output for path")
			}
			path := string(fields[i])
			i++
			normalized := normalizeDiffPath(path)
			entries = append(entries, DiffFileEntry{Status: status, OldPath: normalized, NewPath: normalized})
		}
	}
	return entries, nil
}

func ParseUnifiedDiff(data []byte) ([]FilePatch, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var patches []FilePatch
	var current *FilePatch
	var currentHunk *DiffHunk
	var oldLine int
	var newLine int

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "diff --git ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				oldPath := normalizeDiffPath(parts[2])
				newPath := normalizeDiffPath(parts[3])
				patches = append(patches, FilePatch{OldPath: oldPath, NewPath: newPath})
				current = &patches[len(patches)-1]
				currentHunk = nil
			}
			continue
		}
		if strings.HasPrefix(line, "@@") {
			if current == nil {
				continue
			}
			oldStart, oldLines, newStart, newLines, err := parseHunkHeader(line)
			if err != nil {
				return nil, err
			}
			current.Hunks = append(current.Hunks, DiffHunk{
				OldStart: oldStart,
				OldLines: oldLines,
				NewStart: newStart,
				NewLines: newLines,
			})
			currentHunk = &current.Hunks[len(current.Hunks)-1]
			oldLine = oldStart
			newLine = newStart
			continue
		}
		if currentHunk == nil {
			continue
		}
		if line == "" {
			continue
		}
		switch line[0] {
		case '+':
			if strings.HasPrefix(line, "+++") {
				continue
			}
			text := strings.TrimPrefix(line, "+")
			lineNo := newLine
			newLine++
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{Kind: "+", NewLine: &lineNo, Text: text})
		case '-':
			if strings.HasPrefix(line, "---") {
				continue
			}
			text := strings.TrimPrefix(line, "-")
			lineNo := oldLine
			oldLine++
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{Kind: "-", OldLine: &lineNo, Text: text})
		case ' ':
			text := strings.TrimPrefix(line, " ")
			oldNo := oldLine
			newNo := newLine
			oldLine++
			newLine++
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{Kind: " ", OldLine: &oldNo, NewLine: &newNo, Text: text})
		case '\\':
			continue
		default:
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scan unified diff")
	}
	return patches, nil
}

func parseHunkHeader(line string) (int, int, int, int, error) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "@@") {
		return 0, 0, 0, 0, errors.New("invalid hunk header")
	}
	trimmed = strings.TrimPrefix(trimmed, "@@")
	trimmed = strings.TrimSuffix(trimmed, "@@")
	trimmed = strings.TrimSpace(trimmed)
	parts := strings.Fields(trimmed)
	if len(parts) < 2 {
		return 0, 0, 0, 0, errors.New("invalid hunk header fields")
	}
	oldStart, oldLines, err := parseRange(parts[0], '-')
	if err != nil {
		return 0, 0, 0, 0, err
	}
	newStart, newLines, err := parseRange(parts[1], '+')
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return oldStart, oldLines, newStart, newLines, nil
}

func parseRange(part string, prefix byte) (int, int, error) {
	if len(part) == 0 || part[0] != prefix {
		return 0, 0, errors.New("invalid range header")
	}
	part = part[1:]
	chunks := strings.Split(part, ",")
	start, err := strconv.Atoi(chunks[0])
	if err != nil {
		return 0, 0, errors.Wrap(err, "parse hunk start")
	}
	lines := 1
	if len(chunks) > 1 && chunks[1] != "" {
		lines, err = strconv.Atoi(chunks[1])
		if err != nil {
			return 0, 0, errors.Wrap(err, "parse hunk length")
		}
	}
	return start, lines, nil
}

func normalizeDiffPath(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "/dev/null" {
		return ""
	}
	if strings.HasPrefix(value, "a/") || strings.HasPrefix(value, "b/") {
		return value[2:]
	}
	return value
}

func (d DiffHunk) String() string {
	return fmt.Sprintf("-%d,%d +%d,%d", d.OldStart, d.OldLines, d.NewStart, d.NewLines)
}
