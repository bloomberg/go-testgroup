// Copyright 2020 Bloomberg Finance L.P.

package testgroup_test

import (
	"testing"

	"github.com/bloomberg/go-testgroup"
)

func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// This function is the entry point to the test group.
// Because it starts with "Test" and accepts a *testing.T argument,
// it is detected and called by the Go testing package when you run "go test".
func TestAbsInt(t *testing.T) {
	testgroup.RunSerially(t, &AbsIntTests{})
}

type AbsIntTests struct {
	// Group-wide data/state can be stored here.
}

func (grp *AbsIntTests) DoesNotChangeNonNegativeNumbers(t *testgroup.T) {
	t.Equal(0, AbsInt(0))
	t.Equal(1, AbsInt(1))
	t.Equal(123456789, AbsInt(123456789))
}

func (grp *AbsIntTests) MakesNegativeNumbersPositive(t *testgroup.T) {
	t.Equal(1, AbsInt(-1))
	t.Equal(123456789, AbsInt(-123456789))
}
