package linter

import (
	"github.com/VKCOM/noverify/src/inputs"
)

var (
	// LangServer represents whether or not we run in a language server mode.
	LangServer bool

	CacheDir string

	// AnalysisFiles is a list of files that are being analyzed (in non-git mode)
	AnalysisFiles []string

	// SrcInput implements source code reading from files and buffers.
	//
	// TODO(quasilyte): avoid having it as a global variable?
	SrcInput = inputs.NewDefaultSourceInput()

	// settings
	StubsDir        string
	Debug           bool
	MaxConcurrency  int
	MaxFileSize     int
	DefaultEncoding string
	PHPExtensions   []string

	IsDiscardVar = isUnderscore

	// actually time.Duration
	initParseTime int64
	initWalkTime  int64
)
