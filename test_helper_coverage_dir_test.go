// Copyright 2019 Bloomberg Finance L.P.

package testgroup_test

import (
	"fmt"
	"os"
	"path/filepath"
)

// goTestCoverageArgs generates arguments to add to 'go test' to help gather coverage data from
// tests that run as subprocesses.
//
// Set the COVERAGE_DIR environment variable to the directory where coverage reports should go.
func goTestCoverageArgs(testName string) []string {
	coverageDir := os.Getenv("COVERAGE_DIR")
	if coverageDir == "" {
		return nil
	}

	relPath := filepath.Join(coverageDir, testName)
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		panic(fmt.Sprintf("could not mkdir %q: %v", filepath.Dir(absPath), err))
	}

	return []string{"-coverprofile", absPath}
}
