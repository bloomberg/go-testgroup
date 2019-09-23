// Copyright 2019 Bloomberg Finance L.P.

package testgroup

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type T struct {
	*testing.T
	*assert.Assertions
	Require *require.Assertions
}

var ParallelSeparator = "_"

func RunSerially(t *testing.T, group interface{}) {
	t.Helper()
	run(t, false, group)
}

func RunInParallel(t *testing.T, group interface{}) {
	t.Helper()
	run(t, true, group)
}

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

func (t *T) RunSerially(group interface{}) {
	t.T.Helper()
	RunSerially(t.T, group)
}

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
		t.Run(ParallelSeparator, func(t *testing.T) {
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
