// Copyright 2019 Bloomberg Finance L.P.

// Package testgroup helps you group tests together. A subtest of a group is simply an exported
// method of a type (usually a struct). Being part of a group allows tests to share state and common
// functionality, including pre/post-group and pre/post-test functions.
//
// testgroup was inspired by github.com/stretchr/testify/suite.
package testgroup

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T is a type passed to each test function. It is mainly concerned with test state, and it embeds
// and contains other types for convenience.
type T struct {
	*testing.T
	*assert.Assertions
	Require *require.Assertions
}

// RunInParallelParentTestName is the name of the parent test of RunInParallel subtests.
//
// For example, given a function TestParallel that calls RunInParallel on a test group struct with
// two subtests A and B, the test output might look like this:
//
//     $ go test -v
//     === RUN   TestParallel
//     === RUN   TestParallel/_
//     === RUN   TestParallel/_/A
//     === PAUSE TestParallel/_/A
//     === RUN   TestParallel/_/B
//     === PAUSE TestParallel/_/B
//     === CONT  TestParallel/_/A
//     === CONT  TestParallel/_/B
//     --- PASS: TestParallel (0.00s)
//         --- PASS: TestParallel/_ (0.00s)
//             --- PASS: TestParallel/_/A (0.00s)
//             --- PASS: TestParallel/_/B (0.00s)
//     PASS
//     ok  	example	0.013s
//
// You can change the value of RunInParallelParentTestName to replace "_" above with another string.
var RunInParallelParentTestName = "_"

// RunSerially runs the test methods of a group sequentially in lexicographic order.
func RunSerially(t *testing.T, group interface{}) {
	t.Helper()
	run(t, false, group)
}

// RunInParallel runs the test methods of a group simultaneously and waits for all of them to
// complete before returning.
func RunInParallel(t *testing.T, group interface{}) {
	t.Helper()
	run(t, true, group)
}

// Run is just like testing.T.Run, but the argument to f is a *testgroup.T instead of a *testing.T.
func (t *T) Run(name string, f func(t *T)) {
	t.T.Helper()
	t.T.Run(name, func(t *testing.T) {
		funcT := &T{
			T:          t,
			Assertions: assert.New(t),
			Require:    require.New(t),
		}

		f(funcT)
	})
}

// RunSerially runs the test methods of a group sequentially in lexicographic order.
func (t *T) RunSerially(group interface{}) {
	t.T.Helper()
	RunSerially(t.T, group)
}

// RunInParallel runs the test methods of a group simultaneously and waits for all of them to
// complete before returning.
func (t *T) RunInParallel(group interface{}) {
	t.T.Helper()
	RunInParallel(t.T, group)
}

func run(t *testing.T, parallel bool, group interface{}) {
	t.Helper()

	groupT := &T{
		T:          t,
		Assertions: assert.New(t),
		Require:    require.New(t),
	}

	testMethods := findTestMethods(t, group)
	if len(testMethods) == 0 {
		t.Fatalf(
			"testgroup: no tests found for %T."+
				" Make sure your test methods are exported and that their receiver types"+
				" match what you passed to testgroup.",
			group)
	}

	type preGrouper interface{ PreGroup(t *T) }
	if pg, ok := group.(preGrouper); ok {
		pg.PreGroup(groupT)
	}

	type postGrouper interface{ PostGroup(t *T) }
	if pg, ok := group.(postGrouper); ok {
		defer pg.PostGroup(groupT)
	}

	if parallel {
		// wrap in a t.Run to wait for the parallel tests to finish
		t.Run(RunInParallelParentTestName, func(t *testing.T) {
			runAllTests(t, parallel, group, testMethods)
		})
	} else {
		runAllTests(t, parallel, group, testMethods)
	}
}

func runAllTests(t *testing.T, parallel bool, group interface{}, testMethods []testMethod) {
	t.Helper()

	for _, m := range testMethods {
		t.Run(m.Name, func(t *testing.T) {
			method := m // local copy inside closure
			if parallel {
				t.Parallel()
			}

			methodT := &T{
				T:          t,
				Assertions: assert.New(t),
				Require:    require.New(t),
			}

			type preTester interface{ PreTest(t *T) }
			if pt, ok := group.(preTester); ok {
				pt.PreTest(methodT)
			}

			type postTester interface{ PostTest(t *T) }
			if pt, ok := group.(postTester); ok {
				defer pt.PostTest(methodT)
			}

			method.Method.Call([]reflect.Value{reflect.ValueOf(methodT)})
		})
	}
}

//------------------------------------------------------------------------------

type testMethod struct {
	Name   string
	Method reflect.Value
}

func findTestMethods(t *testing.T, group interface{}) []testMethod {
	t.Helper()

	tests := []testMethod{}

	groupValue := reflect.ValueOf(group)
	groupType := groupValue.Type()

	if groupType.Kind() != reflect.Ptr {
		ptrType := reflect.PtrTo(groupType)
		if ptrType != nil && ptrType.NumMethod() != groupType.NumMethod() {
			t.Fatalf(
				"testgroup: mixed method receivers: %v has %v methods, but %v has %v methods."+
					" You should either pass a pointer or make the extra methods private.",
				groupType, groupType.NumMethod(),
				ptrType, ptrType.NumMethod(),
			)
		}
	}

	expectedTestSignature := reflect.TypeOf(func(*T) {})

	testingTSignature := reflect.TypeOf(func(*testing.T) {})

	reservedMethodNames := map[string]bool{
		"PreGroup":  true,
		"PostGroup": true,
		"PreTest":   true,
		"PostTest":  true,
	}

	for i := 0; i < groupType.NumMethod(); i++ {
		method := groupType.Method(i)
		methodShortName := method.Name
		methodFullName := fmt.Sprintf("%v.%v", groupType, method.Name)

		methodValue := groupValue.Method(i)
		methodSignature := methodValue.Type()

		if reservedMethodNames[methodShortName] {
			// Reserved methods should also conform to the expectedTestSignature.
			if methodSignature != expectedTestSignature {
				t.Errorf(
					"testgroup: %v is a reserved method but does not have type %v",
					methodFullName, expectedTestSignature)
			}
			continue
		}

		switch methodSignature {
		case expectedTestSignature:
			break
		case testingTSignature:
			t.Errorf("testgroup: %v accepts *testing.T, not *testgroup.T", methodFullName)
			continue
		default:
			continue
		}

		tests = append(tests, testMethod{
			Name:   methodShortName,
			Method: methodValue,
		})
	}

	if t.Failed() {
		t.Fatal("testgroup: problems finding test methods -- see previous failures")
	}

	return tests
}
