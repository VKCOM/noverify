package linter

var (
	// LangServer represents whether or not we run in a language server mode.
	LangServer bool

	CacheDir string

	// AnalysisFiles is a list of files that are being analyzed (in non-git mode)
	AnalysisFiles []string

	// settings
	StubsDir        string
	Debug           bool
	MaxConcurrency  int
	MaxFileSize     int
	DefaultEncoding string

	// actually time.Duration
	initParseTime int64
	initWalkTime  int64
)
