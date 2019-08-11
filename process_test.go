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
	"fmt"
	"testing"
)

func TestErrFormat(t *testing.T) {
	tests := []struct {
		name   string
		input  *Err
		format string
		want   string
	}{
		{"string", &Err{[]*JobErr{{1, errors.New("Foo")}}}, "%s", "Foo\n"},
		{"value", &Err{[]*JobErr{{1, errors.New("Foo")}}}, "%v", "Foo\n"},
		{"quote", &Err{[]*JobErr{{1, errors.New("Foo")}}}, "%q", "\"Foo\"\n"},
		{"default", &Err{[]*JobErr{{1, errors.New("Foo")}}}, "%d", "%!d(batch.Err)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := fmt.Sprintf(test.format, test.input)
			if test.want != got {
				t.Errorf("Wanted string %q got %q", test.want, got)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	errFoo1 := errors.New("Foo1")
	errFoo2 := errors.New("Foo2")

	f := func(err error) Job { return JobFunc(func() error { return err }) }
	tests := []struct {
		name  string
		input []Job
		want  []*JobErr
	}{
		{"one job, no error", []Job{f(nil)}, []*JobErr{}},
		{"one job, one error", []Job{f(errFoo1)}, []*JobErr{{0, errFoo1}}},
		{"two jobs, one error", []Job{f(nil), f(errFoo1)}, []*JobErr{{1, errFoo1}}},
		{"two jobs, two errors", []Job{f(errFoo1), f(errFoo2)}, []*JobErr{{0, errFoo1}, {1, errFoo2}}},
		{"three jobs, two errors", []Job{f(errFoo1), f(nil), f(errFoo2)}, []*JobErr{{0, errFoo1}, {2, errFoo2}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Process(test.input...)
			if len(test.want) == 0 && got == nil {
				return
			}

			if be, ok := got.(*Err); ok {
				for i, want := range test.want {
					if *want != *be.Errs[i] {
						t.Errorf("Wanted job error %d: %+v got %+v", i, want, be.Errs[i])
					}
				}
			} else {
				t.Errorf("Wanted an Err got %T", got)
			}
		})
	}
}
