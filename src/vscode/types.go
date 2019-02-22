package vscode

// https://github.com/Microsoft/language-server-protocol/blob/master/versions/protocol-2-x.md

type Position struct {
	/**
	 * Line position in a document (zero-based).
	 */
	Line int `json:"line"`

	/**
	 * Character offset on a line in a document (zero-based).
	 */
	Character int `json:"character"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Range struct {
	/**
	 * The range's start position
	 */
	Start Position `json:"start"`

	/**
	 * The range's end position
	 */
	End Position `json:"end"`
}

const (
	/**
	 * Reports an error.
	 */
	Error = 1
	/**
	 * Reports a warning.
	 */
	Warning = 2
	/**
	 * Reports an information.
	 */
	Information = 3
	/**
	 * Reports a hint.
	 */
	Hint = 4
)

type Diagnostic struct {
	/**
	 * The range at which the message applies
	 */
	Range Range `json:"range"`

	/**
	 * The diagnostic's severity. Can be omitted. If omitted it is up to the
	 * client to interpret diagnostics as error, warning, info or hint.
	 */
	Severity int `json:"severity,omitempty"`

	/**
	 * The diagnostic's code. Can be omitted.
	 */
	Code string `json:"code,omitempty"`

	/**
	 * A human-readable string describing the source of this
	 * diagnostic, e.g. 'typescript' or 'super lint'.
	 */
	Source string `json:"source"`

	/**
	 * The diagnostic's message.
	 */
	Message string `json:"message"`

	/* Experimental "tags" feature for marking unused variables */
	Tags []int `json:"tags,omitempty"`
}

type PublishDiagnosticsParams struct {
	/**
	 * The URI for which diagnostic information is reported.
	 */
	URI string `json:"uri"`

	/**
	 * An array of diagnostic information items.
	 */
	Diagnostics []Diagnostic `json:"diagnostics"`
}

const (
	CompletionKindText        = 1
	CompletionKindMethod      = 2
	CompletionKindFunction    = 3
	CompletionKindConstructor = 4
	CompletionKindField       = 5
	CompletionKindVariable    = 6
	CompletionKindClass       = 7
	CompletionKindInterface   = 8
	CompletionKindModule      = 9
	CompletionKindProperty    = 10
	CompletionKindUnit        = 11
	CompletionKindValue       = 12
	CompletionKindEnum        = 13
	CompletionKindKeyword     = 14
	CompletionKindSnippet     = 15
	CompletionKindColor       = 16
	CompletionKindFile        = 17
	CompletionKindReference   = 18
)

type CompletionItem struct {
	/**
	 * The label of this completion item. By default
	 * also the text that is inserted when selecting
	 * this completion.
	 */
	Label string `json:"label"`
	/**
	 * The kind of this completion item. Based of the kind
	 * an icon is chosen by the editor.
	 */
	Kind int `json:"kind"`
	/**
	 * A human-readable string with additional information
	 * about this item, like type or symbol information.
	 */
	Detail string `json:"detail,omitempty"`
	/**
	 * A human-readable string that represents a doc-comment.
	 */
	Documentation string `json:"documentation,omitempty"`
	/**
	 * A string that shoud be used when comparing this item
	 * with other items. When `falsy` the label is used.
	 */
	SortText string `json:"sortText,omitempty"`
	/**
	 * A string that should be used when filtering a set of
	 * completion items. When `falsy` the label is used.
	 */
	// filterText?: string;
	/**
	 * A string that should be inserted a document when selecting
	 * this completion. When `falsy` the label is used.
	 */
	InsertText string `json:"insertText,omitempty"`
	/**
	 * An edit which is applied to a document when selecting
	 * this completion. When an edit is provided the value of
	 * insertText is ignored.
	 */
	// textEdit?: TextEdit;
	/**
	 * An data entry field that is preserved on a completion item between
	 * a completion and a completion resolve request.
	 */
	// data?: any
}

type SymbolInformation struct {
	/**
	 * The name of this symbol.
	 */
	Name string `json:"name"`

	/**
	 * The kind of this symbol.
	 */
	Kind int `json:"kind"`

	/**
	 * The location of this symbol.
	 */
	Location Location `json:"location"`

	/**
	 * The name of the symbol containing this symbol.
	 */
	ContainerName string `json:"containerName,omitempty"`
}

// enum SymbolKind
const (
	SymbolKindFile        = 1
	SymbolKindModule      = 2
	SymbolKindNamespace   = 3
	SymbolKindPackage     = 4
	SymbolKindClass       = 5
	SymbolKindMethod      = 6
	SymbolKindProperty    = 7
	SymbolKindField       = 8
	SymbolKindConstructor = 9
	SymbolKindEnum        = 10
	SymbolKindInterface   = 11
	SymbolKindFunction    = 12
	SymbolKindVariable    = 13
	SymbolKindConstant    = 14
	SymbolKindString      = 15
	SymbolKindNumber      = 16
	SymbolKindBoolean     = 17
	SymbolKindArray       = 18
)
