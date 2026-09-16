package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

// empty() type narrowing: when empty($x) is known to be false,
// the always-falsy type keys (null, false, empty_array) are erased
// from $x type. The true branch keeps the current type.
func TestEmptyTypeNarrowing(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php

class VersionHelper
{
    /** @return string|null */
    public static function getVersion(): ?string { return null; }

    /**
     * @param string $currentVersion
     * @param string $needVersion
     * @return bool
     */
    public static function isNewVersionAppMobile(string $currentVersion, string $needVersion): bool
    {
        return $currentVersion >= $needVersion;
    }
}

function emptyGuard(): bool
{
    $v = VersionHelper::getVersion();

    if (empty($v)) {
        return false;
    }

    return VersionHelper::isNewVersionAppMobile($v, '1.0.0');
}

function notEmptyIf(): bool
{
    $v = VersionHelper::getVersion();

    if (!empty($v)) {
        return VersionHelper::isNewVersionAppMobile($v, '1.0.0');
    }

    return false;
}

function notEmptyGuardWithElse(): bool
{
    $v = VersionHelper::getVersion();

    if (empty($v)) {
        return false;
    } else {
        return VersionHelper::isNewVersionAppMobile($v, '1.0.0');
    }
}
`)
}
