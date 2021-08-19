<?php

$options = [];
$s = 'efdfw';

preg_match('/clipFrom/([0-9]+)', $options['foo'], $match); // want `invalid modifier (`

preg_match('/foo/-i', $s); // want `invalid modifier -`

// Good modifiers.
preg_match('/foo/imsxADSUXJu', $s);
preg_match('/foo/imsx', $s);
preg_match('/foo/mi', $s);
