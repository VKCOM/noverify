package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ChangeType describes what happened to the file: it has been deleted, modified, etc.
type ChangeType int

const (
	Deleted ChangeType = iota
	Added
	Changed
)

func (c ChangeType) String() string {
	switch c {
	case Deleted:
		return "Deleted"
	case Added:
		return "Added"
	case Changed:
		return "Changed"
	}
	return "Unknown"
}

// LineRange is range of line numbers that have been changed
type LineRange struct {
	From      int
	To        int
	HaveRange bool
	Range     int
}

// Change describes what happened to the file
type Change struct {
	Type                      ChangeType
	OldName, NewName          string
	OldLineRanges, LineRanges []LineRange
	Valid                     bool
}

// Commit represents git commit :)
type Commit struct {
	Hash    string
	Author  string
	Message string
}

func readShortLine(rd *bufio.Reader) (ln []byte, skip bool, err error) {
	ln, isPrefix, err := rd.ReadLine()
	if err != nil {
		return nil, false, err
	}

	if !isPrefix {
		return ln, false, err
	}

	for {
		ln, isPrefix, err = rd.ReadLine()
		if err != nil {
			return nil, false, err
		}

		if !isPrefix {
			return nil, true, nil
		}
	}
}

var (
	diffOldPrefix      = []byte("--- ")
	diffOldNamePrefix  = []byte("a/")
	diffNewPrefix      = []byte("+++ ")
	diffNewNamePrefix  = []byte("b/")
	patchHeaderPrefix2 = []byte("@@ ")
	patchHeaderSuffix2 = []byte(" @@")
	patchHeaderPrefix3 = []byte("@@@ ")
	patchHeaderSuffix3 = []byte(" @@@")
)

// HasPoint checks whether specified point is contained in range [r.From, r.To].
func (r LineRange) HasPoint(point int) bool {
	return point >= r.From && point <= r.To
}

// LineRangesIntersect checks if provided elem line range intersects and of provided ranges from the list.
func LineRangesIntersect(elem LineRange, list []LineRange) bool {
	for _, r := range list {
		if r.HasPoint(elem.From) || r.HasPoint(elem.To) {
			return true
		}

		if elem.From <= r.From && elem.To >= r.To {
			return true
		}
	}

	return false
}

