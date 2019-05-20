# testgroup

This test grouping framework was inspired by and based on
[testify](https://github.com/stretchr/testify)'s
[suite](https://godoc.org/github.com/stretchr/testify/suite) package, but it
supports parallelism in the subtests within a test group.

Testify is copyright &copy; 2012-2018 Mat Ryer and Tyler Bunnell.

## TODO

- Document: Calling `t.Skip()` doesnt skip `PreTest` or `PostTest`.
