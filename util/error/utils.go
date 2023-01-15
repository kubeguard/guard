/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package error

import (
	"fmt"
	"io"

	"k8s.io/klog/v2"
)

// WithCode annotates err with a new code.
// If err is nil, WithCode returns nil.
func WithCode(err error, code int) error {
	if err == nil {
		return nil
	}
	return &withCode{
		cause: err,
		code:  code,
	}
}

type withCode struct {
	cause error
	code  int
}

func (w *withCode) Error() string { return w.cause.Error() }
func (w *withCode) Cause() error  { return w.cause }
func (w *withCode) Code() int     { return w.code }

type HttpStatusCode interface {
	Code() int
}

func (w *withCode) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, err := fmt.Fprintf(s, "%+v\n", w.Cause())
			if err != nil {
				klog.Fatal(err)
			}
			return
		}
		fallthrough
	case 's', 'q':
		_, err := io.WriteString(s, w.Error())
		if err != nil {
			klog.Fatal(err)
		}
	}
}
