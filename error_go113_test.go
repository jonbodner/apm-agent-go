// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// +build go1.13

package apm_test

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	realErr "errors"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.elastic.co/apm/apmtest"

	"go.elastic.co/apm/transport/transporttest"
)

func TestErrorCauseUnwrap(t *testing.T) {
	err := fmt.Errorf("%w", errors.New("cause"))

	tracer, recorder := transporttest.NewRecorderTracer()
	defer tracer.Close()
	tracer.NewError(err).Send()
	tracer.Flush(nil)

	payloads := recorder.Payloads()
	require.Len(t, payloads.Errors, 1)
	assert.Equal(t, "TestErrorCauseUnwrap", payloads.Errors[0].Culprit)

	require.Len(t, payloads.Errors[0].Exception.Cause, 1)
	assert.Equal(t, "cause", payloads.Errors[0].Exception.Cause[0].Message)
}

// StackErr wraps an error with the stack location where the error occurred. Use the WithStack
// function to create a StackErr. There can only be one StackErr in the error chain, ideally at
// the root error location.
type StackErr struct {
	Err   error
	trace []uintptr
}

func (se StackErr) StackTrace() *runtime.Frames {
	frames := runtime.CallersFrames(se.trace)
	return frames
}

// WithStack takes in an error and returns an error wrapped in a StackErr with the location where
// an error was first created or returned from third-party code. If there is already a StackErr
// in the error chain, WithStack returns the passed-in error.
func WithStack(err error) error {
	var se StackErr
	if realErr.As(err, &se) {
		return err
	}
	pc := make([]uintptr, 20)
	n := runtime.Callers(2, pc)
	pc = pc[:n]
	return StackErr{
		Err:   err,
		trace: pc,
	}
}

// Unwrap exposes the error wrapped by StackErr
func (se StackErr) Unwrap() error {
	return se.Err
}

// Error is the marker interface for an error, it returns the wrapped error or an empty string if there is no
// wrapped error
func (se StackErr) Error() string {
	if se.Err == nil {
		return ""
	}
	return se.Err.Error()
}

// Format controls the optional display of the stack trace. Use %+v to output the stack trace, use %v or %s to output
// the wrapped error only, use %q to get a single-quoted character literal safely escaped with Go syntax for the wrapped
// error.
func (se StackErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", se.Unwrap())
			fmt.Fprintf(s, "%s", strings.Join(se.Trace(), "\n"))
			return
		}
		io.WriteString(s, se.Error())
	case 's':
		io.WriteString(s, se.Error())
	case 'q':
		fmt.Fprintf(s, "%q", se.Error())
	}
}

// Trace returns the stack trace information as a slice of strings. Each entry is formatted as
// "FUNCTION_NAME (FILE_NAME:LINE_NUMBER)"
func (se StackErr) Trace() []string {
	s := make([]string,0,len(se.trace))
	frames := se.StackTrace()
	for {
		frame, more := frames.Next()
		s = append(s, fmt.Sprintf("%s (%s:%d)", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return s
}

func TestMyStackErr(t *testing.T) {
	e := realErr.New("This is a test")
	e = WithStack(e)
	e = fmt.Errorf("so sad: %w", e)
	e = WithStack(e)
	tracer := apmtest.NewRecordingTracer()
	defer tracer.Close()

	e2 := tracer.NewError(e)
	e2.Send()
	tracer.Flush(nil)
	fmt.Println(tracer.RecorderTransport.Payloads().Errors[0].Exception.Stacktrace)
}
