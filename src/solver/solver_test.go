package solver

import (
	"reflect"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/meta"
)

func resolve(info *meta.Info, typ string) map[string]struct{} {
	return resolveType(info, "", typ, make(ResolverMap))
}

func makeTyp(typ string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, t := range strings.Split(typ, "|") {
		res[t] = struct{}{}
	}
	return res
}

func typesEqual(a map[string]struct{}, b string) bool {
	return reflect.DeepEqual(a, makeTyp(b))
}

func TestSolver(t *testing.T) {
	tm := meta.NewTypesMap

	sc := meta.NewScope()
	sc.AddVarName("MC", tm("Memcache"), "global", meta.VarAlwaysDefined)

	fm := meta.NewFunctionsMap()
	fm.Set(`\array_map`, meta.FuncInfo{Typ: tm(`array|bool|` + meta.WrapFunctionCall(`\my_func`))})
	fm.Set(`\my_func`, meta.FuncInfo{Typ: tm(meta.WrapFunctionCall(`\array_map`) + `|float`)})

	cmfm := meta.NewFunctionsMap()
	cmfm.Set(`do_something`, meta.FuncInfo{Typ: tm(`string`)})

	cm := meta.NewClassesMap()
	cm.Set(`\Test`, meta.ClassInfo{
		Methods: cmfm,
		Properties: meta.PropertiesMap{
			`$instance`: {Typ: tm(`\Test`)},
		},
	})

	metainfo := meta.NewInfo()
	metainfo.AddToGlobalScopeNonLocked("test", sc)
	metainfo.AddFunctionsNonLocked("test", fm)
	metainfo.AddClassesNonLocked("test", cm)

	if typ := resolve(metainfo, meta.WrapFunctionCall(`\my_func`)); !typesEqual(typ, `array|bool|float`) {
		t.Errorf("My func wrong type: %+v", typ)
	}

	if typ := resolve(metainfo, meta.WrapGlobal(`MC`)); !typesEqual(typ, `Memcache`) {
		t.Errorf("Global $MC wrong: %+v", typ)
	}

	if typ := resolve(metainfo, meta.WrapStaticPropertyFetch(`\Test`, `$instance`)); !typesEqual(typ, `\Test`) {
		t.Errorf(`\Test::$instance wrong: %+v`, typ)
	}

	if typ := resolve(metainfo, meta.WrapInstanceMethodCall(meta.WrapStaticPropertyFetch(`\Test`, `$instance`), `do_something`)); !typesEqual(typ, `string`) {
		t.Errorf(`\Test::$instance::do_something() wrong: %+v`, typ)
	}
}
