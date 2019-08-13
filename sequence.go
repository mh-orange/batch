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
	"fmt"
)

// Step is a step in a sequence, it includes a descriptive name and
// a function to be called
type Step struct {
	Name string
	F    func() error
}

// SequenceErr indicates a given step failed in a sequence
type SequenceErr struct {
	Index int   // Index is the step number that failed (0 based)
	Step  *Step // Step is a pointer to the failed step
	Cause error // Cause is the underlying error that caused the sequence to fail (returned by the Step's function)
}

// Error returns the description of the error, in this case the
// name of the step and the description of the underlying cause
func (se *SequenceErr) Error() string {
	return fmt.Sprintf("%s failed: %v", se.Step.Name, se.Cause)
}

// Sequence will execute a sequence of steps and stop at
// the first step that fails.  If any step fails, execution
// is stopped an an error of type *SequenceErr is returned
func Sequence(steps ...*Step) (err error) {
	for i, step := range steps {
		cause := step.F()
		if cause != nil {
			err = &SequenceErr{i, step, cause}
			break
		}
	}
	return err
}
