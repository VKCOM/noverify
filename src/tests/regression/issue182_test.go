package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue182(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
trait SingletonSelf {
    /** @var self */
    private static $instance = null;

    /** @return self */
    public static function instance() {
        if (self::$instance === null) {
            self::$instance = new self();
        }

        return self::$instance;
    }
}

trait SingletonStatic {
    /** @var static */
    private static $instance = null;

    /** @return static */
    public static function instance() {
        if (static::$instance === null) {
            static::$instance = new static();
        }

        return static::$instance;
    }
}
`)
}
