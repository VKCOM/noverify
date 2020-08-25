package linter

import (
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/git"
)

const (
	// IgnoreLinterMessage is a commit message that you specify if you want to cancel linter checks for this changeset
	IgnoreLinterMessage = "@linter disable"
)

func init() {
	allChecks := []CheckInfo{
		{
			Name:    "intOverflow",
			Default: true,
			Comment: `Report potential integer overflows that may result in unexpected behavior.`,
			Before:  `return -9223372036854775808;`,
			After:   `return PHP_INT_MIN;`,
		},

		{
			Name:    "discardExpr",
			Default: true,
			Comment: `Report expressions that are evaluated but not used.`,
			Before: `if ($cond) {
  [$v, $err];
}`,
			After: `if ($cond) {
  return [$v, $err];
}`,
		},

		{
			Name:    "voidResultUsed",
			Default: true,
			Comment: `Report usages of the void-type expressions`,
			Before:  `$x = var_dump($v); // var_dump returns void`,
			After:   `$x = print_r($v, true);`,
		},

		{
			Name:    "keywordCase",
			Default: true,
			Comment: `Report keywords that are not in the lower case.`,
			Before:  `RETURN $x;`,
			After:   `return $x;`,
		},

		{
			Name:    "accessLevel",
			Default: true,
			Comment: `Report erroneous member access.`,
			Before:  `$x->_run(); // _run() is private, can't be accessed`,
			After:   `$x->run(); // run() is public`,
		},

		{
			Name:    "argCount",
			Default: true,
			Comment: `Report mismatching args count inside call expressions.`,
			Before:  `array_combine($keys)`,
			After:   `array_combine($keys, $values)`,
		},

		{
			Name:    "redundantGlobal",
			Default: true,
			Comment: `Report global statement over superglobal variables (which is redundant).`,
			Before:  `global $Foo, $_GET; // $_GET is superglobal`,
			After:   `global $Foo;`,
		},

		{
			Name:    "arrayAccess",
			Default: true,
			Comment: `Report array access to non-array objects.`,
			Before:  `return $foo[0]; // $foo value may not implement ArrayAccess`,
			After:   `if ($foo instanceof ArrayAccess) { return $foo[0]; }`,
		},

		{
			Name:    "bitwiseOps",
			Default: true,
			Comment: `Report suspicious usage of bitwise operations.`,
			Before:  `if ($isURL & $verify) ...`,
			After:   `if ($isURL && $veriry) ...`,
		},

		{
			Name:    "mixedArrayKeys",
			Default: true,
			Comment: `Report array literals that have both implicit and explicit keys.`,
			Before:  `['a', 5 => 'b']`,
			After:   `[0 => 'a', 5 => 'b']`,
		},

		{
			Name:    "dupGlobal",
			Default: true,
			Comment: `Report repeated global statements over variables.`,
			Before:  `global $x, $y, $x;`,
			After:   `global $x, $y;`,
		},

		{
			Name:    "dupArrayKeys",
			Default: true,
			Comment: `Report duplicated keys in array literals.`,
			Before:  `[A => 1, B => 2, A => 3]`,
			After:   `[A => 1, B => 2, C => 3]`,
		},

		{
			Name:    "dupCond",
			Default: true,
			Comment: `Report duplicated conditions in switch and if/else statements.`,
			Before: `if ($status == OK) {
  return "OK";
} elseif ($status == OK) { // Duplicated condition
  return "NOT OK";
} else {
  return "UNKNOWN";
}`,
			After: `if ($status == OK) {
  return "OK";
} elseif ($status == NOT_OK) {
  return "NOT OK";
} else {
  return "UNKNOWN";
}`,
		},

		{
			Name:    "dupBranchBody",
			Default: true,
			Comment: `Report suspicious conditional branches that execute the same action.`,
			Before:  `$pickLeft ? foo($left) : foo($left)`,
			After:   `$pickLeft ? foo($left) : foo($right)`,
		},

		{
			Name:    "dupSubExpr",
			Default: true,
			Comment: `Report suspicious duplicated operands in expressions.`,
			Before:  `return $x[$i] < $x[$i]; // $x[$i] is duplicated`,
			After:   `return $x[$i] < $x[$j];`,
		},

		{
			Name:    "arraySyntax",
			Default: true,
			Comment: `Report usages of old array() syntax.`,
			Before:  `array(1, 2)`,
			After:   `[1, 2]`,
		},

		{
			Name:    "bareTry",
			Default: true,
			Comment: `Report try blocks without catch/finally.`,
			Before: `try {
  doit();
}`,
			After: `try {
  doit();
} catch (Exception $e) {
  // Handle $e.
}`,
		},

		{
			Name:    "caseBreak",
			Default: true,
			Comment: `Report switch cases without break.`,
			Before: `switch ($v) {
case 1:
  echo "one"; // May want to insert a "break" here
case 2:
  echo "this fallthrough is intentional";
  // fallthrough
case 3:
  echo "two or three";
}`,
			After: `switch ($v) {
case 1:
  echo "one";
  break;
case 2:
  echo "this fallthrough is intentional";
  // fallthrough
case 3:
  echo "two or three";
}`,
		},

		{
			Name:    "complexity",
			Default: true,
			Comment: `Report funcs/methods that are too complex.`,
			Before: `function checkRights() {
  // Super big function.
}`,
			After: `function checkRights() {
  return true; // Or 42 if you need int-typed result
}`,
		},

		{
			Name:    "deadCode",
			Default: true,
			Comment: `Report potentially unreachable code.`,
			Before: `
thisFunctionExits(); // Always exits
foo(); // Dead code`,
			After: `
foo();
thisFunctionExits();`,
		},

		{
			Name:    "phpdocLint",
			Default: true,
			Comment: `Report malformed phpdoc comments.`,
			Before:  `@property $foo`,
			After:   `@property Foo $foo`,
		},

		{
			Name:    "phpdocType",
			Default: true,
			Comment: `Report potential issues in phpdoc types.`,
			Before:  `@var []int $xs`,
			After:   `@var int[] $xs`,
		},

		{
			Name:    "phpdocRef",
			Default: true,
			Comment: `Report invalid symbol references inside phpdoc.`,
			Before:  `@see MyClass`,
			After:   `@see \Foo\MyClass`,
		},

		{
			Name:    "phpdoc",
			Default: true,
			Comment: `Report missing phpdoc on public methods.`,
			Before: `public function process($acts, $config) {
  // Does something very complicated.
}`,
			After: `
/**
 * process executes all $acts in a new context.
 * Processed $acts should never be processed again.
 *
 * @param Act[] $acts - acts to execute
 * @param array $config - options
 */
public function process($acts, $config) {
  // Does something very complicated.
}`,
		},

		{
			Name:    "stdInterface",
			Default: true,
			Comment: `Report issues related to std PHP interfaces.`,
		},

		{
			Name:    "unimplemented",
			Default: true,
			Comment: `Report classes that don't implement their contract.`,
			Before: `class MyObj implements Serializable {
  public function serialize() { /* ... */ }
}`,
			After: `class MyObj implements Serializable {
  public function serialize() { /* ... */ }
  public function unserialize(string $s) { /* ... */ }
}`,
		},

		{
			Name:    "syntax",
			Default: true,
			Comment: `Report syntax errors.`,
			Before:  `foo(1]`,
			After:   `foo(1)`,
		},

		{
			Name:    "undefined",
			Default: true,
			Comment: `Report usages of potentially undefined symbols.`,
			Before: `if ($cond) {
  $v = 10;
}
return $v; // $v may be undefined`,
			After: `$v = 0; // Default value
if ($cond) {
  $v = 10;
}
return $v;`,
		},

		{
			Name:    "unused",
			Default: true,
			Comment: `Report potentially unused variables.`,
			Before: `$result = calculateResult(); // Unused $result
return [$err];`,
			After: `$result = calculateResult();
return [$result, $err];`,
		},

		{
			Name:    "redundantCast",
			Default: false,
			Comment: `Report redundant type casts.`,
			Before:  `return (int)10;`,
			After:   `return 10;`,
		},

		{
			Name:    "newAbstract",
			Default: true,
			Comment: `Report abstract classes usages in new expressions.`,
			Before:  `return new AbstractFactory();`,
			After:   `return new NonAbstractFactory();`,
		},

		{
			Name:    "regexpSimplify",
			Default: true,
			Comment: `Report regular expressions that can be simplified.`,
			Before:  `preg_match('/x(?:a|b|c){0,}/', $s)`,
			After:   `preg_match('/x[abc]*/', $s)`,
		},

		{
			Name:    "regexpVet",
			Default: true,
			Comment: `Report suspicious regexp patterns.`,
			Before:  `preg_match('a\d+a', $s); // 'a' is not a valid delimiter`,
			After:   `preg_match('/\d+/', $s);`,
		},

		{
			Name:    "regexpSyntax",
			Default: true,
			Comment: `Report regexp syntax errors.`,
		},

		{
			Name:    "caseContinue",
			Default: true,
			Comment: `Report suspicious 'continue' usages inside switch cases.`,
			Before: `switch ($v) {
case STOP:
  continue; // This acts like a "break"
case INC:
  $x++;
  break;
}`,
			After: `switch ($v) {
case STOP:
  break;
case INC:
  $x++;
  break;
}`,
		},

		{
			Name:    "deprecated",
			Default: false, // Experimental
			Comment: `Report usages of deprecated symbols.`,
			Before:  `ereg($pat, $s)`,
			After:   `preg_match($pat, $s)`,
		},

		{
			Name:    "callStatic",
			Default: true,
			Comment: `Report static calls of instance methods and vice versa.`,
			Before:  `$object::instance_method()`,
			After:   `$object->instance_method()`,
		},

		{
			Name:    "parentConstructor",
			Default: true,
			Comment: `Report missing parent::__construct calls in class constructors.`,
			Before: `class Foo extends Bar {
  public function __construct($x, $y) {
    $this->y = $y;
  }
}`,
			After: `class Foo extends Bar {
  public function __construct($x, $y) {
    parent::__construct($x);
    $this->y = $y;
  }
}`,
		},

		{
			Name:    "oldStyleConstructor",
			Default: true,
			Comment: `Report old-style (PHP4) class constructors.`,
			Before: `class Foo {
  public function Foo($v) { $this->v = $v; }
}`,
			After: `class Foo {
  public function __construct($v) { $this->v = $v; }
}`,
		},

		{
			Name:    "misspellName",
			Default: true,
			Comment: `Report commonly misspelled words in symbol names.`,
			Before:  `function performace_test() ...`, //nolint:misspell
			After:   `function performance_test() ...`,
		},

		//nolint:misspell
		{
			Name:    "misspellComment",
			Default: true,
			Comment: `Report commonly misspelled words in comments.`,
			Before: `/** This is our performace test. */
function performance_test() {}`,
			After: `/** This is our performance test. */
function performance_test() {}`,
		},

		{
			Name:    "nonPublicInterfaceMember",
			Default: true,
			Comment: `Report illegal non-public access level in interfaces.`,
			Before: `interface Iface {
  function a(); // OK, public implied
  public function b(); // OK, public
  private function c();
  protected function d();
}`,
			After: `interface Iface {
  function a(); // OK, public implied
  public function b(); // OK, public
  public function c();
  public function d();
}`,
		},

		{
			Name:    "linterError",
			Default: true,
			Comment: `Report linter error.`,
		},

		{
			Name:    "magicMethodDecl",
			Default: true,
			Comment: `Report issues in magic method declarations.`,
			Before: `class Foo {
  private function __call($method, $args) {} // The magic method __call() must have public visibility
  public static function __set($name, $value) {} // The magic method __set() cannot be static
}`,
			After: `class Foo {
  public function __call($method, $args) {} // Ok
  public function __set($name, $value) {} // Ok
}`,
		},

		{
			Name:    "nameCase",
			Default: true,
			Comment: `Report symbol case mismatches.`,
			Before: `class Foo {}
$foo = new foo();`,
			After: `class Foo {}
$foo = new Foo();`,
		},

		{
			Name:    "strictCmp",
			Default: true,
			Comment: `Report non-strict comparison with false/true/null.`,
			Before:  `$result == null`,
			After:   `$result === null`,
		},

		{
			Name:    "paramClobber",
			Default: true,
			Comment: `Report assignments that overwrite params prior to their usage.`,
			Before: `function api_get_video($user_id) {
  $user_id = 0;
  return get_video($user_id);
}`,
			After: `function api_get_video($user_id) {
  $user_id = $user_id ?: 0;
  return get_video($user_id);
}`,
		},
	}

	for _, info := range allChecks {
		DeclareCheck(info)
	}
}

