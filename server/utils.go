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
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"go.kubeguard.dev/guard/auth/providers/azure/graph"
	"go.kubeguard.dev/guard/util/httpclient"
	oputil "go.kubeguard.dev/guard/util/operations"

	"github.com/Azure/go-autorest/autorest/azure"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	v "gomodules.xyz/x/version"
	auth "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func fetchListOfResources(clusterType string, environment string, loginURL string, tenantId string) (map[string][]map[string]map[string]oputil.DataAction, error) {
	operationsMap := map[string][]map[string]map[string]oputil.DataAction{}

	apiResourcesList, err := fetchApiResources()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch list of api-resources from apiserver.")
	}

	operationsList, err := fetchDataActionsList(environment, clusterType, loginURL, tenantId)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch operations from Azure.")
	}

	for _, resList := range apiResourcesList {
		group := "v1" // core api group
		if resList.GroupVersion != "" && resList.GroupVersion != "v1" {
			group = strings.Split(resList.GroupVersion, "/")[0]
		}

		if len(resList.APIResources) == 0 {
			continue
		}

		for _, apiResource := range resList.APIResources {
			if strings.Contains(apiResource.Name, "/") {
				continue
			}

			actionId := clusterType
			if group != "v1" {
				actionId = path.Join(actionId, group)
			}

			resourceName := apiResource.Name

			actionId = path.Join(actionId, resourceName)

			for _, operation := range operationsList {
				if strings.Contains(operation.Name, actionId) {
					opNameArr := strings.Split(operation.Name, "/")

					if group != "v1" {
						// extra validation to make sure groups are the same
						if !(group == opNameArr[2]) {
							continue
						}
					} else {
						// make sure resources are the same for core apigroup
						if !(resourceName == opNameArr[2]) {
							continue
						}
					}

					verb := opNameArr[len(opNameArr)-1]
					if verb == "action" {
						verb = path.Join(opNameArr[len(opNameArr)-2], opNameArr[len(opNameArr)-1])
					}

					da := oputil.DataAction{
						ActionInfo: oputil.AuthorizationActionInfo{
							IsDataAction: true,
						},
						IsNamespacedResource: apiResource.Namespaced,
					}
					da.ActionInfo.AuthorizationEntity.Id = operation.Name

					if operationsMap[group] == nil {
						operationsMap[group] = []map[string]map[string]oputil.DataAction{}
					}
					resourceAndVerb := map[string]map[string]oputil.DataAction{}
					resourceAndVerb[resourceName] = map[string]oputil.DataAction{}
					resourceAndVerb[resourceName][verb] = da
					operationsMap[group] = append(operationsMap[group], resourceAndVerb)
				}
			}
		}
	}

	klog.V(5).Infof("Operations list: %v", operationsMap)

	return operationsMap, nil
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

	klog.V(5).Infof("List of ApiResources fetched from apiserver: %v", apiresourcesList)

	return apiresourcesList, nil
}

func fetchDataActionsList(environment string, clusterType string, loginURL string, tenantID string) ([]oputil.Operation, error) {
	env := azure.PublicCloud
	var err error
	if environment != "" {
		env, err = azure.EnvironmentFromName(environment)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse environment for Azure.")
		}
	}

	endpoint := ""

	switch clusterType {
	case oputil.ConnectedClusters:
		endpoint = fmt.Sprintf(oputil.OperationsEndpointFormatARC, env.ResourceManagerEndpoint)
	case oputil.ManagedClusters:
		endpoint = fmt.Sprintf(oputil.OperationsEndpointFormatAKS, env.ResourceManagerEndpoint)
	case oputil.Fleets:
		endpoint = fmt.Sprintf(oputil.OperationsEndpointFormatAKS, env.ResourceManagerEndpoint)
	default:
		return nil, errors.New("Failed to create endpoint for Get Operations call.")
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create request for Get Operations call.")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("guard-%s-%s-%s", v.Version.Platform, v.Version.GoVersion, v.Version.Version))

	var token string
	if clusterType == oputil.ConnectedClusters {
		tokenResp, err := oputil.GetAuthHeaderUsingMSIForARC(env.ResourceManagerEndpoint)
		if err != nil {
			return nil, errors.Wrap(err, "Error getting authorization headers for Get Operations call.")
		}

		token = tokenResp.AccessToken
	} else if clusterType == oputil.ManagedClusters || clusterType == oputil.Fleets {
		tokenProvider := graph.NewAKSTokenProvider(loginURL, tenantID)

		authResp, err := tokenProvider.Acquire("")
		if err != nil {
			return nil, errors.Wrap(err, "Error getting authorization headers for Get Operations call.")
		}

		token = authResp.Token
	} else {
		return nil, errors.New("Unsupported clusterType for Get Operations call.")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := httpclient.DefaultHTTPClient

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to send request for Get Operations call.")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error in reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Request failed with status code: %d and response: %s", resp.StatusCode, string(data))
	}

	operationsList := oputil.OperationList{}
	err = json.Unmarshal(data, &operationsList)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode response")
	}

	var finalOperations []oputil.Operation
	for _, op := range operationsList.Value {
		if *op.IsDataAction && strings.Contains(op.Name, clusterType) {
			finalOperations = append(finalOperations, op)
		}
	}

	return finalOperations, nil
}
