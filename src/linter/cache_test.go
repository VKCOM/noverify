package linter

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/google/go-cmp/cmp"
	"gotest.tools/assert"
)

func TestCache(t *testing.T) {
	go MemoryLimiterThread()

	// If this test is failing, you haven't broken anything (unless the decoding is failing),
	// but meta cache probably needs to be invalidated.
	//
	// What to do in order to fix it:
	//	1. Bump cacheVersion inside "cache.go" (please document what changed)
	//	2. Re-run test again. You might get different values in "have" lines.
	//	3. Copy "have" output to "want" variables in this test.

	code := `<?php
const GLOBAL_CONST = 1;

interface Arrayable {
  public function toArray();
}

final class Consts {
  const C1 = GLOBAL_CONST;
  const C2 = 'a';
  const C3 = ['a'];
}

abstract class AbstractClass {
  abstract public function f1();
}

class Point implements Arrayable {
  public $x = 0.0;
  public $y = 0.0;

  public function toArray() { return [$this->x, $this->y]; }
}

class Point3D extends Point implements Arrayable {
  public $z = 0.0;

  public function toArray() { return [$this->x, $this->y, $this->z]; }
}

function to_array(Arrayable $x) {
  return $x->toArray();
}

/**
 * @param int $a
 * @param int $b
 */
function add($a, $b) { return $a + $b; }

function main() {
  $p = new Point();
  $p->x = 1.5;
  $p->y = 3.3;
  $_ = $p->toArray();
  var_dump(to_array($p));

  $p3 = new Point3D();
  $_ = $p3->toArray();
  var_dump(to_array($p3));
}

/** @param shape(a:integer,b:shape(c:real)) */
function as_shape($x) {
  return $x;
}

main();
`

	runTest := func(iteration int) {
		_, root, err := ParseContents("cachetest.php", []byte(code), nil)
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		var buf bytes.Buffer
		wr := bufio.NewWriter(&buf)
		if err := writeMetaCache(wr, root); err != nil {
			t.Fatalf("write cache: %v", err)
		}
		wr.Flush()

		// We can't test for cache bytes, since gob encoding of maps is
		// not deterministic and we'll get unwanted diffs because of that.
		//
		// But we still can get make some checks that catch at least
		// some cache changes (that should cause version bump).

		// 1. Check cache contents length.
		//
		// If cache encoding changes, there is a very high chance that
		// encoded data lengh will change as well.
		wantLen := 3808
		haveLen := buf.Len()
		if haveLen != wantLen {
			t.Errorf("cache len mismatch:\nhave: %d\nwant: %d", haveLen, wantLen)
		}

		// 2. Check cache "strings" hash.
		//
		// It catches new fields in cached types, field renames and encoding of additional named attributes.
		wantStrings := "b752a530950d476bc523e58746fe52e13bd5277630c61f9cd3c88737dd1bff461083460dc4c157fc15cd56b9fc59711c943b7691cb1baec766cd4f8143447b0d"
		haveStrings := collectCacheStrings(buf.String())
		if haveStrings != wantStrings {
			t.Errorf("cache strings mismatch:\nhave: %q\nwant: %q", haveStrings, wantStrings)
		}

		// 3. Check meta decoding.
		//
		// If it fails, encoding and/or decoding is broken.
		encodedMeta := &root.meta
		decodedMeta := &fileMeta{}
		if err := readMetaCache(bytes.NewReader(buf.Bytes()), "", decodedMeta); err != nil {
			t.Errorf("decoding failed: %v", err)
		} else {
			// TODO: due to lots of important unexported fields,
			// we can't reliably check 2 meta files via assert.
			assert.DeepEqual(t,
				encodedMeta, decodedMeta,
				cmp.AllowUnexported(meta.TypesMap{}),
				cmp.AllowUnexported(meta.Scope{}))
		}

		if t.Failed() {
			t.Logf("cache contents:\n%q", buf.String())
			t.Fatalf("failed on iteration number %d", iteration)
		}
	}

	for i := 0; i < 20; i++ {
		runTest(i)
	}
}

func collectCacheStrings(data string) string {
	re := regexp.MustCompile(`[a-zA-Z_]\w*`)
	parts := re.FindAllString(data, -1)
	sort.Strings(parts)

	enc := sha512.New()
	enc.Write([]byte(strings.Join(parts, ",")))
	return hex.EncodeToString(enc.Sum(nil))
}
