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

type PreGrouper interface{ PreGroup(t *T) }
type PostGrouper interface{ PostGroup(t *T) }

type PreTester interface{ PreTest(t *T) }
type PostTester interface{ PostTest(t *T) }

var ParallelSeparator = "_"

func RunSerial(t *testing.T, group interface{}) {
	run(t, false, group)
}

func RunParallel(t *testing.T, group interface{}) {
	run(t, true, group)
}

func (t *T) RunSerial(group interface{}) {
	RunSerial(t.T, group)
}

func (t *T) RunParallel(group interface{}) {
	RunParallel(t.T, group)
}

func run(t *testing.T, parallel bool, group interface{}) {
	groupT := &T{
		T:          t,
		Assertions: assert.New(t),
		Require:    require.New(t),
	}

	testMethods := findTestMethods(t, group)
	if t.Failed() {
		return
	}

	if pg, ok := group.(PreGrouper); ok {
		pg.PreGroup(groupT)
	}

	if pg, ok := group.(PostGrouper); ok {
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

			if pt, ok := group.(PreTester); ok {
				pt.PreTest(methodT)
			}

			if pt, ok := group.(PostTester); ok {
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
	tests := []testMethod{}

	groupValue := reflect.ValueOf(group)
	groupType := groupValue.Type()

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

	return tests
}
