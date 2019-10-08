# testgroup

`testgroup` helps you group tests together. A subtest of a group is simply an
exported method of a type (usually a `struct`). Being part of a group allows
tests to share state and common functionality, including pre/post-group and
pre/post-test functions.

`testgroup` was inspired by [testify][]'s [suite][testify-suite-godoc] package.

## Why make a new framework?

`testify/suite` is great, but it doesn't do well with tests that run in
parallel. Testify stores each test's `*testing.T` inside the `Suite` struct,
which means that only one `*testing.T` is available at a given time. If you run
tests in parallel, failures can be reported as another test failing, and if a
test failure is reported twice, `testing` panics.

After separating the group-wide state from the test-specific state, it was easy
to make a few usability improvements to the `*testing.T`-like object.

## Example

Here's what a simple test group looks like:

```go
package main

import (
    "testing"

    "github.com/bloomberg/go-testgroup"
)

// This entrypoint function is detected and called by the go test framework.
func TestMyGroup(t *testing.T) {
    testgroup.RunSerially(t, &MyGroup{})
}

type MyGroup struct{
    // Group-wide data/state can be stored here.
}

func (g *MyGroup) FirstTest(t *testgroup.T) {
    // test code
}

func (g *MyGroup) SecondTest(t *testgroup.T) {
    // test code
}
```

## Finding subtests

A group's subtests are the exported methods of your group object. Each subtest
must accept a `*testgroup.T` as its only argument and return nothing. You do
_not_ need to start your subtest methods with the word `Test`.

## Running subtests

`testgroup` has two functions to run subtests in a group: `RunSerially` and
`RunInParallel`.

### RunSerially

`RunSerially` runs a group's subtests in lexicographical order, one after
another.

Here's a contrived example:

```go
func TestSerial(t *testing.T) {
	testgroup.RunSerially(t, &MyGroup{})
}

type MyGroup struct{}

func (*MyGroup) C(t *testgroup.T) {}
func (*MyGroup) A(t *testgroup.T) {}
func (*MyGroup) B(t *testgroup.T) {}
```

Running the test above gives this output:

```console
$ go test -v serial_test.go
=== RUN   TestSerial
=== RUN   TestSerial/A
=== RUN   TestSerial/B
=== RUN   TestSerial/C
--- PASS: TestSerial (0.00s)
    --- PASS: TestSerial/A (0.00s)
    --- PASS: TestSerial/B (0.00s)
    --- PASS: TestSerial/C (0.00s)
PASS
ok  	command-line-arguments	0.014s
```

### RunInParallel

`RunInParallel` runs a group's subtests in parallel. It wraps a parent test
around your subtests to ensure that the `PreGroup`/`PostGroup` hooks run at the
correct time. You do not need to call `t.Parallel()` inside your tests &ndash;
`testgroup` does this for you.

By default, the parent test is named `_`, but you can override this by setting
`RunInParallelParentTestName`.

Here's a similar contrived example that uses `RunInParallel` instead of
`RunSerially`:

```go
func TestParallel(t *testing.T) {
	testgroup.RunInParallel(t, &MyGroup{})
}

type MyGroup struct{}

func (*MyGroup) C(t *testgroup.T) {}
func (*MyGroup) A(t *testgroup.T) {}
func (*MyGroup) B(t *testgroup.T) {}
```

Running the test above gives this output:

```console
$ go test -v parallel_test.go
=== RUN   TestParallel
=== RUN   TestParallel/_
=== RUN   TestParallel/_/A
=== PAUSE TestParallel/_/A
=== RUN   TestParallel/_/B
=== PAUSE TestParallel/_/B
=== RUN   TestParallel/_/C
=== PAUSE TestParallel/_/C
=== CONT  TestParallel/_/A
=== CONT  TestParallel/_/C
=== CONT  TestParallel/_/B
--- PASS: TestParallel (0.00s)
    --- PASS: TestParallel/_ (0.00s)
        --- PASS: TestParallel/_/B (0.00s)
        --- PASS: TestParallel/_/C (0.00s)
        --- PASS: TestParallel/_/A (0.00s)
PASS
ok  	command-line-arguments	0.014s
```

## Hooks

`testgroup` looks for the following specially-named hook methods:

```go
type MyGroup struct{}

func (*MyGroup) PreGroup(t *testgroup.T)  {} // runs before MyGroup's subtests have started
func (*MyGroup) PostGroup(t *testgroup.T) {} // runs after MyGroup's subtests have finished

func (*MyGroup) PreTest(t *testgroup.T)  {} // runs before each subtest in MyGroup
func (*MyGroup) PostTest(t *testgroup.T) {} // runs after each subtest in MyGroup
```

Like subtests, these methods accept a single `*testgroup.T` argument.

## Using `testgroup.T`

`testgroup.T` contains a number of useful members.

- It embeds a `*testing.T` struct, which lets you can write

  ```go
  func (*MyGroup) MySubtest(t *testgroup.T) {
      if testing.Short() {
          t.Skip("skipping due to short mode")
      }

      // set up a table of testcases ...

      for _, row := range table {
          t.Run(name, func (t *testing.T) {
              // test code
          })
      }
  }
  ```

- It embeds a [`testify/assert`][testify-assert-godoc] `*Assertions` struct,
  which lets you write

  ```go
  func (*MyGroup) MySubtest(t *testgroup.T) {
      result := callSomeFunction(input)
      t.Equal(expectedValue, result)
  }
  ```

- It contains a [`testify/require`][testify-require-godoc] `*Assertions` struct
  named `Require`, which lets you write

  ```go
  func (*MyGroup) MySubtest(t *testgroup.T) {
      result, err := somethingThatMightError(input)
      t.Require.NoError(err)
      t.Require.NotNil(result)
      t.Equal(expectedValue, result)
  }
  ```

## Notes

If you skip a test by calling `t.Skip()`, both the `PreTest` and `PostTest`
hooks will still run for that test.

## Acknowledgements

[Testify][] is copyright &copy; 2012-2018 Mat Ryer and Tyler Bunnell.

[testify]: https://github.com/stretchr/testify
[testify-assert-godoc]: https://godoc.org/github.com/stretchr/testify/assert
[testify-require-godoc]: https://godoc.org/github.com/stretchr/testify/require
[testify-suite-godoc]: https://godoc.org/github.com/stretchr/testify/suite
