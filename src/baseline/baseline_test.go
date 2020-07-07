package baseline

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWriteReadBaseline(t *testing.T) {
	type testHashFields struct {
		Lines     [3]string
		CheckName string
		Message   string
		Scope     string
	}

	makeReports := func(filename string, testFieldsList []testHashFields) map[uint64]Report {
		fieldsList := make([]HashFields, len(testFieldsList))
		counters := make(map[uint64]int, len(testFieldsList))
		for i, fields := range testFieldsList {
			fieldsList[i] = HashFields{
				Filename:  filename,
				PrevLine:  []byte(fields.Lines[0]),
				StartLine: []byte(fields.Lines[1]),
				NextLine:  []byte(fields.Lines[2]),
				CheckName: fields.CheckName,
				Message:   fields.Message,
				Scope:     fields.Scope,
			}
			fieldsList[i].Filename = filename
			counters[ReportHash(fieldsList[i])]++
		}

		reports := make(map[uint64]Report, len(fieldsList))
		for _, fields := range fieldsList {
			hash := ReportHash(fields)
			reports[hash] = Report{
				Hash:  hash,
				Count: counters[hash],
			}
		}

		return reports
	}
	makeFileProfile := func(filename string, testFieldsList []testHashFields) FileProfile {
		return FileProfile{
			Filename: filename,
			Reports:  makeReports(filename, testFieldsList),
		}
	}
	makeProfile := func(linterVersion string, files ...FileProfile) *Profile {
		m := make(map[string]FileProfile, len(files))
		for _, f := range files {
			m[f.Filename] = f
		}
		return &Profile{
			LinterVersion: linterVersion,
			Files:         m,
		}
	}

	const expectedOutput = `{
	"LinterVersion": "3cfde307d8fbb5acd13d3c346b442172c4433dcb",
	"Version": 2,
	"Stats": {
		"CountTotal": 0,
		"CountPerCheck": null
	},
	"Files": [
		{
			"File": "/a/Bar.php",
			"Hashes": [
				"zr6nxkhb4gdh"
			]
		},
		{
			"File": "/a/Foo.php",
			"Hashes": [
				"19nhmzlil1g8o,1f8786sb08uia,35fa1tga0130h*2"
			]
		}
	]
}
`

	f1 := makeFileProfile("/a/Foo.php", []testHashFields{
		// Two same code patterns inside one method.
		{
			Lines: [3]string{
				"  if ($cond) {",
				"    die('test');",
				"  }",
			},
			Scope:     "Foo::init",
			CheckName: "noDie",
			Message:   "Do not debug with die('test')",
		},
		{
			Lines: [3]string{
				"  if ($cond) {",
				"    die('test');",
				"  }",
			},
			Scope:     "Foo::init",
			CheckName: "noDie",
			Message:   "Do not debug with die('test')",
		},

		// die('test') with different lines around it.
		{
			Lines: [3]string{
				"  if (false) {",
				"    die('test');",
				"  }",
			},
			Scope:     "Foo::init",
			CheckName: "noDie",
			Message:   "Do not debug with die('test')",
		},

		// Completely different warning.
		{
			Lines: [3]string{
				"    $i++;",
				"    $foo->save($i);",
				"  }",
			},
			Scope:     "foo_increment_counter",
			CheckName: "noDie",
			Message:   "Calling undefined method $foo->save()",
		},
	})

	f2 := makeFileProfile("/a/Bar.php", []testHashFields{
		// More die('test'), now inside another class.
		{
			Lines: [3]string{
				"  if ($cond) {",
				"    die('test');",
				"  }",
			},
			Scope:     "Bar::init",
			CheckName: "noDie",
			Message:   "Do not debug with die('test')",
		},
	})

	x := makeProfile("3cfde307d8fbb5acd13d3c346b442172c4433dcb", f1, f2)

	// Run test more than once to verify that the output is stable.
	for i := 0; i < 10; i++ {
		var buf bytes.Buffer
		if err := WriteProfile(&buf, x, &Stats{}); err != nil {
			t.Fatalf("iter=%d error while encoding profile: %v", i, err)
		}
		if diff := cmp.Diff(buf.String(), expectedOutput); diff != "" {
			t.Fatalf("iter=%d printed output differs:\n%s", i, diff)
		}

		y, err := ReadProfile(&buf)
		if err != nil {
			t.Fatalf("iter=%d error while reading encoded profile: %v", i, err)
		}

		if diff := cmp.Diff(x, y); diff != "" {
			t.Fatalf("iter=%d decoded profile differs:\n%s", i, diff)
		}
	}

	// Test bigger profile size.
	// This test is useful to monitor how our changes affect
	// the profile size.
	files := make([]FileProfile, 200)
	for i := range files {
		fieldsList := make([]testHashFields, 30)
		for j := range fieldsList {
			fieldsList[j] = testHashFields{
				Lines:   [3]string{"a", "b", "c"},
				Scope:   fmt.Sprintf("func%d", i),
				Message: "baseline file size test",
			}
		}
		files[i] = makeFileProfile(fmt.Sprintf("file%d.php", i), fieldsList)
	}
	bigProfile := makeProfile("3cfde307d8fbb5acd13d3c346b442172c4433dcb", files...)
	var buf bytes.Buffer
	if err := WriteProfile(&buf, bigProfile, &Stats{}); err != nil {
		t.Fatalf("encoding big profile: %v", err)
	}
	expectedSize := 15597
	if expectedSize != buf.Len() {
		t.Fatalf("big profile size differs:\nhave: %d\nwant: %d", buf.Len(), expectedSize)
	}
}