// Report is a linter report message.
type Report struct {
	CheckName string `json:"check_name"`
	Level     int    `json:"level"`
	Context   string `json:"context"`
	Message   string `json:"message"`
	Filename  string `json:"filename"`
	Line      int    `json:"line"`
	StartChar int    `json:"start_char"`
	EndChar   int    `json:"end_char"`
	Hash      uint64 `json:"hash"`
}

var severityNames = map[int]string{
	LevelError:       "ERROR",
	LevelWarning:     "WARNING",
	LevelInformation: "INFO",
	LevelHint:        "HINT",
	LevelUnused:      "UNUSED",
	LevelDoNotReject: "MAYBE",
	LevelSyntax:      "SYNTAX",
	LevelSecurity:    "WARNING",
}

func (r *Report) Severity() string {
	return severityNames[r.Level]
}

// IsCritical returns whether or not we need to reject whole commit when found this kind of report.
func (r *Report) IsCritical() bool {
	return r.Level != LevelDoNotReject
}

// DiffReports returns only reports that are new.
// Pass diffArgs=nil if we are called from diff in working copy.
func DiffReports(gitRepo string, diffArgs []string, changesList []git.Change, changeLog []git.Commit, oldList, newList []*Report, maxConcurrency int) (res []*Report, err error) {
	ignoreCommits := make(map[string]struct{})
	for _, c := range changeLog {
		if strings.Contains(c.Message, IgnoreLinterMessage) {
			ignoreCommits[c.Hash] = struct{}{}
		}
	}

	old := reportListToMap(oldList)
	new := reportListToMap(newList)
	changes := gitChangesToMap(changesList)

	var mu sync.Mutex
	var wg sync.WaitGroup

	var resErr error

	limitCh := make(chan struct{}, maxConcurrency)

	for filename, list := range new {
		wg.Add(1)
		go func(filename string, list []*Report) {
			limitCh <- struct{}{}
			defer func() { <-limitCh }()
			defer wg.Done()

			var oldName string

			c, ok := changes[filename]
			if ok {
				oldName = c.OldName
			} else {
				oldName = filename // full diff mode
			}

			reports, err := diffReportsList(gitRepo, ignoreCommits, diffArgs, filename, c, old[oldName], list)
			if err != nil {
				mu.Lock()
				resErr = err
				mu.Unlock()
				return
			}

			mu.Lock()
			res = append(res, reports...)
			mu.Unlock()
		}(filename, list)
	}

	wg.Wait()

	if resErr != nil {
		return nil, err
	}

	return res, nil
}

