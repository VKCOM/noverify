package regression_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/git"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/workspace"
)

const classFile = "example_service.php"

const classOld = `<?php

class ExampleService
{
    public static $instance;
}
`

const classNew = `<?php

class ExampleService
{
    public static $instance;

    protected function caller(string $a): string {
        return $this->callee($a);
    }
    protected function callee(int $a): int {
        return $a;
    }
}
`

func diffModeReports(t *testing.T, oldContents, newContents string) []*linter.Report {
	t.Helper()

	l := linter.NewLinter(linter.NewConfig("8.1"))

	readFile := func(contents string) workspace.ReadCallback {
		return func(ch chan workspace.FileInfo) {
			ch <- workspace.FileInfo{Name: classFile, Contents: []byte(contents)}
		}
	}

	// Step 1: index all files at the new commit.
	l.AnalyzeFiles(readFile(newContents))

	// Step 2: indexing complete.
	l.MetaInfo().SetIndexingComplete(true)

	// Step 3: parse old file versions (diff mode).
	oldReports := l.AnalyzeFiles(readFile(oldContents))

	// Step 4: re-index changed files.
	l.MetaInfo().SetIndexingComplete(false)
	l.AnalyzeFiles(readFile(newContents))
	l.MetaInfo().SetIndexingComplete(true)

	// Step 5: parse new file versions.
	newReports := l.AnalyzeFiles(readFile(newContents))

	changes := []git.Change{{
		OldName: classFile,
		NewName: classFile,
		Type:    git.Changed,
		Valid:   true,
	}}

	diff, err := linter.DiffReports("", nil, changes, nil, oldReports, newReports, 1)
	if err != nil {
		t.Fatalf("DiffReports: %v", err)
	}
	return diff
}

func classMethodCount(l *linter.Linter) int {
	cl, ok := l.MetaInfo().GetClass(`\ExampleService`)
	if !ok {
		return 0
	}
	return cl.Methods.Len()
}

func hasUndefinedMethod(reports []*linter.Report, method string) bool {
	for _, r := range reports {
		if r.CheckName == "undefinedMethod" && strings.Contains(r.Message, method) {
			return true
		}
	}
	return false
}

func hasReport(reports []*linter.Report, checkName string) bool {
	for _, r := range reports {
		if r.CheckName == checkName {
			return true
		}
	}
	return false
}

func diffModeReportsFromOldIndex(t *testing.T, oldContents, newContents string) []*linter.Report {
	t.Helper()

	l := linter.NewLinter(linter.NewConfig("8.1"))

	readFile := func(contents string) workspace.ReadCallback {
		return func(ch chan workspace.FileInfo) {
			ch <- workspace.FileInfo{Name: classFile, Contents: []byte(contents)}
		}
	}

	// Local work-tree flow indexes merge-base first (git_main.go).
	l.AnalyzeFiles(readFile(oldContents))
	l.MetaInfo().SetIndexingComplete(true)

	oldReports := l.AnalyzeFiles(readFile(oldContents))

	l.MetaInfo().SetIndexingComplete(false)
	l.AnalyzeFiles(readFile(newContents))
	l.MetaInfo().SetIndexingComplete(true)

	newReports := l.AnalyzeFiles(readFile(newContents))

	changes := []git.Change{{
		OldName: classFile,
		NewName: classFile,
		Type:    git.Changed,
		Valid:   true,
	}}

	diff, err := linter.DiffReports("", nil, changes, nil, oldReports, newReports, 1)
	if err != nil {
		t.Fatalf("DiffReports: %v", err)
	}
	return diff
}

func TestDiffModeSameFileMethodCall(t *testing.T) {
	reports := diffModeReports(t, classOld, classNew)
	if hasUndefinedMethod(reports, "callee") {
		t.Fatalf("unexpected undefinedMethod for callee in diff mode: %+v", reports)
	}
}

func TestDiffModeSameFileMethodCallFromOldIndex(t *testing.T) {
	reports := diffModeReportsFromOldIndex(t, classOld, classNew)
	if hasUndefinedMethod(reports, "callee") {
		t.Fatalf("unexpected undefinedMethod for callee when indexed from old commit: %+v", reports)
	}
}

func TestDiffModeReindexUpdatesGlobalMeta(t *testing.T) {
	l := linter.NewLinter(linter.NewConfig("8.1"))
	readFile := func(contents string) workspace.ReadCallback {
		return func(ch chan workspace.FileInfo) {
			ch <- workspace.FileInfo{Name: classFile, Contents: []byte(contents)}
		}
	}

	l.AnalyzeFiles(readFile(classOld))
	if n := classMethodCount(l); n != 0 {
		t.Fatalf("old index: want 0 methods, got %d", n)
	}

	l.MetaInfo().SetIndexingComplete(true)
	l.MetaInfo().SetIndexingComplete(false)
	l.AnalyzeFiles(readFile(classNew))

	if n := classMethodCount(l); n != 2 {
		t.Fatalf("after reindex: want 2 methods, got %d", n)
	}
}

func diffModeReportsWithoutReindex(t *testing.T, oldContents, newContents string) []*linter.Report {
	t.Helper()

	l := linter.NewLinter(linter.NewConfig("8.1"))

	readFile := func(contents string) workspace.ReadCallback {
		return func(ch chan workspace.FileInfo) {
			ch <- workspace.FileInfo{Name: classFile, Contents: []byte(contents)}
		}
	}

	l.AnalyzeFiles(readFile(oldContents))
	l.MetaInfo().SetIndexingComplete(true)

	oldReports := l.AnalyzeFiles(readFile(oldContents))

	// Skip re-indexing changed files (bug scenario).
	l.MetaInfo().SetIndexingComplete(false)
	l.MetaInfo().SetIndexingComplete(true)

	newReports := l.AnalyzeFiles(readFile(newContents))

	changes := []git.Change{{
		OldName: classFile,
		NewName: classFile,
		Type:    git.Changed,
		Valid:   true,
	}}

	diff, err := linter.DiffReports("", nil, changes, nil, oldReports, newReports, 1)
	if err != nil {
		t.Fatalf("DiffReports: %v", err)
	}
	return diff
}

func TestDiffModeWithoutReindexFromOldIndex(t *testing.T) {
	reports := diffModeReportsWithoutReindex(t, classOld, classNew)
	if hasUndefinedMethod(reports, "callee") {
		t.Fatalf("same-file method calls must work even when global index is stale: %+v", reports)
	}
	if !hasReport(reports, "notSafeCall") {
		t.Fatalf("argument type mismatch must still be reported in diff mode: %+v", reports)
	}
}

func TestFullFileSameClassMethodCall(t *testing.T) {
	l := linter.NewLinter(linter.NewConfig("8.1"))
	readClassFile := func(ch chan workspace.FileInfo) {
		ch <- workspace.FileInfo{Name: classFile, Contents: []byte(classNew)}
	}
	l.AnalyzeFiles(readClassFile)
	l.MetaInfo().SetIndexingComplete(true)
	reports := l.AnalyzeFiles(readClassFile)
	if hasUndefinedMethod(reports, "callee") {
		t.Fatalf("unexpected undefinedMethod for callee in full file mode: %+v", reports)
	}
}
