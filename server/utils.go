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

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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

	if klog.V(10).Enabled() {
		data, _ := json.MarshalIndent(resp, "", "  ")
		klog.V(10).Infoln(string(data))
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		panic(err)
	}
}

func writeAuthzResponse(w http.ResponseWriter, spec *authzv1.SubjectAccessReviewSpec, accessInfo *authzv1.SubjectAccessReviewStatus, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")

	resp := authzv1.SubjectAccessReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: authzv1.SchemeGroupVersion.String(),
			Kind:       "SubjectAccessReview",
		},
	}

	if spec != nil {
		resp.Spec = *spec
	}

	if accessInfo != nil {
		resp.Status = *accessInfo
	} else {
		accessInfo := authzv1.SubjectAccessReviewStatus{Allowed: false, Denied: true}
		if err != nil {
			accessInfo.Reason = err.Error()
		}
		resp.Status = accessInfo
	}

	if err != nil {
		w.WriteHeader(err.Code())
		printStackTrace(err)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if klog.V(7).Enabled() {
		if _, ok := spec.Extra["oid"]; ok {
			data, _ := json.Marshal(resp)
			klog.V(7).Infof("final data:%s", string(data))
		}
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
	klog.Errorln(err)

	if c, ok := errors.Cause(err).(stackTracer); ok {
		st := c.StackTrace()
		klog.V(5).Infof("Stacktrace: %+v", st) // top two frames
	}
}
