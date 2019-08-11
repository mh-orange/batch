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

func TestSequence(t *testing.T) {
	errFoo := errors.New("Foo")
	f := func(name string, err error) *Step {
		return &Step{Name: name, F: func() error { return err }}
	}

	tests := []struct {
		name    string
		input   []*Step
		wantErr *SequenceErr
	}{
		{"one step, no error", []*Step{f("one", nil)}, nil},
		{"one step, one error", []*Step{f("one", errFoo)}, &SequenceErr{Index: 0, Step: f("one", errFoo), Cause: errFoo}},
		{"two steps, no error", []*Step{f("one", nil), f("two", nil)}, nil},
		{"two steps, first error", []*Step{f("one", errFoo), f("two", nil)}, &SequenceErr{Index: 0, Step: f("one", errFoo), Cause: errFoo}},
		{"two steps, second error", []*Step{f("one", nil), f("two", errFoo)}, &SequenceErr{Index: 1, Step: f("two", errFoo), Cause: errFoo}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotErr := Sequence(test.input...)
			if gotErr == nil && test.wantErr == nil {
				return
			}

			if se, ok := gotErr.(*SequenceErr); ok {
				if test.wantErr.Index != test.wantErr.Index || test.wantErr.Cause != test.wantErr.Cause {
					t.Errorf("Wanted error %v got %v", test.wantErr, se)
				}

				if se.Step != test.input[se.Index] {
					t.Errorf("Wanted step %v got %v", test.input[se.Index], se.Step)
				}
			} else {
				t.Errorf("Wanted a sequence error, got %T", gotErr)
			}
		})
	}
}
