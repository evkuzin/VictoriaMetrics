package unittest

import (
	"testing"
	"time"
)

func TestEvalContext(t *testing.T) {
	f := func(timeout time.Duration, deadlineExpected bool) {
		t.Helper()

		oldTimeout := queryTimeout
		queryTimeout = timeout
		defer func() {
			queryTimeout = oldTimeout
		}()

		ctx, cancel := evalContext()
		defer cancel()
		deadline, ok := ctx.Deadline()
		if ok != deadlineExpected {
			t.Fatalf("unexpected deadline presence for timeout=%s; got %v; want %v", timeout, ok, deadlineExpected)
		}
		if !ok {
			return
		}
		if d := time.Until(deadline); d > timeout {
			t.Fatalf("deadline is too far for timeout=%s; got %s", timeout, d)
		}
	}

	// zero and negative timeouts keep the evaluation unbounded
	f(0, false)
	f(-time.Second, false)

	f(time.Second, true)
	f(time.Minute, true)
}

// TestUnitTest_QueryTimeoutReached makes sure the timeout is applied to the queries
// of both rule evaluation and metricsql_expr_test assertions. A timeout which can
// never be met must fail the test instead of being ignored.
//
// Every file holds a single kind of assertion, so that a timeout missing on one of
// the two query paths cannot be hidden by the other one.
func TestUnitTest_QueryTimeoutReached(t *testing.T) {
	f := func(file string) {
		t.Helper()

		if failed := UnitTest([]string{file}, false, nil, "", "", "", time.Nanosecond); !failed {
			t.Fatalf("expecting %q to fail with an unreachable -queryTimeout", file)
		}
		if failed := UnitTest([]string{file}, false, nil, "", "", "", 0); failed {
			t.Fatalf("unexpected failure of %q without a timeout", file)
		}
	}

	// rule evaluation queries the datasource for every rule of a group
	f("./testdata/timeout-alert.yaml")

	// metricsql_expr_test queries the datasource for every assertion
	f("./testdata/timeout-expr.yaml")
}
