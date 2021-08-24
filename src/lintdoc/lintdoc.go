package lintdoc

import (
	"io"
	"strings"
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

// RenderMarkdownCheckDocumentation pretty-prints info to the provided writer with markdown syntax.
func RenderMarkdownCheckDocumentation(w io.Writer, info linter.CheckerInfo) error {
	if info.Before == "" {
		return templateMarkdownShort.Execute(w, info)
	}
	return templateMarkdownFull.Execute(w, info)
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

var templateMarkdownShort = template.Must(template.New("markdown-short").Parse(
	strings.ReplaceAll(`
### '{{.Name}}' checker
{{if .Quickfix}}
> Auto fix available
{{end}}
#### Description

{{.Comment}}

<p><br></p>`, "'", "`"),
))

var templateMarkdownFull = template.Must(template.New("markdown").Parse(
	strings.ReplaceAll(
		strings.ReplaceAll(`
### '{{.Name}}' checker
{{if .Quickfix}}
> Auto fix available
{{end}}
#### Description

{{.Comment}}

#### Non-compliant code:
"""php
{{.Before}}
"""

#### Compliant code:
"""php
{{.After}}
"""
<p><br></p>
`, `"""`, "```"),
		"'", "`"),
))
