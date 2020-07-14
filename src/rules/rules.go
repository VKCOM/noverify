package rules

import (
	"io"

	"github.com/VKCOM/noverify/src/phpdoc"
	"github.com/VKCOM/noverify/src/phpgrep"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (*Parser) Parse(filename string, r io.Reader) (*Set, error) {
	p := parser{typeParser: phpdoc.NewTypeParser()}
	return p.parse(filename, r)
}

// NewSet returns a new empty rules set.
func NewSet() *Set {
	return &Set{
		Any:   &ScopedSet{},
		Root:  &ScopedSet{},
		Local: &ScopedSet{},
	}
}

// Set is a result of rule file parsing.
type Set struct {
	Any     *ScopedSet // Anywhere
	Root    *ScopedSet // Only outside of functions
	Local   *ScopedSet // Only inside functions
	Builtin bool       // Whether this is a NoVerify builtin rule set

	Names []string // All rule names
}

// ScopedSet is a categorized rules collection.
// Categories help to assign a better execution strategy for a rule.
type ScopedSet struct {
	RulesByKind [_KindCount][]Rule
}

// Rule is a dynamically-loaded linter rule.
//
// A rule is called unnamed if no @name attribute is given.
// Unnamed rules receive auto-generated name that includes
// a rule file name and a line that defines that rule.
type Rule struct {
	// Name tells whether this rule causes critical report.
	Name string

	// Matcher is an object that is used to check whether a given AST node
	// should trigger a warning that is associated with rule.
	Matcher *phpgrep.Matcher

	// Level is a severity level that is used during report generation.
	Level int

	// Message is a report text that is printed when this rule matches.
	Message string

	// Fix is a quickfix template.
	Fix string

	// Location is a phpgrep variable name that should be used as a warning location.
	// Empty string selects the root node.
	Location string

	// Path is a filter-like rule switcher.
	// A rule is only applied to a file that contains a Path as a substring in its name.
	Path string

	// Filters is a list of OR-connected filter sets.
	// Every filter set is a mapping of phpgrep variable to a filter.
	Filters []map[string]Filter

	scope string
}

// String returns a rule printer representation.
func (r *Rule) String() string {
	return formatRule(r)
}

// Filter describes constraints that should be applied to a given phpgrep variable.
type Filter struct {
	Type *phpdoc.Type
	Pure bool
}
