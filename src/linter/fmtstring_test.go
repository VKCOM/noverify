package linter

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseFormatString(t *testing.T) {
	tests := []struct {
		input      string
		directives []string
	}{
		{input: ``, directives: []string{}},
		{input: `foo`, directives: []string{}},

		{
			input: `foo%d%0$fbar`,
			directives: []string{
				`%d arg=1 (implicit)`,
				`%f arg=0`,
			},
		},

		{
			input: `%10.f%.5f`,
			directives: []string{
				`%f arg=1 (implicit) w=10`,
				`%f arg=2 (implicit) p=5`,
			},
		},

		{
			input: ` %10.f %.5f `,
			directives: []string{
				`%f arg=1 (implicit) w=10`,
				`%f arg=2 (implicit) p=5`,
			},
		},

		{
			input: ` %10$+-+500.71g `,
			directives: []string{
				`%g arg=10 p=71 w=500 flags=+-+`,
			},
		},

		{
			input: `%10$+-+500g`,
			directives: []string{
				`%g arg=10 w=500 flags=+-+`,
			},
		},

		// Explicit arg nums do not increment the arg counter.
		{
			input: `%1$d %1d %d`,
			directives: []string{
				"%d arg=1",
				"%d arg=1 (implicit) w=1",
				"%d arg=2 (implicit)",
			},
		},

		// %% does not increment the arg counter.
		{
			input: `%% %4$% %d %.5% %d`,
			directives: []string{
				"%%",
				"%% arg=4",
				"%d arg=1 (implicit)",
				"%% p=5",
				"%d arg=2 (implicit)",
			},
		},
	}

	for _, test := range tests {
		parsed, err := parseFormatString(test.input)
		if err != nil {
			t.Errorf("parse `%s`: %v", test.input, err)
			continue
		}
		directives := make([]string, len(parsed.directives))
		for i, d := range parsed.directives {
			directives[i] = d.String()
		}
		wantDirectives := test.directives
		haveDirectives := directives
		if diff := cmp.Diff(wantDirectives, haveDirectives); diff != "" {
			t.Errorf("directives mismatch (+ have) (- want): %s", diff)
			continue
		}
	}
}
