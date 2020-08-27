package golden

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
)

func TestQuickFix(t *testing.T) {
	s := NewQuickFixTest(t, `testdata/quickfix`)
	s.runQuickFixTest()
}

const (
	expectedExtension = ".fix.expected"
	fixExtension      = ".fix"
)

type quickFixTest struct {
	t      *testing.T
	folder string
}

func NewQuickFixTest(t *testing.T, folder string) quickFixTest {
	return quickFixTest{
		t:      t,
		folder: folder,
	}
}

func openFile(filename string) (f *os.File, found bool, err error) {
	f, err = os.Open(filename)
	if err != nil {
		f, err = os.Create(filename)
		if err != nil {
			return nil, false, fmt.Errorf("file creation %s failed: %s", filename, err)
		}
		return f, false, nil
	}
	return f, true, nil
}

func (t *quickFixTest) runQuickFixTest() {
	files, err := linttest.FindPHPFiles(t.folder)
	if err != nil {
		t.t.Errorf("Error while searching for files in the %s folder: %s", t.folder, err)
	}

	for _, file := range files {
		t.t.Run(strings.TrimSuffix(filepath.Base(file), ".php"), func(t *testing.T) {
			testFileName := file
			expectedFileName := file + expectedExtension
			fixedFileName := file + fixExtension

			testFileContent, err := ioutil.ReadFile(testFileName)
			if err != nil {
				t.Errorf("Reading file %s failed: %s", testFileName, err)
			}

			expectedFile, expectedFileFound, err := openFile(expectedFileName)
			if err != nil {
				t.Errorf("File open %s failed: %s", expectedFileName, err)
			}
			defer expectedFile.Close()

			var expectedFileContent []byte
			if expectedFileFound {
				expectedFileContent, err = ioutil.ReadAll(expectedFile)
				if err != nil {
					t.Errorf("Reading file %s failed: %s", expectedFileName, err)
				}
			}

			fixedFile, _, err := openFile(fixedFileName)
			if err != nil {
				t.Errorf("File open %s failed: %s", fixedFileName, err)
			}
			_, _ = fixedFile.Write(testFileContent)
			fixedFile.Close()

			test := linttest.NewSuite(t)
			test.AddNamedFile(fixedFileName, string(testFileContent))
			linter.ApplyQuickFixes = true
			defer func() {
				linter.ApplyQuickFixes = false
			}()
			_ = test.RunLinter()

			fixedFileContent, err := ioutil.ReadFile(fixedFileName)
			if err != nil {
				t.Errorf("Reading file %s failed: %s", fixedFileName, err)
			}

			if !expectedFileFound {
				_, _ = expectedFile.Write(fixedFileContent)
				t.Logf("The expected files for \"%s\" were not found and were generated automatically.", filepath.Base(testFileName))
				return
			}

			expected := string(expectedFileContent)
			want := string(fixedFileContent)

			if !cmp.Equal(expected, want) {
				t.Error(cmp.Diff(expected, want))
			}

			if !t.Failed() {
				err = os.Remove(fixedFileName)
				if err != nil {
					t.Errorf("Removing file %s failed: %s", fixedFileName, err)
				}
			}
		})
	}
}
