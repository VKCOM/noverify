package meta

import (
	"testing"
)

func TestDeleteMetaPreservesLongerClassDefinition(t *testing.T) {
	info := NewInfo()

	longMethods := NewFunctionsMap()
	longMethods.Set("getMessage", FuncInfo{Name: `getMessage`})

	shortMethods := NewFunctionsMap()

	stubsFile := "stubs/Core_c.php"
	projectFile := "vendor/codeception/autoload.php"

	info.AddClassesNonLocked(stubsFile, ClassesMap{H: map[lowercaseString]ClassInfo{
		`\throwable`: {
			Name:    `\Throwable`,
			Pos:     ElementPosition{Filename: stubsFile, Length: 2231},
			Methods: longMethods,
		},
	}})

	info.AddClassesNonLocked(projectFile, ClassesMap{H: map[lowercaseString]ClassInfo{
		`\throwable`: {
			Name:    `\Throwable`,
			Pos:     ElementPosition{Filename: projectFile, Length: 22},
			Methods: shortMethods,
		},
	}})

	cls, ok := info.GetClass(`\Throwable`)
	if !ok {
		t.Fatal("expected Throwable to exist after adding both definitions")
	}
	if cls.Methods.Len() != 1 {
		t.Fatalf("expected 1 method (from stubs), got %d", cls.Methods.Len())
	}

	info.DeleteMetaForFileNonLocked(projectFile)

	cls, ok = info.GetClass(`\Throwable`)
	if !ok {
		t.Fatal("expected Throwable to still exist after deleting project file meta")
	}
	if cls.Methods.Len() != 1 {
		t.Fatalf("expected 1 method (from stubs) to survive re-indexing, got %d", cls.Methods.Len())
	}
	if cls.Pos.Length != 2231 {
		t.Fatalf("expected stub definition (len=2231) to survive, got len=%d", cls.Pos.Length)
	}
}

func TestDeleteMetaRemovesMatchingClassDefinition(t *testing.T) {
	info := NewInfo()

	methods := NewFunctionsMap()
	methods.Set("getMessage", FuncInfo{Name: `getMessage`})

	file := "stubs/Core_c.php"

	info.AddClassesNonLocked(file, ClassesMap{H: map[lowercaseString]ClassInfo{
		`\throwable`: {
			Name:    `\Throwable`,
			Pos:     ElementPosition{Filename: file, Length: 2231},
			Methods: methods,
		},
	}})

	info.DeleteMetaForFileNonLocked(file)

	_, ok := info.GetClass(`\Throwable`)
	if ok {
		t.Fatal("expected Throwable to be removed when deleting its own file meta")
	}
}
