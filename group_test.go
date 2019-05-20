// Copyright 2019 Bloomberg Finance L.P.

package testgroup_test

import (
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
	s := SerialTests{calls: []string{}}
	testgroup.RunSerial(t, &s)

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

type SerialTests struct {
	calls []string
}

func (s *SerialTests) Called(t *testgroup.T, name string) {
	t.Logf("called %s", name)
	s.calls = append(s.calls, name)
}

func (s *SerialTests) ignoredNonExported(t *testgroup.T) { t.FailNow("should not happen") }
func (s *SerialTests) IgnoredExported(int)               { panic("should not happen") }

func (s *SerialTests) PreGroup(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreGroup", t.Name())) }
func (s *SerialTests) PostGroup(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostGroup", t.Name())) }

func (s *SerialTests) PreTest(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreTest", t.Name())) }
func (s *SerialTests) PostTest(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostTest", t.Name())) }

func (s *SerialTests) doTest(t *testgroup.T) { s.Called(t, t.Name()) }

// These methods are recognized as tests:
func (s *SerialTests) A(t *testgroup.T)    { s.doTest(t) }
func (s *SerialTests) B(t *testgroup.T)    { s.doTest(t) }
func (s *SerialTests) C(t *testgroup.T)    { s.doTest(t) }
func (s *SerialTests) Skip(t *testgroup.T) { t.SkipNow() }

//------------------------------------------------------------------------------

func Test_Parallel(t *testing.T) {
	s := ParallelTests{calls: []string{}}
	testgroup.RunParallel(t, &s)

	for i, call := range s.calls {
		t.Logf("s.calls[%2d]: = %v", i, call)
	}

	assert.Len(t, s.calls, 13)
	assert.Equal(t, fmt.Sprintf("%s PreGroup", t.Name()), s.calls[0])
	assert.Equal(t, fmt.Sprintf("%s PostGroup", t.Name()), s.calls[len(s.calls)-1])

	for _, name := range []string{"A", "B", "C", "Skip"} {
		var pre, test, post bool
		for _, call := range s.calls[1 : len(s.calls)-1] {
			if strings.HasPrefix(call, fmt.Sprintf("%s/%s/%s", t.Name(), testgroup.ParallelSeparator, name)) {
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

type ParallelTests struct {
	calls []string
	mutex sync.Mutex
}

func (s *ParallelTests) Called(t *testgroup.T, name string) {
	t.Logf("called %s", name)
	s.mutex.Lock()
	s.calls = append(s.calls, name)
	s.mutex.Unlock()
}

func (s *ParallelTests) PreGroup(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreGroup", t.Name())) }
func (s *ParallelTests) PostGroup(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostGroup", t.Name())) }

func (s *ParallelTests) PreTest(t *testgroup.T)  { s.Called(t, fmt.Sprintf("%s PreTest", t.Name())) }
func (s *ParallelTests) PostTest(t *testgroup.T) { s.Called(t, fmt.Sprintf("%s PostTest", t.Name())) }

func (s *ParallelTests) doTest(t *testgroup.T) { s.Called(t, t.Name()) }

// These methods are recognized as tests:
func (s *ParallelTests) A(t *testgroup.T)    { s.doTest(t) }
func (s *ParallelTests) B(t *testgroup.T)    { s.doTest(t) }
func (s *ParallelTests) C(t *testgroup.T)    { s.doTest(t) }
func (s *ParallelTests) Skip(t *testgroup.T) { t.SkipNow() }

//------------------------------------------------------------------------------

func Test_ThingsYouCanDoWithT(t *testing.T) {
	testgroup.RunSerial(t, &ThingsYouCanDoWithTTests{})
}

type ThingsYouCanDoWithTTests struct{}

func (g *ThingsYouCanDoWithTTests) Assert(t *testgroup.T) {
	t.Zero(*g)
	t.Equal(2, 1+1)
	t.Len("one", 3)
}

func (g *ThingsYouCanDoWithTTests) Require(t *testgroup.T) {
	var err error
	t.Require.NoError(err)
}

func (g *ThingsYouCanDoWithTTests) TestingT(t *testgroup.T) {
	if t.Failed() {
		t.Log("How did the test already fail?")
	}
}

type Subgroup struct {
	Count int32
}

func (sg *Subgroup) AddOne(t *testgroup.T) { atomic.AddInt32(&sg.Count, 1) }
func (sg *Subgroup) AddTwo(t *testgroup.T) { atomic.AddInt32(&sg.Count, 2) }

func (g *ThingsYouCanDoWithTTests) RunSubgroupInSerial(t *testgroup.T) {
	sg := Subgroup{}
	t.RunSerial(&sg)
	t.Equal(int32(3), sg.Count)
}

func (g *ThingsYouCanDoWithTTests) RunSubgroupInParallel(t *testgroup.T) {
	sg := Subgroup{}
	t.RunParallel(&sg)
	t.Equal(int32(3), sg.Count)
}

//------------------------------------------------------------------------------

func Test_Failures(t *testing.T) {
	ctx := context.Background()

	tests := []string{
		"BadReservedMethodSignature",
		"BadTestMethodSignature",
	}

	for _, testName := range tests {
		t.Run(testName, func(t *testing.T) {
			cmd := exec.CommandContext(ctx,
				"go", "test", "-tags=failures", "-run", "FailureGroups/"+testName)

			out, err := cmd.CombinedOutput()
			if err != nil && err.(*exec.ExitError).ExitCode() != 0 {
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
