# testgroup

[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev-page]
[![Workflow:Main][workflow-main-badge]][workflow-main-page]
[![Coverage][coveralls-main-badge]][coveralls-main-page]
[![Go Report Card][go-report-card-badge]][go-report-card-page]

[coveralls-main-badge]:
  https://coveralls.io/repos/github/bloomberg/go-testgroup/badge.svg?branch=main
[coveralls-main-page]:
  https://coveralls.io/github/bloomberg/go-testgroup?branch=main
  "Coverage for main branch on Coveralls"
[go-report-card-badge]:
  https://goreportcard.com/badge/github.com/bloomberg/go-testgroup
[go-report-card-page]:
  https://goreportcard.com/report/github.com/bloomberg/go-testgroup
  "Go Report Card"
[pkg-go-dev-badge]: https://pkg.go.dev/badge/github.com/bloomberg/go-testgroup
[pkg-go-dev-page]:
  https://pkg.go.dev/github.com/bloomberg/go-testgroup
  "Reference Documentation on pkg.go.dev"
[workflow-main-badge]:
  https://github.com/bloomberg/go-testgroup/workflows/Main/badge.svg
[workflow-main-page]:
  https://github.com/bloomberg/go-testgroup/actions?query=workflow%3AMain
  "Main Github Workflow"

`testgroup` helps you organize tests into groups. A test group is a `struct` (or
other type) whose exported methods are its subtests. The subtests can share
data, helper functions, and pre/post-group and pre/post-test hooks.

