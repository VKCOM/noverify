<?php

/**
 * @throws Exception
 */
function throwException() { throw new Exception; }

function finallyReturnBadReturnInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        return 2;
    } finally { // want `finally block contains a return`
        return 1;
    }
}

function finallyReturnBadThrowInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        throw new Exception();
    } finally { // want `finally block contains a return`
        return 1;
    }
}

function finallyReturnBadDieInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        die();
    } finally { // want `finally block contains a return`
        return 1;
    }
}

function finallyReturnBadMultiplyExitPointInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        if (0) {
            die();
        } else {
            return 2;
        }
    } finally { // want `finally block contains a return`
        return 1;
    }
}

function finallyReturnBadMultiplyExitPointInTry(): int {
    try {
        if (0) {
            throwException();
        } else {
            return 1;
        }
    } catch (Exception $_) {
    } finally { // want `finally block contains a return`
        return 1;
    }
}

function finallyReturnBadMultiplyCatch(): int {
    try {
        throwException();
    } catch (RuntimeException $_) {
        return 2;
    } catch (Exception $_) {
        die();
    } finally { // want `finally block contains a return`
        return 1;
    }
}

function finallyReturnOkWithoutReturnInTryCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
    } finally {
        return 1; /* ok, catch and try blocks don't contain return/exceptions/die/etc */
    }
}

function finallyReturnOkFinallyWithoutReturn(): int {
    try {
        throwException();
    } catch (Exception $_) {
        return 1;
    } finally {
        echo 1; /* ok, finally don't contain return */
    }
}

function finallyReturnOkFinallyWithConditionalReturn(): int {
    try {
        throwException();
    } catch (Exception $_) {
        return 1;
    } finally {
        if (2) {
            return 2;
        }
    }
}
