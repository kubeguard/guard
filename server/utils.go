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
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiextensionClientSet "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
		printStackTrace(err)
	}

	w.WriteHeader(http.StatusOK)
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

func fetchApiResources() ([]*metav1.APIResourceList, error) {
	// creates the in-cluster config
	klog.V(5).Infof("Fetch apiresources list")
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error building kubeconfig")
	}

	kubeclientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error building kubernetes clientset")
	}

	apiresourcesList, err := kubeclientset.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	crdClientset, err := apiextensionClientSet.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	apiresourcesList, err = filterOutCRDs(crdClientset, apiresourcesList)
	if err != nil {
		return nil, err
	}

	projectCalicoApiService := "projectcalico.org/v3"

	noProjectCalicoCondition := func(resourceList *metav1.APIResourceList) bool {
		return projectCalicoApiService != resourceList.GroupVersion
	}

	apiresourcesList = filterResources(apiresourcesList, noProjectCalicoCondition)

	klog.V(5).Infof("Apiresources list : %v", apiresourcesList)

	return apiresourcesList, nil
}

func filterOutCRDs(crdClientset *apiextensionClientSet.Clientset, apiresourcesList []*metav1.APIResourceList) ([]*metav1.APIResourceList, error) {
	crdV1List, err := crdClientset.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if crdV1List != nil && len(crdV1List.Items) >= 0 {
		for _, crd := range crdV1List.Items {
			for _, version := range crd.Spec.Versions {
				groupVersion := path.Join(crd.Spec.Group, version.Name)
				noCrdsCondition := func(resourceList *metav1.APIResourceList) bool {
					return groupVersion != resourceList.GroupVersion
				}

				apiresourcesList = filterResources(apiresourcesList, noCrdsCondition)
			}
		}
	}

	return apiresourcesList, nil
}

func filterResources(apiresourcesList []*metav1.APIResourceList, criteria func(*metav1.APIResourceList) bool) (filteredResources []*metav1.APIResourceList) {
	for _, res := range apiresourcesList {
		if criteria(res) {
			filteredResources = append(filteredResources, res)
		}
	}
	return
}
