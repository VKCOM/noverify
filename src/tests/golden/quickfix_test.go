package golden

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

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

type QuickFixTest struct {
	t      *testing.T
	folder string
}

func NewQuickFixTest(t *testing.T, folder string) QuickFixTest {
	return QuickFixTest{
		t:      t,
		folder: folder,
	}
}

func (t *QuickFixTest) runQuickFixTest() {
	files, err := linttest.FindPHPFiles(t.folder)
	if err != nil {
		t.t.Errorf("Error while searching for files in the %s folder: %s", t.folder, err)
	}

	for _, file := range files {
		t.t.Run(strings.TrimSuffix(filepath.Base(file), ".php"), func(t *testing.T) {
			testFileName := file
			expectedFileName := file + expectedExtension
			fixedFileName := file + fixExtension

			testFile, err := os.Open(testFileName)
			if err != nil {
				t.Errorf("File %s not open: %s", testFileName, err)
				return
			}
			testFileContent, err := ioutil.ReadAll(testFile)
			if err != nil {
				t.Errorf("Reading file %s failed: %s", testFileName, err)
			}
			testFile.Close()

			expectedFile, err := os.Open(expectedFileName)
			expectedFileFound := true
			if err != nil || expectedFile == nil {
				expectedFileFound = false
				expectedFile, err = os.Create(expectedFileName)
				if err != nil {
					t.Errorf("File creation %s failed: %s", expectedFileName, err)
				}
			}

			var expectedFileContent []byte
			if expectedFileFound {
				expectedFileContent, err = ioutil.ReadAll(expectedFile)
				if err != nil {
					t.Errorf("Reading file %s failed: %s", expectedFileName, err)
				}
				expectedFile.Close()
			}

			fixedFile, err := os.Open(fixedFileName)
			if err != nil || fixedFile == nil {
				fixedFile, err = os.Create(fixedFileName)
				if err != nil {
					t.Errorf("File creation %s failed: %s", fixedFileName, err)
				}
			}
			_, _ = fixedFile.Write(testFileContent)
			fixedFile.Close()

			test := linttest.NewSuite(t)
			test.ApplyQuickFixes = true
			test.AddNamedFile(fixedFileName, string(testFileContent))
			_ = test.RunLinter()

			fixedFile, err = os.Open(fixedFileName)
			if err != nil {
				t.Errorf("File %s not open: %s", fixedFileName, err)
			}
			fixedFileContent, err := ioutil.ReadAll(fixedFile)
			if err != nil {
				t.Errorf("Reading file %s failed: %s", fixedFileName, err)
			}
			fixedFile.Close()

			if !expectedFileFound {
				_, _ = expectedFile.Write(fixedFileContent)
				expectedFile.Close()
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
