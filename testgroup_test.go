// Copyright 2019 Bloomberg Finance L.P.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testgroup_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bloomberg/go-testgroup"
	"github.com/stretchr/testify/assert"
)

//------------------------------------------------------------------------------

func Test_Serial(t *testing.T) {
	tests := Serial{calls: []string{}}
	testgroup.RunSerially(t, &tests)

	callName := func(c string) string {
		return fmt.Sprintf("%s/%s", t.Name(), c)
	}

	assert.Equal(
		t,
		[]string{
			fmt.Sprintf("%s PreGroup", t.Name()),
			callName("A PreTest"),
			callName("A"),
			callName("A PostTest"),
			callName("B PreTest"),
			callName("B"),
			callName("B PostTest"),
			callName("C PreTest"),
			callName("C"),
			callName("C PostTest"),
			callName("Skip PreTest"),
			callName("Skip PostTest"),
			fmt.Sprintf("%s PostGroup", t.Name()),
		},
		tests.calls)
}

type Serial struct {
	calls []string
}

func (s *Serial) called(t *testgroup.T, name string) {
	t.Logf("called %s", name)
	s.calls = append(s.calls, name)
}

//nolint:unused // This function being unused is itself a test.
func (s *Serial) ignoredNonExported(t *testgroup.T) { t.FailNow("should not happen") }

func (s *Serial) PreGroup(t *testgroup.T)  { s.called(t, fmt.Sprintf("%s PreGroup", t.Name())) }
func (s *Serial) PostGroup(t *testgroup.T) { s.called(t, fmt.Sprintf("%s PostGroup", t.Name())) }

func (s *Serial) PreTest(t *testgroup.T)  { s.called(t, fmt.Sprintf("%s PreTest", t.Name())) }
func (s *Serial) PostTest(t *testgroup.T) { s.called(t, fmt.Sprintf("%s PostTest", t.Name())) }

func (s *Serial) doTest(t *testgroup.T) { s.called(t, t.Name()) }

// These methods are recognized as tests:

func (s *Serial) A(t *testgroup.T)    { s.doTest(t) }
func (s *Serial) B(t *testgroup.T)    { s.doTest(t) }
func (s *Serial) C(t *testgroup.T)    { s.doTest(t) }
func (s *Serial) Skip(t *testgroup.T) { t.SkipNow() }

//------------------------------------------------------------------------------

func Test_Parallel(t *testing.T) {
	tests := Parallel{calls: []string{}, mutex: sync.Mutex{}}
	testgroup.RunInParallel(t, &tests)

	for i, call := range tests.calls {
		t.Logf("tests.calls[%2d]: = %v", i, call)
	}

	assert.Len(t, tests.calls, 13)
	assert.Equal(t, fmt.Sprintf("%s PreGroup", t.Name()), tests.calls[0])
	assert.Equal(t, fmt.Sprintf("%s PostGroup", t.Name()), tests.calls[len(tests.calls)-1])

	for _, name := range []string{"A", "B", "C", "Skip"} {
		validateCallOrderForFunc(t, &tests, name)
	}
}

func validateCallOrderForFunc(t *testing.T, tests *Parallel, funcName string) {
	t.Helper()

	prefix := fmt.Sprintf("%s/%s/%s", t.Name(), testgroup.RunInParallelParentTestName, funcName)
	pre := false
	test := false
	post := false

	for _, call := range tests.calls[1 : len(tests.calls)-1] {
		if strings.HasPrefix(call, prefix) {
			switch {
			case strings.HasSuffix(call, "PreTest"):
				assert.False(t, pre)
				assert.False(t, test)
				assert.False(t, post)

				pre = true
			case strings.HasSuffix(call, "PostTest"):
				assert.True(t, pre)
				assert.Equal(t, funcName != "Skip", test)
				assert.False(t, post)

				post = true
			default:
				assert.NotEqual(t, "Skip", funcName)
				assert.True(t, pre)
				assert.False(t, test)
				assert.False(t, post)

				test = true
			}
		}
	}

	assert.True(t, pre)
	assert.Equal(t, funcName != "Skip", test)
	assert.True(t, post)
}

type Parallel struct {
	calls []string
	mutex sync.Mutex
}

func (s *Parallel) called(t *testgroup.T, name string) {
	t.Logf("called %s", name)
	s.mutex.Lock()
	s.calls = append(s.calls, name)
	s.mutex.Unlock()
}

func (s *Parallel) PreGroup(t *testgroup.T)  { s.called(t, fmt.Sprintf("%s PreGroup", t.Name())) }
func (s *Parallel) PostGroup(t *testgroup.T) { s.called(t, fmt.Sprintf("%s PostGroup", t.Name())) }

