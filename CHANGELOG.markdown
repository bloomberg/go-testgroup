# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][], and this project adheres to
[Semantic Versioning][].

[keep a changelog]: https://keepachangelog.com/en/1.0.0/
[semantic versioning]: https://semver.org/spec/v2.0.0.html

## Unreleased ([&#x0394;][unreleased changes])

### Changed

- Renamed `ParallelSeparator` to `RunInParallelParentTestName`.

### Removed

- The `PreGrouper`, `PostGrouper`, `PreTester`, and `PostTester` hook interfaces
  are no longer exported.

## [0.2.0][] ([&#x0394;][0.2.0 changes]) - 2019-09-19

### Added

- `testgroup.T.Run()` wraps `testing.T.Run()`, but passes your test a
  `*testgroup.T` instead of a `*testing.T`. This makes it convenient to use
  `testgroup.T`'s helpers when writing table-driven tests.

### Changed

- The test group will now fail if:

  - the group object has no exported methods
  - the group object is passed by value, and it has exported methods with a
    pointer receiver

  This should help catch mistakes when writing tests.

## [0.1.0][] ([&#x0394;][0.1.0 changes]) - 2019-05-24

First release of the library.

[0.2.0]:
  https://github.com/bloomberg/go-testgroup/releases/tag/v0.2.0
  "version 0.2.0"
[0.1.0]:
  https://github.com/bloomberg/go-testgroup/releases/tag/v0.1.0
  "version 0.1.0"
[unreleased changes]:
  https://github.com/bloomberg/go-testgroup/compare/v0.2.0...HEAD
  "unreleased changes since 0.2.0"
[0.2.0 changes]:
  https://github.com/bloomberg/go-testgroup/compare/v0.1.0...v0.2.0
  "changes from 0.1.0 to 0.2.0"
[0.1.0 changes]:
  https://github.com/bloomberg/go-testgroup/commits/v0.1.0
  "changes from root to 0.1.0"
