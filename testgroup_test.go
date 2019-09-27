// Copyright 2019 Bloomberg Finance L.P.

package testgroup_test

import (
	"bufio"
	"bytes"
	"context"
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
	s := Serial{calls: []string{}}
	testgroup.RunSerially(t, &s)

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
		s.calls)
}

type Serial struct {
	calls []string
}

func (s *Serial) Called(t *testgroup.T, name string) {
	t.Logf("called %s", name)
	s.calls = append(s.calls, name)
}

func (s *Serial) ignoredNonExported(t *testgroup.T) { t.FailNow("should not happen") }
func (s *Serial) IgnoredExported(int)               { panic("should not happen") }

func (s *Serial) PreGroup(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreGroup", t.Name())) }
func (s *Serial) PostGroup(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostGroup", t.Name())) }

func (s *Serial) PreTest(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreTest", t.Name())) }
func (s *Serial) PostTest(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostTest", t.Name())) }

func (s *Serial) doTest(t *testgroup.T) { s.Called(t, t.Name()) }

// These methods are recognized as tests:
func (s *Serial) A(t *testgroup.T)    { s.doTest(t) }
func (s *Serial) B(t *testgroup.T)    { s.doTest(t) }
func (s *Serial) C(t *testgroup.T)    { s.doTest(t) }
func (s *Serial) Skip(t *testgroup.T) { t.SkipNow() }

//------------------------------------------------------------------------------

func Test_Parallel(t *testing.T) {
	s := Parallel{calls: []string{}}
	testgroup.RunInParallel(t, &s)

	for i, call := range s.calls {
		t.Logf("s.calls[%2d]: = %v", i, call)
	}

	assert.Len(t, s.calls, 13)
	assert.Equal(t, fmt.Sprintf("%s PreGroup", t.Name()), s.calls[0])
	assert.Equal(t, fmt.Sprintf("%s PostGroup", t.Name()), s.calls[len(s.calls)-1])

	for _, name := range []string{"A", "B", "C", "Skip"} {
		prefix := fmt.Sprintf("%s/%s/%s", t.Name(), testgroup.RunInParallelParentTestName, name)
		var pre, test, post bool
		for _, call := range s.calls[1 : len(s.calls)-1] {
			if strings.HasPrefix(call, prefix) {
				switch {
				case strings.HasSuffix(call, "PreTest"):
					assert.False(t, pre)
					assert.False(t, test)
					assert.False(t, post)
					pre = true
				case strings.HasSuffix(call, "PostTest"):
					assert.True(t, pre)
					assert.Equal(t, name != "Skip", test)
					assert.False(t, post)
					post = true
				default:
					assert.NotEqual(t, "Skip", name)
					assert.True(t, pre)
					assert.False(t, test)
					assert.False(t, post)
					test = true
				}
			}
		}
		assert.True(t, pre)
		assert.Equal(t, name != "Skip", test)
		assert.True(t, post)
	}
}

type Parallel struct {
	calls []string
	mutex sync.Mutex
}

func (s *Parallel) Called(t *testgroup.T, name string) {
	t.Logf("called %s", name)
	s.mutex.Lock()
	s.calls = append(s.calls, name)
	s.mutex.Unlock()
}

func (s *Parallel) PreGroup(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreGroup", t.Name())) }
func (s *Parallel) PostGroup(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostGroup", t.Name())) }

func (s *Parallel) PreTest(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreTest", t.Name())) }
func (s *Parallel) PostTest(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostTest", t.Name())) }

func (s *Parallel) doTest(t *testgroup.T) { s.Called(t, t.Name()) }

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
	var err error
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
	sg := Subgroup{}
	t.RunSerially(&sg)
	t.Equal(int32(3), sg.Count)
}

func (g *ThingsYouCanDoWithT) RunSubgroupInParallel(t *testgroup.T) {
	sg := Subgroup{}
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
		t.Run(testName, func(t *testing.T) {
			ctx := context.Background()

			cmd := exec.CommandContext(ctx,
				"go", "test",
				"-tags", "testgroup_errors",
				"-run", "^"+testName+"$",
			)

			t.Logf("cmd.Args: %v", cmd.Args)

			out, err := cmd.CombinedOutput()
			if err != nil && err.(*exec.ExitError).ExitCode() == 1 {
				// It failed, as expected.
				return
			}

			t.Logf("expected test to fail!")
			t.Logf("cmd.Args: %#v", cmd.Args)
			t.Logf("err: %v", err)
			t.Logf("combined output:\n%s", out)
			t.FailNow()
		})
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
		return nil, fmt.Errorf("err: %q combined output:\n%s", err, out)
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
		return nil, err
	}

	return tests, nil
}
