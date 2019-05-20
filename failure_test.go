// Copyright 2019 Bloomberg Finance L.P.

// +build failures

package testgroup_test

import (
	"testing"

	"github.com/bloomberg/go-testgroup"
)

// These tests are run by Test_Failures in group_test.go, so if you are adding a failure test,
// be sure to add it to the list in that function as well.

func Test_FailureGroups(t *testing.T) {
	testgroup.RunSerial(t, &FailureGroups{})
}

type FailureGroups struct{}

//------------------------------------------------------------------------------

func (*FailureGroups) BadReservedMethodSignature(t *testgroup.T) {
	t.RunSerial(&BadReservedMethodSignatureGroup{})
}

type BadReservedMethodSignatureGroup struct{}

// This is a bad PreTest method since it doesn't accept *testgroup.T.
func (*BadReservedMethodSignatureGroup) PreTest(t *testing.T) {}

//------------------------------------------------------------------------------

func (*FailureGroups) BadTestMethodSignature(t *testgroup.T) {
	t.RunSerial(&BadTestMethodSignatureGroup{})
}

type BadTestMethodSignatureGroup struct{}

func (*BadTestMethodSignatureGroup) Test_accepts_the_wrong_T_type(t *testing.T) {}
