package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue387(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function f1(&$a) {
    $a1[0] = 1;
}

// TODO: report that $result is unchanged even if we tried to
// modify its arguments through references? See #388
function f2($result) {
    foreach ($result as &$file) {
        $file['filesystem'] = 'ntfs';
    }

    // Should be reported by #388.
    $arr = [];
    $arr['x'] = 10;
}
`)
}
