// Copyright 2020 Bloomberg Finance L.P.
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
