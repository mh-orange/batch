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

// Job represents one item in a list of things (usually similar) that need
// to be processed as a batch
type Job interface {
	Execute() error // Execute is the function that will be called by Process
}

// JobErr wraps any error that Job.Execute returns. A JobErr will include the
// index of the job (from the list passed in to Process) as well as the Cause
// (the error returned by Execute())
type JobErr struct {
	Index int
	Cause error
}

// JobFunc wraps a standalone function as a Job
type JobFunc func() error

// Execute will call the wrapped function
func (jf JobFunc) Execute() error { return jf() }

// Error will return the error string of the underlying cause
func (je *JobErr) Error() string {
	return je.Cause.Error()
}

// Err is an error returned by Process when one or more of the
// jobs failed
type Err struct {
	Errs []*JobErr // Errs is the list of JobErr errors that occurred during the batch processing
}

// Error will indicate the number of errors that occurred during the batch
// process
func (be *Err) Error() string {
	errstr := "error"
	if len(be.Errs) > 1 {
		errstr = "errors"
	}
	return fmt.Sprintf("%d %s occurred during batch processing", len(be.Errs), errstr)
}

// Format implements fmt.Formatter
func (be *Err) Format(f fmt.State, c rune) {
	switch c {
	case 's':
		fallthrough
	case 'v':
		for _, err := range be.Errs {
			fmt.Fprintf(f, "%s\n", err.Error())
		}
	case 'q':
		for _, err := range be.Errs {
			fmt.Fprintf(f, "%q\n", err.Error())
		}
	default:
		fmt.Fprintf(f, "%%!%c(batch.Err)", c)
	}
}

// Process takes a list of Jobs and batch processes them. Any
// job that results in an error will be recorded as a JobErr
// and will be returned in the list of JobErrs.
func Process(jobs ...Job) error {
	var be *Err
	for i, job := range jobs {
		err := job.Execute()
		if err != nil {
			if be == nil {
				be = &Err{}
			}
			be.Errs = append(be.Errs, &JobErr{i, err})
		}
	}
	return be
}
