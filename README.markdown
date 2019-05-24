# testgroup

`testgroup` is a test grouping framework inspired by and based on [testify][]'s
[suite][testify-suite-godoc] package.

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

`testgroup` considers _all_ of your group's exported methods as possible
subtests. You do _not_ need to start your subtests with the word `Test`. A valid
subtest accepts a `*testgroup.T` as its only argument.

## Run tests serially or in parallel

`testgroup` has two functions to run subtests in a group: `RunSerially` and
`RunInParallel`.

`RunSerially()` runs a group's subtests in alphabetical order, one after
another.

`RunInParallel()` runs a group's subtests in parallel. It adds a test wrapper
around your subtests to ensure that the `PreGroup`/`PostGroup` hooks run at the
correct time. You do not need to call `t.Parallel()` inside your tests &ndash;
`testgroup` does this for you.

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