`testgroup` was inspired by
[`github.com/stretchr/testify/suite`](https://pkg.go.dev/github.com/stretchr/testify/suite).

## Contents

- [Features](#features)
- [Example](#example)
- [Documentation](#documentation)
  - [Motivation ("Why not `testify/suite`?")](#motivation-why-not-testifysuite)
  - [Writing test groups](#writing-test-groups)
    - [Pre/post-group and pre/post-test hooks (optional)](#prepost-group-and-prepost-test-hooks-optional)
  - [Running test groups](#running-test-groups)
    - [Serially](#serially)
    - [In parallel](#in-parallel)
  - [Using `testgroup.T`](#using-testgroupt)
    - [Running subtests](#running-subtests)
    - [Running subgroups](#running-subgroups)
    - [Using `testing.T`](#using-testingt)
    - [Asserting with `testify/assert` and `testify/require`](#asserting-with-testifyassert-and-testifyrequire)
- [Code of Conduct](#code-of-conduct)
- [Contributing](#contributing)
- [License](#license)
- [Security Policy](#security-policy)

## Features

- Support for parallel execution of tests, including waiting for parallel tests
  to finish
- Easy access to assertion helpers
- Pre/post-group and pre/post-test hooks

## Example

Here's a simple test group from
[`example_absint_test.go`](example_absint_test.go):

```go
package testgroup_test

import (
	"testing"

	"github.com/bloomberg/go-testgroup"
)

func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// This function is the entry point to the test group.
// Because it starts with "Test" and accepts a *testing.T argument,
// it is detected and called by the Go testing package when you run "go test".
func TestAbsInt(t *testing.T) {
	testgroup.RunSerially(t, &AbsIntTests{})
}

type AbsIntTests struct {
	// Group-wide data/state can be stored here.
}

func (grp *AbsIntTests) DoesNotChangeNonNegativeNumbers(t *testgroup.T) {
	t.Equal(0, AbsInt(0))
	t.Equal(1, AbsInt(1))
	t.Equal(123456789, AbsInt(123456789))
}

func (grp *AbsIntTests) MakesNegativeNumbersPositive(t *testgroup.T) {
	t.Equal(1, AbsInt(-1))
	t.Equal(123456789, AbsInt(-123456789))
}
```

When you run `go test`, you'll see something like this:

```console
$ go test -v example_absint_test.go
=== RUN   TestAbsInt
=== RUN   TestAbsInt/DoesNotChangeNonNegativeNumbers
=== RUN   TestAbsInt/MakesNegativeNumbersPositive
--- PASS: TestAbsInt (0.00s)
    --- PASS: TestAbsInt/DoesNotChangeNonNegativeNumbers (0.00s)
    --- PASS: TestAbsInt/MakesNegativeNumbersPositive (0.00s)
PASS
ok  	command-line-arguments	0.014s
```

## Documentation

### Motivation ("Why not `testify/suite`?")

`testgroup` was inspired by [testify][]'s [suite][testify-suite-docs] package.
We really like `testify/suite`, but we had trouble getting subtests to run in
parallel. Testify stores each test's `testing.T` inside the `Suite` struct,
which means that only one `testing.T` is available at a given time. If you run
tests in parallel, failures can be reported as another test failing, and if a
test failure is reported twice, `testing` panics.

After separating the group-wide state from the test-specific state, we also made
a few usability improvements to the `testing.T`-like object.

[testify]: https://github.com/stretchr/testify
[testify-suite-docs]: https://pkg.go.dev/github.com/stretchr/testify/suite

### Writing test groups

A test group is a `struct` (or other type) whose exported methods are its
subtests.

```go
type MyGroup struct{}

func (*MyGroup) Subtest(t *testgroup.T) {
	// ...
}
```

`testgroup` considers _all_ of the type's exported methods to be subtests.
(Unlike `testing`-style tests or `testify/suite` subtests, you don't have to
start `testgroup` subtests with the prefix `Test`.)

A valid subtest accepts a `*testgroup.T` as its only argument and does not
return anything. If a subtest (exported method) has a different signature,
`testgroup` will fail the parent test to avoid accidentally skipping malformed
tests.

#### Pre/post-group and pre/post-test hooks (optional)

`testgroup` considers a few names to be hook methods that have special behavior.
You may find them useful to help clarify your code and avoid some repetition.

`testgroup` supports the following hooks:

```go
func (*MyGroup) PreGroup(t *testgroup.T)  {
	// code that will run before any of MyGroup's subtests have started
}

func (*MyGroup) PostGroup(t *testgroup.T) {
	// code that will run after all of MyGroup's subtests have finished
}

func (*MyGroup) PreTest(t *testgroup.T)  {
	// code that will run at the beginning of each subtest in MyGroup
}

func (*MyGroup) PostTest(t *testgroup.T) {
	// code that will run at the end of each subtest in MyGroup
}
```

Like subtests, these methods accept a single `*testgroup.T` argument.

If you skip a test by calling `t.Skip()`, the `PreTest` and `PostTest` hook
functions will still run before and after that test.

### Running test groups

Here's an example of a top-level `testing`-style test running the subtests in a
test group:

```go
func TestMyGroup(t *testing.T) {
	testgroup.RunSerially(t, &MyGroup{})
}
```

`testgroup` has two ways to run subtests: `RunSerially` and `RunInParallel`.

#### Serially

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

#### In parallel

`RunInParallel` runs a group's subtests in parallel.

When using this mode, **you must not call `t.Parallel()` inside your tests**
&ndash; `testgroup` does this for you.

In order to make sure [hooks](#prepost-group-and-prepost-test-hooks) run at the
correct time, `RunInParallel` wraps a parent test around your subtests. By
default, the parent test is named `_`, but you can override this by setting
`RunInParallelParentTestName`.

The execution order of subtests and hooks looks like this:

1. Run `PreGroup`.
2. In parallel, run the following sequence of steps for each subtest:
   1. Run `PreTest`.
   2. Run the subtest method.
   3. Run `PostTest`.
3. After all subtests finish, run `PostGroup`.

Here's another contrived example:

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

### Using `testgroup.T`

`testgroup.T` is a type passed to each test function. It is mainly concerned
with test state, and it embeds and contains other types for convenience.

#### Running subtests

`testgroup.T.Run` is just like `testing.T.Run`, but its test function has a
`*testgroup.T` argument for convenience.

```go
func (*MyGroup) MySubtest(t *testgroup.T) {
	// set up a table of testcases
	type testcase struct{
		input, output int
	}
	table := []testcase{
		// ...
	}

	for _, tc := range table {
		tc := tc // local copy to pin range variable
		t.Run(fmt.Sprintf("%d", tc.input), func (t *testgroup.T) {
			t.Equal(tc.output, someCalculation(tc.input))
		})
	}
}
```

#### Running subgroups

`testgroup.T` has a few convenience methods that simply wrap the package-level
functions.

- `testgroup.T.RunSerially` calls `testgroup.RunSerially`.
- `testgroup.T.RunInParallel` calls `testgroup.RunInParallel`.

#### Using `testing.T`

`testgroup.T` embeds a `*testing.T`, which lets you write

```go
func (*MyGroup) MySubtest(t *testgroup.T) {
	if testing.Short() {
		t.Skip("skipping due to short mode")
	}

	t.Logf("Have we failed yet? %v", t.Failed())
}
```

#### Asserting with `testify/assert` and `testify/require`

Similar to `testify/suite`, `testgroup.T` embeds a
[`testify/assert`][testify-assert-docs] `*Assertions`, so you can call its
member functions directly from a `testgroup.T`:

```go
func (*MyGroup) MySubtest(t *testgroup.T) {
	const expectedValue = 42
	result := callSomeFunction(input)
	t.Equal(expectedValue, result)
}
```

`testgroup.T` also contains a [`testify/require`][testify-require-docs]
`*Assertions` named `Require`. You can use it to fail your test immediately
instead of continuing.

```go
func (*MyGroup) MySubtest(t *testgroup.T) {
	const expectedValue = 42
	result, err := somethingThatMightError(input)
	t.NoError(err)                  // testify/assert assertion -- continues test execution if it fails
	t.Require.NotNil(result)        // testify/require assertion -- stops test execution if it fails
	t.Equal(expectedValue, result)
}
```

[testify-assert-docs]: https://pkg.go.dev/github.com/stretchr/testify/assert
[testify-require-docs]: https://pkg.go.dev/github.com/stretchr/testify/require

## Code of Conduct

`testgroup` has adopted a
[Code of Conduct](https://github.com/bloomberg/.github/blob/master/CODE_OF_CONDUCT.md).
If you have any concerns about the Code or behavior which you have experienced
in the project, please contact us at opensource@bloomberg.net.

## Contributing

We'd love to hear from you, whether you've found a bug or want to suggest how
`testgroup` could be better. Please
[open an issue](https://github.com/bloomberg/go-testgroup/issues/new/choose) and
let us know what you think!

If you want to contribute code to `testgroup`, please be sure to read our
[contribution guidelines](https://github.com/bloomberg/.github/blob/master/CONTRIBUTING.md).
**We highly recommend opening an issue before you start working on your pull
request.** We'd like to talk with you about the change you want to make _before_
you start making it. :smile:

## License

`testgroup` is licensed under the [Apache License, Version 2.0](LICENSE).

## Security Policy

If you believe you have identified a security vulnerability in this project,
please send an email to the project team at opensource@bloomberg.net detailing
the suspected issue and any methods you've found to reproduce it.

Please do _not_ open an issue in the GitHub repository, as we'd prefer to keep
vulnerability reports private until we've had an opportunity to review and
address them. Thank you.
