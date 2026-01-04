/*
 *
 * Copyright 2020 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package status implements errors returned by gRPC.  These errors are
// serialized and transmitted on the wire between server and client, and allow
// for additional data to be transmitted via the Details field in the status
// proto.  gRPC service handlers should return an error created by this
// package, and gRPC clients should expect a corresponding error to be
// returned from the RPC call.
//
// This package upholds the invariants that a non-nil error may not
// contain an OK code, and an OK code must result in a nil error.
package status

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	perrors "github.com/pkg/errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	spb "google.golang.org/genproto/googleapis/rpc/status"
)

import (
	"github.com/dubbogo/grpc-go/codes"
)

// Status represents an RPC status code, message, and details.  It is immutable
// and should be created with New, Newf, or FromProto.
type Status struct {
	s     *spb.Status
	stack *stack
}

// New returns a Status representing c and msg, with user-made error as stack
func New(c codes.Code, msg string) *Status {
	newStatus := &Status{s: &spb.Status{Code: int32(c), Message: msg}}
	newStatusWithDetail, _ := newStatus.WithDetails(&errdetails.DebugInfo{
		StackEntries: []string{
			fmt.Sprintf("%+v", callers().StackTrace()), // use e.String() as triple stack
		},
	})
	return newStatusWithDetail
}

// New returns a Status representing c and msg. with e.String() as triple stack field
func NewWithoutStacks(c codes.Code, e error) *Status {
	newStatus := &Status{s: &spb.Status{Code: int32(c), Message: e.Error()}}
	newStatusWithDetail, _ := newStatus.WithDetails(&errdetails.DebugInfo{
		StackEntries: []string{
			fmt.Sprintf("%+v", e), // use e.String() as triple stack
		},
	})
	return newStatusWithDetail
}

// Newf returns New(c, fmt.Sprintf(format, a...)).
func Newf(c codes.Code, format string, a ...interface{}) *Status {
	return New(c, fmt.Sprintf(format, a...))
}

// FromProto returns a Status representing s.
func FromProto(s *spb.Status) *Status {
	return &Status{s: proto.Clone(s).(*spb.Status)}
}

// Err returns an error representing c and msg.  If c is OK, returns nil.
func Err(c codes.Code, msg string) error {
	return New(c, msg).Err()
}

// Errorf returns Error(c, fmt.Sprintf(format, a...)).
func Errorf(c codes.Code, format string, a ...interface{}) error {
	return Err(c, fmt.Sprintf(format, a...))
}

// Code returns the status code contained in s.
func (s *Status) Code() codes.Code {
	if s == nil || s.s == nil {
		return codes.OK
	}
	return codes.Code(s.s.Code)
}

// Message returns the message contained in s.
func (s *Status) Message() string {
	if s == nil || s.s == nil {
		return ""
	}
	return s.s.Message
}

// Proto returns s's status as an spb.Status proto message.
func (s *Status) Proto() *spb.Status {
	if s == nil {
		return nil
	}
	return proto.Clone(s.s).(*spb.Status)
}

// Err returns an immutable error representing s; returns nil if s.Code() is OK.
func (s *Status) Err() error {
	if s.Code() == codes.OK {
		return nil
	}
	return &Error{s: s}
}

// WithDetails returns a new status with the provided details messages appended to the status.
// If any errors are encountered, it returns nil and the first error encountered.
func (s *Status) WithDetails(details ...proto.Message) (*Status, error) {
	if s.Code() == codes.OK {
		return nil, errors.New("no error details for status with code OK")
	}
	// s.Code() != OK implies that s.Proto() != nil.
	p := s.Proto()
	for _, detail := range details {
		any, err := ptypes.MarshalAny(detail)
		if err != nil {
			return nil, err
		}
		p.Details = append(p.Details, any)
	}
	return &Status{s: p}, nil
}

// Details returns a slice of details messages attached to the status.
// If a detail cannot be decoded, the error is returned in place of the detail.
func (s *Status) Details() []interface{} {
	if s == nil || s.s == nil {
		return nil
	}
	details := make([]interface{}, 0, len(s.s.Details))
	for _, any := range s.s.Details {
		detail := &ptypes.DynamicAny{}
		if err := ptypes.UnmarshalAny(any, detail); err != nil {
			details = append(details, err)
			continue
		}
		details = append(details, detail.Message)
	}
	return details
}

func (s *Status) String() string {
	return fmt.Sprintf("%s", s.Message())
}

// Error wraps a pointer of a status proto. It implements error and Status,
// and a nil *Error should never be returned by this package.
type Error struct {
	s     *Status
	stack stack
}

func (e *Error) Error() string {
	return e.s.String()
}

func (e *Error) Message() string {
	return e.s.String()
}

func (e *Error) Code() codes.Code {
	return e.s.Code()
}

// GRPCStatus returns the Status represented by se.
func (e *Error) GRPCStatus() *Status {
	return e.s
}

// Status returns the Status represented by se.
func (e *Error) Stacks() string {
	if e.s == nil {
		return ""
	}
	if len(e.s.s.Details) == 0 {
		return ""
	}
	stackTracesStr := strings.Replace(e.s.s.Details[0].String(), `\n`, "\n", -1)
	stackTracesStr = strings.Replace(stackTracesStr, `\t`, "\t", -1)
	return stackTracesStr
}

// Is implements future error.Is functionality.
// A Error is equivalent if the code and message are identical.
func (e *Error) Is(target error) bool {
	tse, ok := target.(*Error)
	if !ok {
		return false
	}
	return proto.Equal(e.s.s, tse.s.s)
}

// stack represents a stack of program counters.
type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := perrors.Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (s *stack) StackTrace() perrors.StackTrace {
	f := make([]perrors.Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = perrors.Frame((*s)[i])
	}
	return f
}

func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(5, pcs[:])
	var st stack = pcs[0:n]
	return &st
}
