// Copyright 2019 Andrew Bates
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

package batch

import (
	"errors"
	"testing"
)

func TestErrorFuncs(t *testing.T) {
	tests := []struct {
		name  string
		input error
		want  string
	}{
		{"SequenceErr", &SequenceErr{1, &Step{"step", nil}, errors.New("Foo")}, "step failed: Foo"},
		{"JobErr", &JobErr{1, errors.New("Foo")}, "Foo"},
		{"Err (1 JobErr)", &Err{[]*JobErr{{1, errors.New("Foo")}}}, "1 error occurred during batch processing"},
		{"Err (2 JobErrs)", &Err{[]*JobErr{{1, errors.New("Foo")}, {2, errors.New("Bar")}}}, "2 errors occurred during batch processing"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.Error()
			if test.want != got {
				t.Errorf("Wanted %q got %q", test.want, got)
			}
		})
	}
}
