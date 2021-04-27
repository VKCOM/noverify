<?php

/**
 * @throws Exception
 */
function throwException() { throw new Exception; }

function finallyReturnBadReturnInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        return 2; // want `return is unreachable (because finally block contains a return on line 14)`
    } finally {
        return 1;
    }
}

function finallyReturnBadThrowInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        throw new Exception(); // want `throw is unreachable (because finally block contains a return on line 24)`
    } finally {
        return 1;
    }
}

function finallyReturnBadDieInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        die();
    } finally { // want `block finally is unreachable (because catch block 1 contains a exit/die)`
        return 1;
    }
}

function finallyReturnBadMultiplyExitPointInCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
        if (1) {
            die();
        } else {
            return 2;
        }
    } finally { // want `block finally is unreachable (because catch block 1 contains a exit/die)`
        return 1;
    }
}

function finallyReturnBadMultiplyExitPointInTry(): int {
    try {
        if (0) {
            throwException();
        } else {
            return 1; // want `return is unreachable (because finally block contains a return on line 61)`
        }
    } catch (Exception $_) {
    } finally {
        return 1;
    }
}

function finallyReturnBadMultiplyCatch(): int {
    try {
        throwException();
    } catch (RuntimeException $_) {
        return 2; // want `return is unreachable (because finally block contains a return on line 73)`
    } catch (Exception $_) {
       return 3; // want `return is unreachable (because finally block contains a return on line 73)`
    } finally {
        return 1;
    }
}

function finallyReturnBadMultiplyCatchWithDie(): int {
    try {
        throwException();
    } catch (RuntimeException $_) {
        return 2;
    } catch (Exception $_) {
        die();
    } finally { // want `block finally is unreachable (because catch block 2 contains a exit/die)`
        return 1;
    }
}

function finallyReturnBadMultiplyCatchWithExit(): int {
    try {
        throwException();
    } catch (RuntimeException $_) {
        return 2;
    } catch (Exception $_) {
        exit();
    } finally { // want `block finally is unreachable (because catch block 2 contains a exit/die)`
        return 1;
    }
}

function finallyReturnOkWithoutReturnInTryCatch(): int {
    try {
        throwException();
    } catch (Exception $_) {
    } finally {
        return 1; // ok, catch and try blocks don't contain return/exceptions/die/etc
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
