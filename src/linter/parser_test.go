package linter

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestReadFilenamesConcurrently(t *testing.T) {
	PHPExtensions = []string{"php"}

	dir, err := ioutil.TempDir(os.TempDir(), "testfilenames")
	if err != nil {
		t.Fatalf("creating tmpdir: %v", err)
	}
	defer os.RemoveAll(dir)

	files := []string{
		"file1.php",
		"file2.php",
		"file3.php",
		"dir1/d1file1.php",
		"dir1/d1file2.php",
		"dir1/ignored.php",
		"dir2/d2file1.php",
		"dir2/d2file2.php",
	}

	var want []string

	for _, file := range files {
		path := filepath.Join(dir, file)
		subdir := filepath.Dir(path)
		if err := os.MkdirAll(subdir, 0777); err != nil {
			t.Fatalf("Failed to create %q: %v", subdir, err)
		}

		if err := ioutil.WriteFile(path, []byte("<?php echo 'test';"), 0666); err != nil {
			t.Fatalf("Failed to create %q: %v", path, err)
		}

		if !strings.Contains(file, "ignored") {
			want = append(want, path)
		}
	}

	limitCh := make(chan struct{}, 3)
	ch := make(chan FileInfo)

	go func() {
		readFilenamesConcurrently(dir, ch, regexp.MustCompile("ignored"), limitCh)
		close(ch)
	}()

	res := make([]string, 0, len(files))
	for file := range ch {
		res = append(res, file.Filename)
	}

	sort.Strings(res)
	sort.Strings(want)

	if !reflect.DeepEqual(want, res) {
		t.Fatalf("Different file lists: want %+v, got %+v", files, res)
	}
}