type lineRangeChange struct {
	old, new git.LineRange
}

// compute blame only if refspec is not nil
func blameIfNeeded(gitDir string, refspec []string, filename string) (git.BlameResult, error) {
	if refspec == nil {
		return git.BlameResult{}, nil
	}

	return git.Blame(gitDir, refspec, filename)
}

func diffReportsList(gitRepo string, ignoreCommits map[string]struct{}, diffArgs []string, filename string, c git.Change, oldList, newList []*Report) (res []*Report, err error) {
	var blame git.BlameResult

	if c.Valid {
		blame, err = blameIfNeeded(gitRepo, diffArgs, filename)
		if err != nil {
			return nil, err
		}
	}

	changesMap := make(map[int]lineRangeChange, len(c.OldLineRanges))

	for idx, r := range c.OldLineRanges {
		for i := r.From; i <= r.To; i++ {
			changesMap[i] = lineRangeChange{old: r, new: c.LineRanges[idx]}
		}
	}

	old, oldMaxLine := reportListToPerLineMap(oldList)
	new, newMaxLine := reportListToPerLineMap(newList)

	var maxLine = oldMaxLine
	if newMaxLine > maxLine {
		maxLine = newMaxLine
	}

	var oldLine, newLine int

	for i := 0; i < maxLine; i++ {
		oldLine++
		newLine++

		ch, ok := changesMap[oldLine]
		// just deletion
		if ok && ch.new.HaveRange && ch.new.Range == 0 {
			oldLine = ch.old.To
			newLine-- // cancel the increment of newLine, because code was deleted, no new lines added
			continue
		}

		res = maybeAppendReports(res, new, old, newLine, oldLine, blame, ignoreCommits)

		if ok {
			oldLine = 0 // all changes and additions must be checked
			for j := newLine + 1; j <= ch.new.To; j++ {
				newLine = j
				res = maybeAppendReports(res, new, old, newLine, oldLine, blame, ignoreCommits)
			}
			oldLine = ch.old.To
		}
	}

	return res, nil
}

