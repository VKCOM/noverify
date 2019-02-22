package vscode

type Capability struct {
	DynamicRegistration bool `json:"dynamicRegistration"`
	WillSave            bool `json:"willSave"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil"`
	DidSave             bool `json:"didSave"`
}

type WorkspaceCapabilities struct {
	ApplyEdit              bool
	DidChangeConfiguration Capability `json:"didChangeConfiguration"`
	DidChangeWatchedFiles  Capability `json:"didChangeWatchedFiles"`
	Symbol                 Capability `json:"symbol"`
	ExecuteCommand         Capability `json:"executeCommand"`
}

type TextDocumentCapabilities struct {
	Synchronization   Capability `json:"synchronization"`
	Completion        Capability `json:"completion"`
	Hover             Capability `json:"hover"`
	SignatureHelp     Capability `json:"signatureHelp"`
	Definition        Capability `json:"definition"`
	References        Capability `json:"references"`
	DocumentHighlight Capability `json:"documentHighlight"`
	DocumentSymbol    Capability `json:"documentSymbol"`
	CodeAction        Capability `json:"codeAction"`
	CodeLens          Capability `json:"codeLens"`
	Formatting        Capability `json:"formatting"`
	RangeFormatting   Capability `json:"rangeFormatting"`
	OnTypeFormatting  Capability `json:"onTypeFormatting"`
	Rename            Capability `json:"rename"`
	DocumentLink      Capability `json:"documentLink"`
}

type CapabilitiesSections struct {
	Workspace    WorkspaceCapabilities    `json:"workspace"`
	TextDocument TextDocumentCapabilities `json:"textDocument"`
}

type TextDocumentDidOpenParams struct {
	TextDocument struct {
		URI        string `json:"uri"`
		LanguageID string `json:"languageId"`
		Version    int    `json:"version"`
		Text       string `json:"text"`
	} `json:"textDocument"`
}

type InitializeParams struct {
	ProcessID    int                  `json:"processId"`
	RootPath     string               `json:"rootPath"`
	RootURI      string               `json:"rootUri"`
	Capabilities CapabilitiesSections `json:"capabilities"`
	Trace        string               `json:"trace"`
}

type ContentChange struct {
	Text string `json:"text"`
}

type TextDocumentDidChangeParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []ContentChange `json:"contentChanges"`
}

type DefinitionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

type ReferencesParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
	Context  struct {
		IncludeDeclaration bool `json:"includeDeclaration"`
	} `json:"context"`
}

const (
	Created = 1
	Changed = 2
	Deleted = 3
)

type FileEvent struct {
	URI  string `json:"uri"`
	Type int    `json:"type"`
}

type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}