func (s *Parallel) PreTest(t *testgroup.T)  { s.called(t, fmt.Sprintf("%s PreTest", t.Name())) }
func (s *Parallel) PostTest(t *testgroup.T) { s.called(t, fmt.Sprintf("%s PostTest", t.Name())) }

func (s *Parallel) doTest(t *testgroup.T) { s.called(t, t.Name()) }

// These methods are recognized as tests:

func (s *Parallel) A(t *testgroup.T)    { s.doTest(t) }
func (s *Parallel) B(t *testgroup.T)    { s.doTest(t) }
func (s *Parallel) C(t *testgroup.T)    { s.doTest(t) }
func (s *Parallel) Skip(t *testgroup.T) { t.SkipNow() }

//------------------------------------------------------------------------------

func Test_ThingsYouCanDoWithT(t *testing.T) {
	testgroup.RunSerially(t, &ThingsYouCanDoWithT{})
}

type ThingsYouCanDoWithT struct{}

func (g *ThingsYouCanDoWithT) Assert(t *testgroup.T) {
	t.Zero(*g)
	t.Equal(2, 1+1)
	t.Len("one", 3)
}

func (g *ThingsYouCanDoWithT) Require(t *testgroup.T) {
	var b strings.Builder
	_, err := fmt.Fprintf(&b, "The answer is %d.", 42)
	t.Require.NoError(err)
}

func (g *ThingsYouCanDoWithT) UseTestingT(t *testgroup.T) {
	if t.Failed() {
		t.Log("How did the test already fail?")
	}
}

func (g *ThingsYouCanDoWithT) RunSubtests(t *testgroup.T) {
	positiveNumbers := []int{1, 3, 7, 42}
	for _, n := range positiveNumbers {
		num := n
		t.Run(fmt.Sprintf("%d", num), func(t *testgroup.T) {
			t.Greater(num, 0)
		})
	}
}

type Subgroup struct {
	Count int32
}

func (sg *Subgroup) AddOne(t *testgroup.T) { atomic.AddInt32(&sg.Count, 1) }
func (sg *Subgroup) AddTwo(t *testgroup.T) { atomic.AddInt32(&sg.Count, 2) }

func (g *ThingsYouCanDoWithT) RunSubgroupInSerial(t *testgroup.T) {
	sg := Subgroup{Count: 0}
	t.RunSerially(&sg)
	t.Equal(int32(3), sg.Count)
}

func (g *ThingsYouCanDoWithT) RunSubgroupInParallel(t *testgroup.T) {
	sg := Subgroup{Count: 0}
	t.RunInParallel(&sg)
	t.Equal(int32(3), sg.Count)
}

//------------------------------------------------------------------------------

// The go testing package doesn't include support for asserting that a particular test failed,
// so we run "go test" in a subprocess to confirm that a particular test reports an error.
//
// The erroring tests are located in a separate file guarded by a build tag so that they aren't part
// of the regular set of tests.
func Test_Errors(t *testing.T) {
	tests, err := findErrorTests()
	if err != nil {
		t.Fatalf("failed to find error tests: %v", err)
	}

	for _, tn := range tests {
		testName := tn
		t.Run(testName, func(t *testing.T) { runTestExpectingFailure(t, testName) })
	}
}

func findErrorTests() ([]string, error) {
	const prefix = "Test_Error_"

	ctx := context.Background()

	cmd := exec.CommandContext(ctx,
		"go", "test",
		"-tags", "testgroup_errors",
		"-list", "^"+prefix)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: combined output:\n%s", err, out)
	}

	tests := []string{}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			tests = append(tests, strings.TrimSpace(line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning for error tests: %w", err)
	}

	return tests, nil
}

func runTestExpectingFailure(t *testing.T, testName string) {
	t.Helper()

	ctx := context.Background()

	//nolint:gosec // no risk using the testName param
	cmd := exec.CommandContext(ctx,
		"go", "test",
		"-tags", "testgroup_errors",
		"-run", fmt.Sprintf("^%s$", testName),
	)

	if raceDetectorEnabled {
		cmd.Args = append(cmd.Args, "-race")
	}

	cmd.Args = append(cmd.Args, goTestCoverageArgs(t.Name())...)

	t.Logf("cmd.Args: %v", cmd.Args)

	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if err != nil && errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		// It failed, as expected.
		return
	}

	t.Logf("expected test to fail!")
	t.Logf("cmd.Args: %#v", cmd.Args)
	t.Logf("err: %v", err)
	t.Logf("combined output:\n%s", out)
	t.FailNow()
}
