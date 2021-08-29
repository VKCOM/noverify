package linter

import (
	"fmt"
	"testing"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/phpdoctypes"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/workspace"
)

func TestParseClassPHPDoc(t *testing.T) {
	tests := []struct {
		line     string
		method   string
		property string
		typ      string
	}{
		{
			line:   `@method int foo`,
			method: `foo`,
			typ:    `int`,
		},
		{
			line:   `@method Foo m()`,
			method: `m`,
			typ:    `\Foo`,
		},
		{
			line:   `@method \A\B m2(int $x, ...$rest)`,
			method: `m2`,
			typ:    `\A\B`,
		},
		{
			line:   `@method integer m()`,
			method: `m`,
			typ:    `int`,
		},

		{
			line:     `@property $x int`,
			property: `x`,
			typ:      `int`,
		},
		{
			line:     `@property int $x`,
			property: `x`,
			typ:      `int`,
		},
		{
			line:     `@property int $cost -- item cost in $`,
			property: `cost`,
			typ:      `int`,
		},
		{
			line:     `@property $enabled boolean`,
			property: `enabled`,
			typ:      `bool`,
		},
	}

	l := NewLinter(NewConfig("8.1"))
	st := &meta.ClassParseState{Info: l.MetaInfo()}
	walker := rootWalker{
		ctx: newRootContext(l.config, NewWorkerContext(), st),
	}
	walker.checker = newRootChecker(&walker, NewQuickFixGenerator(workspace.NewFile("test.php", []byte{})))
	for _, test := range tests {
		docString := fmt.Sprintf(`/** %s */`, test.line)
		doc := phpdoc.Parse(walker.ctx.phpdocTypeParser, docString)
		result := walker.parseClassPHPDoc(nil, doc)

		switch {
		case test.method != "":
			if result.methods.Len() != 1 {
				t.Errorf("parse(`%s`): expected 1 methods, found %d", test.line, result.methods.Len())
				continue
			}
			m, ok := result.methods.Get(test.method)
			if !ok {
				foundInstead := ""
				for _, info := range result.methods.H {
					foundInstead = info.Name
					break
				}
				t.Errorf("parse(`%s`): expected %s method, found %s", test.line, test.method, foundInstead)
				continue
			}
			if m.Typ.String() != test.typ {
				t.Errorf("parse(`%s`): expected %s type, found %s", test.line, test.typ, m.Typ.String())
				continue
			}

		case test.property != "":
			if len(result.properties) != 1 {
				t.Errorf("parse(`%s`): expected 1 properties, found %d", test.line, len(result.properties))
				continue
			}
			m, ok := result.properties[test.property]
			if !ok {
				foundInstead := ""
				for name := range result.properties {
					foundInstead = name
					break
				}
				t.Errorf("parse(`%s`): expected %s property, found %s", test.line, test.property, foundInstead)
				continue
			}
			if m.Typ.String() != test.typ {
				t.Errorf("parse(`%s`): expected %s type, found %s", test.line, test.typ, m.Typ.String())
				continue
			}
		}
	}
}

func BenchmarkParseTypes(b *testing.B) {
	l := NewLinter(NewConfig("8.1"))
	st := &meta.ClassParseState{}
	ctx := newRootContext(l.config, NewWorkerContext(), st)
	typeString := `?x|array<int>|T[]`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedType := ctx.phpdocTypeParser.Parse(typeString)

		converted := phpdoctypes.ToRealType(ctx.typeNormalizer.ClassFQNProvider(), parsedType)
		moveShapesToContext(&ctx, converted.Shapes)

		_ = types.NewMapWithNormalization(ctx.typeNormalizer, converted.Types)
	}
}
