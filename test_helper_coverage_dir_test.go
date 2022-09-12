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

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		panic(fmt.Sprintf("could not mkdir %q: %v", filepath.Dir(absPath), err))
	}

	return []string{"-coverprofile", absPath}
}