// Diff computes diff given the refspec (e.g. {"php7_more_fixes", "^php7_testing", "^master"}) and returns
// changed lines in the final version.
// Set workTreeDir to "" if you compute changes only between branches without working copy.
func Diff(gitDir, workTreeDir string, refspec []string) ([]Change, error) {
	args := make([]string, 0, 6+len(refspec))
	args = append(args, "--git-dir="+gitDir, "--no-pager")
	if workTreeDir != "" {
		args = append(args, "--work-tree="+workTreeDir)
	}
	args = append(args, "diff", "-U0")
	args = append(args, refspec...)
	args = append(args, "--")

	cmd := exec.Command("git", args...)
	defer cmd.Wait()

	var out io.Reader
	var err error

	out, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if os.Getenv("LINTER_DEBUG_DIFF") != "" {
		out = io.TeeReader(out, os.Stderr)
	}

	rd := bufio.NewReader(out)

	res, err := parseDiff(rd)
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

// Fetch does git fetch origin from:to
func Fetch(gitDir, from, to string) error {
	args := []string{"--git-dir=" + gitDir, "fetch", "-q", "--no-tags", "origin", from + ":" + to}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

// Log computes log in refspec
func Log(gitDir string, refspec []string) (res []Commit, err error) {
	args := make([]string, 0, 6+len(refspec))
	args = append(args, "--git-dir="+gitDir, "--no-pager", "log", "--oneline", "--format=%H/%an/%s")
	args = append(args, refspec...)

	cmd := exec.Command("git", args...)
	defer cmd.Wait()

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	for _, ln := range lines {
		if ln == "" {
			continue
		}
		parts := strings.SplitN(ln, "/", 3)
		if len(parts) != 3 {
			log.Printf("BAD COMMIT LINE: %s", ln)
			continue
		}
		res = append(res, Commit{Hash: parts[0], Author: parts[1], Message: parts[2]})
	}

	return res, nil
}

func parseDiff(rd *bufio.Reader) ([]Change, error) {
	var res []Change
	var cur Change

	cur.Valid = true

	for {
		ln, skip, err := readShortLine(rd)
		switch {
		case err == io.EOF:
			break
		case err != nil:
			return nil, err
		case skip:
			continue
		}

		switch {
		case bytes.HasPrefix(ln, diffOldPrefix):
			if cur.OldName != "" {
				res = append(res, cur)
				cur.OldName = ""
				cur.NewName = ""
				cur.LineRanges = nil
				cur.OldLineRanges = nil
				cur.Type = 0
			}

			cur.parseOld(ln)
		case bytes.HasPrefix(ln, diffNewPrefix):
			cur.parseNew(ln)
		case bytes.HasPrefix(ln, patchHeaderPrefix2) && bytes.Contains(ln, patchHeaderSuffix2):
			trimmed := bytes.TrimPrefix(ln, patchHeaderPrefix2)
			suffixIdx := bytes.Index(trimmed, patchHeaderSuffix2)
			if suffixIdx < 0 {
				return nil, fmt.Errorf("Could not parse line '%s': no '%s' found", ln, patchHeaderSuffix2)
			}
			if err := cur.parsePatchHeader(trimmed[0:suffixIdx]); err != nil {
				return nil, err
			}
		case bytes.HasPrefix(ln, patchHeaderPrefix3) && bytes.Contains(ln, patchHeaderSuffix3):
			trimmed := bytes.TrimPrefix(ln, patchHeaderPrefix3)
			suffixIdx := bytes.Index(trimmed, patchHeaderSuffix3)
			if suffixIdx < 0 {
				return nil, fmt.Errorf("Could not parse line '%s': no '%s' found", ln, patchHeaderSuffix3)
			}
			if err := cur.parsePatchHeader(trimmed[0:suffixIdx]); err != nil {
				return nil, err
			}
		}
	}

	if cur.OldName != "" {
		res = append(res, cur)
	}

	return res, nil
}

// --- a/oldfile
// --- /dev/null
func (c *Change) parseOld(ln []byte) {
	c.OldName = string(bytes.TrimPrefix(bytes.TrimPrefix(ln, diffOldPrefix), diffOldNamePrefix))
	if c.OldName == "/dev/null" {
		c.Type = Added
	} else {
		c.Type = Changed
	}
}

// +++ b/newfile
// +++ /dev/null
func (c *Change) parseNew(ln []byte) {
	c.NewName = string(bytes.TrimPrefix(bytes.TrimPrefix(ln, diffNewPrefix), diffNewNamePrefix))
	if c.NewName == "/dev/null" {
		c.Type = Deleted
	}
}

// 20433,10
// 284
func (c *Change) parseLineRange(toFileRange []byte) (LineRange, error) {
	commaIdx := bytes.IndexByte(toFileRange, ',')

	var lineNumStr, rangeLenStr string
	var rangeLen int

	if commaIdx > 0 {
		lineNumStr = string(toFileRange[0:commaIdx])
		rangeLenStr = string(toFileRange[commaIdx+1:])
	} else {
		lineNumStr = string(toFileRange)
	}

	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil {
		return LineRange{}, fmt.Errorf("could not parse line number in file range '%s': %s", toFileRange, err.Error())
	}

	if rangeLenStr != "" {
		rangeLen, err = strconv.Atoi(rangeLenStr)
		if err != nil {
			return LineRange{}, fmt.Errorf("could not parse line range in '%s': %s", toFileRange, err.Error())
		}
	}

	// e.g. +1357,2 means two lines, 1357 and 1358, not 1357-1359
	if rangeLen > 0 {
		rangeLen--
	}

	return LineRange{From: lineNum, To: lineNum + rangeLen, HaveRange: rangeLenStr != "", Range: rangeLen}, nil
}

// @@@ <from-file-range> <from-file-range> <to-file-range> @@@ [<class or function>]
// @@@ -20433,288 -21302,345 +20433,10 @@@ class Unrealsync
func (c *Change) parsePatchHeader(ln []byte) error {
	lastIdx := bytes.LastIndexByte(ln, '+')
	if lastIdx < 0 {
		return fmt.Errorf("Could not parse line '%s': no '+' found", ln)
	}

	r, err := c.parseLineRange(ln[lastIdx+1:])
	if err != nil {
		return fmt.Errorf("Could not parse new line range in line '%s': %s", ln, err.Error())
	}

	minusIdx := bytes.IndexByte(ln, '-')
	if minusIdx < 0 {
		return fmt.Errorf("Could not parse line '%s': no '-' found", ln)
	}

	ln = ln[minusIdx+1:]

	spaceIdx := bytes.IndexByte(ln, ' ')
	if spaceIdx < 0 {
		return fmt.Errorf("Could not parse line '%s': no ' ' found", ln)
	}

	oldR, err := c.parseLineRange(ln[0:spaceIdx])
	if err != nil {
		return fmt.Errorf("Could not parse old line range in line '%s': %s", ln, err.Error())
	}

	c.LineRanges = append(c.LineRanges, r)
	c.OldLineRanges = append(c.OldLineRanges, oldR)

	return nil
}
