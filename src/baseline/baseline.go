package baseline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"sort"
	"strconv"
	"strings"
)

// profileVersion helps to figure out which profile version is used in a suppression file.
// If JSON representation is changed, newer version could probably fail at the decoding
// phase, but if we only change the interpretation of the fields, it can be
// useful to have a version field to reject older versions to avoid weird behavior.
//
// Versions log:
// 1 - initial version.
// 2 - added Profile.LinterVersion field.
// 3 - added Profile.CreatedAt field.
const profileVersion = 3

// Profile is a project-wide suppression profile (baseline file).
type Profile struct {
	LinterVersion string

	// CreatedAt is a Unix time that represents the moment at which
	// this profile was generated.
	CreatedAt int64

	Files map[string]FileProfile
}

// FileProfile contains all reports suppressed for the associated file.
type FileProfile struct {
	Filename string

	// Reports maps a warning hash to a suppression report info.
	Reports map[uint64]Report
}

// Stats is an extra information that is saved to the suppression profile.
// It's not used by the linter itself, but it's useful for the other tools.
type Stats struct {
	// CountTotal is a total number of reports stored in the suppression profile.
	CountTotal int

	CountPerCheck map[string]int
}

// Count returns the number of file reports that have the same hash.
//
// How to interpret the return value:
// n=0 - warning is not suppressed
// n=1 - warning is suppressed
// n>1 - warning is suppressed and there are hash collisions
//
// Hash collisions are handled by reports counting.
func (f *FileProfile) Count(hash uint64) int {
	return f.Reports[hash].Count
}

// Report is a suppressed warning information.
type Report struct {
	// Count is a number of reports that had the same hash.
	Count int

	Hash uint64
}

// ReadProfile reads a suppression profile from a provided reader.
func ReadProfile(r io.Reader) (*Profile, *Stats, error) {
	var p jsonProfile
	dec := json.NewDecoder(r)
	if err := dec.Decode(&p); err != nil {
		return nil, nil, fmt.Errorf("can't decode baseline file: %v (version mismatch?)", err)
	}

	if p.Version != profileVersion {
		return nil, nil, fmt.Errorf("version mismatch: want %d, have %d", p.Version, profileVersion)
	}

	files := make(map[string]FileProfile, len(p.Files))
	for _, f := range p.Files {
		reports := make(map[uint64]Report)
		for _, hashList := range f.HashLists {
			for _, hash := range strings.Split(hashList, ",") {
				var r Report
				var err error
				var hashPart string
				if strings.Contains(hash, "*") {
					parts := strings.Split(hash, "*")
					hashPart = parts[0]
					r.Count, err = strconv.Atoi(parts[1])
					if err != nil {
						return nil, nil, fmt.Errorf("%s: parse hash count: %v", f.Filename, err)
					}
				} else {
					r.Count = 1
					hashPart = hash
				}
				r.Hash, err = strconv.ParseUint(hashPart, 36, 64)
				if err != nil {
					return nil, nil, fmt.Errorf("%s: parse hash: %v", f.Filename, err)
				}
				reports[r.Hash] = r
			}
		}

		files[f.Filename] = FileProfile{
			Filename: f.Filename,
			Reports:  reports,
		}
	}

	result := &Profile{
		LinterVersion: p.LinterVersion,
		CreatedAt:     p.CreatedAt,
		Files:         files,
	}
	return result, p.Stats, nil
}

// WriteProfile writes a given suppression profile to w.
//
// Stats are included into the output as well.
func WriteProfile(w io.Writer, p *Profile, stats *Stats) error {
	const hashesPerLine = 15

	files := make([]jsonFileProfile, 0, len(p.Files))
	for filename, f := range p.Files {
		parts := make([]string, 0, len(f.Reports))
		for hash := range f.Reports {
			r := f.Reports[hash]
			part := strconv.FormatUint(hash, 36)
			if r.Count > 1 {
				part += fmt.Sprintf("*%d", r.Count)
			}
			parts = append(parts, part)
		}

		sort.Strings(parts)

		lists := make([]string, 0, len(parts)/hashesPerLine)
		var list bytes.Buffer
		i := 0
		for i < len(parts) {
			list.Reset()
			end := i + hashesPerLine
			for i < end && i < len(parts) {
				list.WriteString(parts[i])
				list.WriteByte(',')
				i++
			}
			list.Truncate(list.Len() - 1)
			lists = append(lists, list.String())
		}

		files = append(files, jsonFileProfile{
			Filename:  filename,
			HashLists: lists,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Filename < files[j].Filename
	})

	// We pretty-print the JSON to avoid a huge single-line file
	// that can't be handled by the most text editors.
	// Tabs give a nice indentation with the price of 1 byte.
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(jsonProfile{
		LinterVersion: p.LinterVersion,
		CreatedAt:     p.CreatedAt,
		Version:       profileVersion,
		Stats:         stats,
		Files:         files,
	})
}

// HashFields is a set of fields that are used during the report hash calculation.
type HashFields struct {
	Filename  string
	PrevLine  []byte
	StartLine []byte
	NextLine  []byte
	CheckName string
	Message   string
	Scope     string
}

// ReportHash computes the report signature hash for the baseline.
func ReportHash(scratchBuf *bytes.Buffer, fields HashFields) uint64 {
	const partSeparator = byte('#')

	buf := scratchBuf
	buf.WriteString(fields.Filename)
	buf.WriteByte(partSeparator)
	buf.WriteString(fields.Scope)
	buf.WriteByte(partSeparator)
	buf.Write(fields.PrevLine)
	buf.WriteByte(partSeparator)
	buf.Write(fields.StartLine)
	buf.WriteByte(partSeparator)
	buf.Write(fields.NextLine)
	buf.WriteByte(partSeparator)
	buf.WriteString(fields.CheckName)
	buf.WriteByte(partSeparator)
	buf.WriteString(fields.Message)

	// fnv64a gives the same results (no collisions) as md5 for the 3kk+ SLOC code base.
	hasher := fnv.New64a()
	hasher.Write(buf.Bytes())
	return hasher.Sum64()
}

// jsonProfile is a Profile representation that is used for JSON encoding/decoding.
// Using slices instead of maps guarantees the stable output as well as makes it more compact.
type jsonProfile struct {
	LinterVersion string
	CreatedAt     int64
	Version       int
	Stats         *Stats
	Files         []jsonFileProfile
}

// jsonFileProfile is a FileProfile representation that is used for JSON encoding/decoding.
// Using slices instead of maps guarantees the stable output as well as makes it more compact.
type jsonFileProfile struct {
	Filename string `json:"File"`

	// Every element is a comma separated list of hashes.
	// Split(list, ",") should produce a slice of hash strings.
	//
	// Every hash string can have one of the two forms:
	// "<hash>"         - Report{Count: 1, Hash: $hash}, eg. "38423934"
	// "<hash>*<count>" - Report{Count: $count, Hash: $hash}, eg. "283128*3"
	//
	// We do several steps to reduce the baseline file size:
	// - Hash value part is uint64 encoded as base-36 number.
	// - Since 99% of files have no collisions, we omit *1 for them.
	// - Multiple hashes are stored on one line.
	//
	// Having 100k-200k reports inside a baseline is not unrealistic.
	HashLists []string `json:"Hashes"`
}
