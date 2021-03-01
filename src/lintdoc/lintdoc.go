package lintdoc

import (
	"io"
	"text/template"

	"github.com/VKCOM/noverify/src/linter"
)

// RenderCheckDocumentation pretty-prints info to the provided writer.
func RenderCheckDocumentation(w io.Writer, info linter.CheckerInfo) error {
	if info.Before == "" {
		return templateShort.Execute(w, info)
	}
	return templateFull.Execute(w, info)
}

var templateShort = template.Must(template.New("short").Parse(`
{{- .Name}} checker documentation

{{.Comment -}}
`))

var templateFull = template.Must(template.New("short").Parse(`
{{- .Name}} checker documentation{{if .Quickfix}} (auto fix available){{end}}

{{.Comment}}

Non-compliant code:
{{.Before}}

Compliant code:
{{.After -}}
`))
