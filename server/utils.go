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

package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/golang/glog"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// write replies to the request with the specified TokenReview object and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
func write(w http.ResponseWriter, info *auth.UserInfo, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")

	resp := auth.TokenReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auth.SchemeGroupVersion.String(),
			Kind:       "TokenReview",
		},
	}

	if err != nil {
		code := http.StatusUnauthorized
		if v, ok := err.(httpStatusCode); ok {
			code = v.Code()
		}
		printStackTrace(err)
		w.WriteHeader(code)
		resp.Status = auth.TokenReviewStatus{
			Authenticated: false,
			Error:         err.Error(),
		}
	} else {
		w.WriteHeader(http.StatusOK)
		resp.Status = auth.TokenReviewStatus{
			Authenticated: true,
			User:          *info,
		}
	}

	if glog.V(10) {
		data, _ := json.MarshalIndent(resp, "", "  ")
		glog.V(10).Infoln(string(data))
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		panic(err)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type httpStatusCode interface {
	Code() int
}

func printStackTrace(err error) {
	glog.Errorln(err)

	if c, ok := errors.Cause(err).(stackTracer); ok {
		st := c.StackTrace()
		glog.V(5).Infof("Stacktrace: %+v", st) // top two frames
	}
}

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

func (w *withCode) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, err := fmt.Fprintf(s, "%+v\n", w.Cause())
			if err != nil {
				glog.Fatal(err)
			}
			return
		}
		fallthrough
	case 's', 'q':
		_, err := io.WriteString(s, w.Error())
		if err != nil {
			glog.Fatal(err)
		}
	}
}