func maybeAppendReports(res []*Report, new, old map[int][]*Report, newLine, oldLine int, blame git.BlameResult, ignoreCommits map[string]struct{}) []*Report {
	newReports, ok := new[newLine]

	if !ok {
		return res
	}

	if _, ok := old[oldLine]; ok {
		return res
	}

	changedCommit := blame.Lines[newLine]

	if _, ok := ignoreCommits[changedCommit]; ok {
		return res
	}

	return append(res, newReports...)
}

func reportListToPerLineMap(list []*Report) (res map[int][]*Report, maxLine int) {
	res = make(map[int][]*Report)

	for _, l := range list {
		res[l.Line] = append(res[l.Line], l)
		if l.Line > maxLine {
			maxLine = l.Line
		}
	}

	return res, maxLine
}

func gitChangesToMap(changes []git.Change) map[string]git.Change {
	res := make(map[string]git.Change)
	for _, c := range changes {
		res[c.NewName] = c
	}
	return res
}

func reportListToMap(list []*Report) map[string][]*Report {
	res := make(map[string][]*Report)

	for _, r := range list {
		res[r.Filename] = append(res[r.Filename], r)
	}

	for _, l := range res {
		sort.Slice(l, func(i, j int) bool {
			return l[i].Line < l[j].Line
		})
	}

	return res
}

func isUnderscore(s string) bool {
	return s == "_"
}

func linterError(filename, format string, args ...interface{}) {
	log.Printf("error: "+filename+": "+format, args...)
}
