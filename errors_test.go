// Copyright 2019 Bloomberg Finance L.P.

// +build testgroup_errors

package testgroup_test

import (
	"testing"

	"github.com/bloomberg/go-testgroup"
)

// These tests are run in their own processes by Test_Errors in testgroup_test.go.
// All tests should have the prefix "Test_Error_" to be found by Test_Errors.

//------------------------------------------------------------------------------

func Test_Error_ReservedMethodWithBadArg(t *testing.T) {
	testgroup.RunSerially(t, &ReservedMethodWithBadArgGroup{})
}

type ReservedMethodWithBadArgGroup struct{}

// This is a bad PreTest method since it doesn't accept *testgroup.T.
func (*ReservedMethodWithBadArgGroup) PreTest(t *testing.T) {}

//------------------------------------------------------------------------------

func Test_Error_TestMethodWithBadArg(t *testing.T) {
	testgroup.RunSerially(t, &TestMethodWithBadArgGroup{})
}

type TestMethodWithBadArgGroup struct{}

func (*TestMethodWithBadArgGroup) This_method_accepts_the_wrong_type_of_T(t *testing.T) {}

//------------------------------------------------------------------------------

func Test_Error_TestMethodWithoutArg(t *testing.T) {
	testgroup.RunSerially(t, &TestMethodWithoutArgGroup{})
}

type TestMethodWithoutArgGroup struct{}

func (*TestMethodWithoutArgGroup) This_method_is_missing_its_argument() {}

//------------------------------------------------------------------------------

func Test_Error_NoTestMethodsFound(t *testing.T) {
	testgroup.RunSerially(t, &GroupWithoutTestMethods{})
}

type GroupWithoutTestMethods struct{}

//------------------------------------------------------------------------------

func Test_Error_MixedReceiverMethods(t *testing.T) {
	// If a pointer-to-struct were passed as the argument, this would not fail.
	testgroup.RunSerially(t, GroupWithMixedReceiverMethods{})
}

type GroupWithMixedReceiverMethods struct{}

// Since it has a pointer type receiver, this method is not part of the struct's method set.
func (*GroupWithMixedReceiverMethods) PointerMethod(t *testgroup.T) {}

func (GroupWithMixedReceiverMethods) NonPointerMethod(t *testgroup.T) {}
