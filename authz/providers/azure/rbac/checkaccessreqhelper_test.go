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

package rbac

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"

	azureutils "go.kubeguard.dev/guard/util/azure"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	authzv1 "k8s.io/api/authorization/v1"
)

const (
	resourceId      = "resourceId"
	aksClusterType  = "aks"
	subresourceAttr = "Microsoft.ContainerService/managedClusters/resources:subresource"
)

func createOperationsMap(clusterType string) azureutils.OperationsMap {
	return azureutils.OperationsMap{
		"apps": azureutils.ResourceAndVerbMap{
			"deployments": azureutils.VerbAndActionsMap{
				"read":   azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/apps/deployments/read", clusterType)}, IsDataAction: true}, IsNamespacedResource: true},
				"write":  azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/apps/deployments/write", clusterType)}, IsDataAction: true}, IsNamespacedResource: true},
				"delete": azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/apps/deployments/delete", clusterType)}, IsDataAction: true}, IsNamespacedResource: true},
			},
		},
		"v1": azureutils.ResourceAndVerbMap{
			"persistentvolumes": azureutils.VerbAndActionsMap{
				"read":   azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/persistentvolumes/read", clusterType)}, IsDataAction: true}, IsNamespacedResource: false},
				"write":  azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/persistentvolumes/write", clusterType)}, IsDataAction: true}, IsNamespacedResource: false},
				"delete": azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/persistentvolumes/delete", clusterType)}, IsDataAction: true}, IsNamespacedResource: false},
			},
			"pods": azureutils.VerbAndActionsMap{
				"read":        azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/pods/read", clusterType)}, IsDataAction: true}, IsNamespacedResource: true},
				"write":       azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/pods/write", clusterType)}, IsDataAction: true}, IsNamespacedResource: true},
				"delete":      azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/pods/delete", clusterType)}, IsDataAction: true}, IsNamespacedResource: true},
				"exec/action": azureutils.DataAction{ActionInfo: azureutils.AuthorizationActionInfo{AuthorizationEntity: azureutils.AuthorizationEntity{Id: fmt.Sprintf("%s/pods/exec/action", clusterType)}, IsDataAction: true}, IsNamespacedResource: true},
			},
		},
	}
}

func Test_getScope(t *testing.T) {
	type args struct {
		resourceId                      string
		attr                            *authzv1.ResourceAttributes
		useNamespaceResourceScopeFormat bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nilAttr", args{"resourceId", nil, false}, "resourceId"},
		{"bothnil", args{"", nil, false}, ""},
		{"emptyRes", args{"", &authzv1.ResourceAttributes{Namespace: ""}, false}, ""},
		{"emptyNS", args{"resourceId", &authzv1.ResourceAttributes{Namespace: ""}, false}, "resourceId"},
		{"bothPresent", args{"resourceId", &authzv1.ResourceAttributes{Namespace: "test"}, false}, "resourceId/namespaces/test"},
		{"nilAttrNewScope", args{"resourceId", nil, true}, "resourceId"},
		{"bothnilNewScope", args{"", nil, true}, ""},
		{"emptyResNewScope", args{"", &authzv1.ResourceAttributes{Namespace: ""}, true}, ""},
		{"emptyNSNewScope", args{"resourceId", &authzv1.ResourceAttributes{Namespace: ""}, true}, "resourceId"},
		{"bothPresentNewScope", args{"resourceId", &authzv1.ResourceAttributes{Namespace: "test"}, true}, "resourceId/providers/Microsoft.KubernetesConfiguration/namespaces/test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getScope(tt.args.resourceId, tt.args.attr, tt.args.useNamespaceResourceScopeFormat); got != tt.want {
				t.Errorf("getScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getManagedNamespaceScope(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		request  *authzv1.SubjectAccessReviewSpec
		wantOk   bool
		wantPath string
	}{
		{
			name:     "nil ResourceAttributes",
			request:  &authzv1.SubjectAccessReviewSpec{ResourceAttributes: nil},
			wantOk:   false,
			wantPath: "",
		},
		{
			name:     "empty Namespace",
			request:  &authzv1.SubjectAccessReviewSpec{ResourceAttributes: &authzv1.ResourceAttributes{Namespace: ""}},
			wantOk:   false,
			wantPath: "",
		},
		{
			name:     "valid Namespace",
			request:  &authzv1.SubjectAccessReviewSpec{ResourceAttributes: &authzv1.ResourceAttributes{Namespace: "dev"}},
			wantOk:   true,
			wantPath: "managedNamespaces/dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotOk, gotPath := getManagedNameSpaceScope(tt.request)
			if gotOk != tt.wantOk || gotPath != tt.wantPath {
				t.Errorf("getManagedNamespaceScope() = (%v, %q), want (%v, %q)", gotOk, gotPath, tt.wantOk, tt.wantPath)
			}
		})
	}
}

func Test_getValidSecurityGroups(t *testing.T) {
	type args struct {
		groups []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"nilGroup", args{nil}, nil},
		{"emptyGroup", args{[]string{}}, nil},
		{"noGuidGroup", args{[]string{"abc", "def", "system:ghi"}}, nil},
		{
			"someGroup",
			args{[]string{"abc", "1cffe3ae-93c0-4a87-9484-2e90e682aae9", "sys:admin", "", "0ab7f20f-8e9a-43ba-b5ac-1811c91b3d40"}},
			[]string{"1cffe3ae-93c0-4a87-9484-2e90e682aae9", "0ab7f20f-8e9a-43ba-b5ac-1811c91b3d40"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getValidSecurityGroups(tt.args.groups); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValidSecurityGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDataActions(t *testing.T) {
	type args struct {
		isCrTest       bool
		isSubresTest   bool
		isWildcardTest bool
		subRevReq      *authzv1.SubjectAccessReviewSpec
		clusterType    string
	}
	tests := []struct {
		name string
		args args
		want []azureutils.AuthorizationActionInfo
	}{
		{
			aksClusterType,
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					NonResourceAttributes: &authzv1.NonResourceAttributes{Path: "/apis", Verb: "list"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apis/read"}, IsDataAction: true}},
		},

		{
			"aks2",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					NonResourceAttributes: &authzv1.NonResourceAttributes{Path: "/logs", Verb: "get"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/logs/read"}, IsDataAction: true}},
		},

		{
			"fleet",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					NonResourceAttributes: &authzv1.NonResourceAttributes{Path: "/apis", Verb: "list"},
				}, clusterType: "fleet",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "fleet/apis/read"}, IsDataAction: true}},
		},

		{
			"fleet2",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					NonResourceAttributes: &authzv1.NonResourceAttributes{Path: "/logs", Verb: "get"},
				}, clusterType: "fleet",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "fleet/logs/read"}, IsDataAction: true}},
		},

		{
			"arc",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "pods", Subresource: "status", Version: "v1", Name: "test", Verb: "delete"},
				}, clusterType: "arc",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "arc/pods/delete"}, IsDataAction: true}},
		},

		{
			"arc2",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "apps", Resource: "apps/deployments", Subresource: "status", Version: "v1", Name: "test", Verb: "create"},
				}, clusterType: "arc",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "arc/apps/apps/deployments/write"}, IsDataAction: true}},
		},

		{
			"arc3",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "policy", Resource: "podsecuritypolicies", Subresource: "status", Version: "v1", Name: "test", Verb: "use"},
				}, clusterType: "arc",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "arc/policy/podsecuritypolicies/use/action"}, IsDataAction: true}},
		},

		{
			"aks3",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "authentication.k8s.io", Resource: "userextras", Subresource: "scopes", Version: "v1", Name: "test", Verb: "impersonate"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/authentication.k8s.io/userextras/impersonate/action"}, IsDataAction: true}},
		},

		{
			"fleet3",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "authentication.k8s.io", Resource: "userextras", Subresource: "scopes", Version: "v1", Name: "test", Verb: "impersonate"},
				}, clusterType: "fleet",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "fleet/authentication.k8s.io/userextras/impersonate/action"}, IsDataAction: true}},
		},

		{
			"arc4",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "rbac.authorization.k8s.io", Resource: "clusterroles", Subresource: "status", Version: "v1", Name: "test", Verb: "bind"},
				}, clusterType: "arc",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "arc/rbac.authorization.k8s.io/clusterroles/bind/action"}, IsDataAction: true}},
		},

		{
			"aks4",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "rbac.authorization.k8s.io", Resource: "clusterroles", Subresource: "status", Version: "v1", Name: "test", Verb: "escalate"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/rbac.authorization.k8s.io/clusterroles/escalate/action"}, IsDataAction: true}},
		},

		{
			"fleet4",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "rbac.authorization.k8s.io", Resource: "clusterroles", Subresource: "status", Version: "v1", Name: "test", Verb: "escalate"},
				}, clusterType: "fleet",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "fleet/rbac.authorization.k8s.io/clusterroles/escalate/action"}, IsDataAction: true}},
		},

		{
			"arc5",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "scheduling.k8s.io", Resource: "priorityclasses", Subresource: "status", Version: "v1", Name: "test", Verb: "update"},
				}, clusterType: "arc",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "arc/scheduling.k8s.io/priorityclasses/write"}, IsDataAction: true}},
		},

		{
			"aks5",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "events.k8s.io", Resource: "events", Subresource: "status", Version: "v1", Name: "test", Verb: "watch"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/events.k8s.io/events/read"}, IsDataAction: true}},
		},

		{
			"fleet5",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "events.k8s.io", Resource: "events", Subresource: "status", Version: "v1", Name: "test", Verb: "watch"},
				}, clusterType: "fleet",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "fleet/events.k8s.io/events/read"}, IsDataAction: true}},
		},

		{
			"arc6",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "batch", Resource: "cronjobs", Subresource: "status", Version: "v1", Name: "test", Verb: "patch"},
				}, clusterType: "arc",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "arc/batch/cronjobs/write"}, IsDataAction: true}},
		},

		{
			"aks6",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "certificates.k8s.io", Resource: "certificatesigningrequests", Subresource: "approvals", Version: "v1", Name: "test", Verb: "deletecollection"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/certificates.k8s.io/certificatesigningrequests/delete"}, IsDataAction: true}},
		},

		{
			"fleet6",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "certificates.k8s.io", Resource: "certificatesigningrequests", Subresource: "approvals", Version: "v1", Name: "test", Verb: "deletecollection"},
				}, clusterType: "fleet",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "fleet/certificates.k8s.io/certificatesigningrequests/delete"}, IsDataAction: true}},
		},

		{
			aksClusterType,
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "pods", Subresource: "exec", Version: "v1", Name: "test", Verb: "create"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/exec/action"}, IsDataAction: true}},
		},

		{
			"fleet",
			args{
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "pods", Subresource: "exec", Version: "v1", Name: "test", Verb: "create"},
				}, clusterType: "fleet",
			},
			[]azureutils.AuthorizationActionInfo{{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "fleet/pods/exec/action"}, IsDataAction: true}},
		},

		{
			"allStar",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "*", Resource: "*", Subresource: "*", Version: "*", Name: "test", Verb: "*"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/persistentvolumes/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/persistentvolumes/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/persistentvolumes/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/exec/action"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/read"}, IsDataAction: true},
			},
		},

		{
			"allStarNSscope",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "*", Resource: "*", Subresource: "*", Version: "*", Name: "test", Verb: "*", Namespace: "kube-system"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/exec/action"}, IsDataAction: true},
			},
		},

		{
			"resAndGroupStarNSScope",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "*", Resource: "*", Subresource: "*", Version: "*", Name: "test", Verb: "get", Namespace: "kube-system"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/read"}, IsDataAction: true},
			},
		},

		{
			"verbAndGroupStar",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "*", Resource: "pods", Subresource: "*", Version: "*", Name: "test", Verb: "*", Namespace: "kube-system"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/exec/action"}, IsDataAction: true},
			},
		},

		{
			"verbAndResStarNsScope",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "*", Subresource: "*", Version: "*", Name: "test", Verb: "*", Namespace: "kube-system"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/exec/action"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/read"}, IsDataAction: true},
			},
		},

		{
			"verbIsStar",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "persistentvolumes", Subresource: "*", Version: "*", Name: "test", Verb: "*"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/persistentvolumes/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/persistentvolumes/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/persistentvolumes/delete"}, IsDataAction: true},
			},
		},

		{
			"resourceIsStar",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "*", Subresource: "*", Version: "*", Name: "test", Verb: "delete"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/persistentvolumes/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/delete"}, IsDataAction: true},
			},
		},

		{
			"groupIsStar",
			args{
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "*", Resource: "deployments", Subresource: "*", Version: "*", Name: "test", Verb: "*", Namespace: "kube-system"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/delete"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/apps/deployments/write"}, IsDataAction: true},
			},
		},

		{
			"customResource",
			args{
				isCrTest:       true,
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "customresources.contoso.io", Resource: "contosoCustomResource", Subresource: "*", Version: "*", Name: "test", Verb: "list"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/read"}, IsDataAction: true},
			},
		},

		{
			"crResourceIsStar",
			args{
				isCrTest:       true,
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "customresources.contoso.io", Resource: "*", Subresource: "*", Version: "*", Name: "test", Verb: "get"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/read"}, IsDataAction: true},
			},
		},

		{
			"crVerbIsStar",
			args{
				isCrTest:       true,
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "customresources.contoso.io", Resource: "contosoCustomResource", Subresource: "*", Version: "*", Name: "test", Verb: "*"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/delete"}, IsDataAction: true},
			},
		},

		{
			"crResourceAndVerbIsStar",
			args{
				isCrTest:       true,
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "customresources.contoso.io", Resource: "*", Subresource: "*", Version: "*", Name: "test", Verb: "*"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/read"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/write"}, IsDataAction: true},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/delete"}, IsDataAction: true},
			},
		},

		{
			"customResourceBuiltinAPI",
			args{
				isCrTest:       true,
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "apps", Resource: "aCustomResource", Subresource: "status", Version: "v1", Name: "test", Verb: "get"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/read"}, IsDataAction: true},
			},
		},

		{
			"subresourceLogs",
			args{
				isSubresTest:   true,
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "pods", Subresource: "logs", Version: "*", Name: "test", Verb: "get"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/read"}, IsDataAction: true, Attributes: map[string]string{subresourceAttr: "logs"}},
			},
		},

		{
			"subresourceScale",
			args{
				isSubresTest:   true,
				isWildcardTest: false,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "deployments", Subresource: "scale", Version: "*", Name: "test", Verb: "update"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/deployments/write"}, IsDataAction: true, Attributes: map[string]string{subresourceAttr: "scale"}},
			},
		},

		{
			"subresourceVerbIsStar",
			args{
				isSubresTest:   true,
				isWildcardTest: true,
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					ResourceAttributes: &authzv1.ResourceAttributes{Group: "", Resource: "pods", Subresource: "logs", Version: "*", Name: "test", Verb: "*"},
				}, clusterType: aksClusterType,
			},
			[]azureutils.AuthorizationActionInfo{
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/read"}, IsDataAction: true, Attributes: map[string]string{subresourceAttr: "logs"}},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/write"}, IsDataAction: true, Attributes: map[string]string{subresourceAttr: "logs"}},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/delete"}, IsDataAction: true, Attributes: map[string]string{subresourceAttr: "logs"}},
				{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/pods/exec/action"}, IsDataAction: true, Attributes: map[string]string{subresourceAttr: "logs"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operationsMap := createOperationsMap(tt.args.clusterType)
			getStoredOperationsMap = func() azureutils.OperationsMap {
				return operationsMap
			}

			ctx := context.Background()
			got, _ := getDataActions(ctx, tt.args.subRevReq, tt.args.clusterType, tt.args.isCrTest, tt.args.isSubresTest)
			if !tt.args.isWildcardTest {
				if !reflect.DeepEqual(got[0].AuthorizationEntity, tt.want[0].AuthorizationEntity) {
					t.Errorf("getDataActions() = %v, want %v", got, tt.want)
				}

				for key, value := range tt.want[0].Attributes {
					if got[0].Attributes[key] != value {
						t.Errorf("getDataActions() Attributes [%v]: got \"%v\", want \"%v\"", key, got[0].Attributes[key], value)
						break
					}
				}
			} else {
				gotSet := createSet(got)
				wantSet := createSet(tt.want)
				if len(gotSet) != len(wantSet) {
					t.Errorf("getDataActions() Length is not equal got = %d, want %d", len(got), len(tt.want))
				}

				for authinfo := range gotSet {
					if _, ok := wantSet[authinfo]; !ok {
						t.Errorf("getDataActions() = %v, want %v", got, tt.want)
						break
					}
					gotAttr := gotSet[authinfo].Attributes
					wantAttr := wantSet[authinfo].Attributes
					for key, value := range wantAttr {
						if gotAttr[key] != value {
							t.Errorf("getDataActions() Attributes [%v]: got \"%v\", want \"%v\"", key, gotAttr[key], value)
							break
						}
					}
				}
			}
		})
	}
}

func createSet(authinfos []azureutils.AuthorizationActionInfo) map[azureutils.AuthorizationEntity]azureutils.AuthorizationActionInfo {
	set := make(map[azureutils.AuthorizationEntity]azureutils.AuthorizationActionInfo)
	for _, elem := range authinfos {
		set[elem.AuthorizationEntity] = elem
	}
	return set
}

func Test_getNameSpaceScope(t *testing.T) {
	req := authzv1.SubjectAccessReviewSpec{ResourceAttributes: nil}
	want := false
	got, str := getNameSpaceScope(&req, false)
	if got || str != "" {
		t.Errorf("Want:%v, got:%v", want, got)
	}

	req = authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{Namespace: ""},
	}
	want = false
	got, str = getNameSpaceScope(&req, false)
	if got || str != "" {
		t.Errorf("Want:%v, got:%v", want, got)
	}

	req = authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{Namespace: "dev"},
	}
	outputstring := "namespaces/dev"
	want = true
	got, str = getNameSpaceScope(&req, false)
	if !got || str != outputstring {
		t.Errorf("Want:%v - %s, got: %v - %s", want, outputstring, got, str)
	}
}

func Test_getNameSpaceScopeUsingNewNsFormat(t *testing.T) {
	req := authzv1.SubjectAccessReviewSpec{ResourceAttributes: nil}
	want := false
	got, str := getNameSpaceScope(&req, true)
	if got || str != "" {
		t.Errorf("Want:%v, got:%v", want, got)
	}

	req = authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{Namespace: ""},
	}
	want = false
	got, str = getNameSpaceScope(&req, true)
	if got || str != "" {
		t.Errorf("Want:%v, got:%v", want, got)
	}

	req = authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{Namespace: "dev"},
	}
	outputstring := "/providers/Microsoft.KubernetesConfiguration/namespaces/dev"
	want = true
	got, str = getNameSpaceScope(&req, true)
	if !got || str != outputstring {
		t.Errorf("Want:%v - %s, got: %v - %s", want, outputstring, got, str)
	}
}

func Test_prepareCheckAccessRequestBody(t *testing.T) {
	req := &authzv1.SubjectAccessReviewSpec{Extra: nil}
	clusterType := aksClusterType
	createOperationsMap(clusterType)
	wantErr := errors.New("oid info not sent from authenticatoin module")

	ctx := context.Background()
	got, gotErr := prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, false, true, false)

	if got != nil && gotErr != wantErr {
		t.Errorf("Want:%v WantErr:%v, got:%v, gotErr:%v", nil, wantErr, got, gotErr)
	}

	req = &authzv1.SubjectAccessReviewSpec{Extra: map[string]authzv1.ExtraValue{"oid": {"test"}}}
	clusterType = "arc"
	wantErr = errors.New("oid info sent from authenticatoin module is not valid")

	got, gotErr = prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, false, true, false)

	if got != nil && gotErr != wantErr {
		t.Errorf("Want:%v WantErr:%v, got:%v, gotErr:%v", nil, wantErr, got, gotErr)
	}
}

func Test_prepareCheckAccessRequestBodyWithNamespace(t *testing.T) {
	dummyUuid := uuid.New()
	req := &authzv1.SubjectAccessReviewSpec{ResourceAttributes: &authzv1.ResourceAttributes{Namespace: "dev"}, Extra: map[string]authzv1.ExtraValue{"oid": {dummyUuid.String()}}}
	clusterType := aksClusterType
	createOperationsMap(clusterType)

	// testing with new ns scope format
	var want string = "resourceId/providers/Microsoft.KubernetesConfiguration/namespaces/dev"

	ctx := context.Background()
	got, gotErr := prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, true, true, false)

	if got == nil {
		t.Errorf("Want: not nil Got: nil, gotErr:%v", gotErr)
	}

	if got != nil && got[0].Resource.Id != want {
		t.Errorf("Want:%v, got:%v", want, got)
	}

	// testing with the old namespace format
	want = "resourceId/namespaces/dev"

	got, gotErr = prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, false, true, false)
	if got == nil {
		t.Errorf("Want: not nil Got: nil, gotErr:%v", gotErr)
	}

	if got != nil && got[0].Resource.Id != want {
		t.Errorf("Want:%v, got:%v", want, got)
	}
}

func Test_prepareCheckAccessRequestBodyWithCustomResource(t *testing.T) {
	req := &authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "dev",
			Group:     "customresources.contoso.io",
			Resource:  "contosoCustomResource",
			Version:   "v1",
			Name:      "test",
			Verb:      "get",
		},
		Extra: map[string]authzv1.ExtraValue{
			"oid": {
				uuid.NewString(),
			},
		},
	}
	clusterType := aksClusterType
	createOperationsMap(clusterType)

	ctx := context.Background()
	got, _ := prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, false, true, false)

	if got == nil {
		t.Errorf("Want: not nil Got: nil")
	}

	if got[0].Actions[0].AuthorizationEntity.Id != "aks/customresources/read" {
		t.Errorf("Want:%v, got:%v", "aks/customresources/read", got[0].Actions[0].AuthorizationEntity.Id)
	}

	if _, found := got[0].Actions[0].Attributes["Microsoft.ContainerService/managedClusters/customResources:kind"]; !found {
		t.Errorf("Microsoft.ContainerService/managedClusters/customResources:kind Attribute is not present")
	}

	if _, found := got[0].Actions[0].Attributes["Microsoft.ContainerService/managedClusters/customResources:group"]; !found {
		t.Errorf("Microsoft.ContainerService/managedClusters/customResources:group Attribute is not present")
	}
}

func Test_prepareCheckAccessRequestBodyWithCustomResourceOperationsMapEmpty(t *testing.T) {
	req := &authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "dev",
			Group:     "customresources.contoso.io",
			Resource:  "contosoCustomResource",
			Version:   "v1",
			Name:      "test",
			Verb:      "get",
		},
		Extra: map[string]authzv1.ExtraValue{
			"oid": {
				uuid.NewString(),
			},
		},
	}
	clusterType := aksClusterType

	getStoredOperationsMap = func() azureutils.OperationsMap {
		return azureutils.OperationsMap{}
	}

	ctx := context.Background()
	got, _ := prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, false, true, false)

	if got == nil {
		t.Errorf("Want: not nil Got: nil")
	}

	if got[0].Actions[0].AuthorizationEntity.Id != "aks/customresources.contoso.io/contosoCustomResource/read" {
		t.Errorf("Want:%v, got:%v", "aks/customresources.contoso.io/contosoCustomResource/read", got[0].Actions[0].AuthorizationEntity.Id)
	}
}

func Test_prepareCheckAccessRequestBodyWithCustomResourceTypeCheckDisabled(t *testing.T) {
	req := &authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "dev",
			Group:     "customresources.contoso.io",
			Resource:  "contosoCustomResource",
			Version:   "v1",
			Name:      "test",
			Verb:      "get",
		},
		Extra: map[string]authzv1.ExtraValue{
			"oid": {
				uuid.NewString(),
			},
		},
	}
	clusterType := aksClusterType
	operationsMap := createOperationsMap(clusterType)

	getStoredOperationsMap = func() azureutils.OperationsMap {
		return operationsMap
	}

	ctx := context.Background()
	got, _ := prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, false, false, false)

	if got == nil {
		t.Errorf("Want: not nil Got: nil")
	}

	if got[0].Actions[0].AuthorizationEntity.Id != "aks/customresources.contoso.io/contosoCustomResource/read" {
		t.Errorf("Want:%v, got:%v", "aks/customresources.contoso.io/contosoCustomResource/read", got[0].Actions[0].AuthorizationEntity.Id)
	}
}

func Test_prepareCheckAccessRequestBodyWithCustomResourceAndStars(t *testing.T) {
	req := &authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace: "dev",
			Group:     "customresources.contoso.io",
			Resource:  "*",
			Version:   "v1",
			Name:      "test",
			Verb:      "*",
		},
		Extra: map[string]authzv1.ExtraValue{
			"oid": {
				uuid.NewString(),
			},
		},
	}
	clusterType := aksClusterType
	operationsMap := createOperationsMap(clusterType)

	getStoredOperationsMap = func() azureutils.OperationsMap {
		return operationsMap
	}

	ctx := context.Background()
	got, _ := prepareCheckAccessRequestBody(ctx, req, clusterType, resourceId, false, true, false)

	if got == nil {
		t.Errorf("Want: not nil Got: nil")
	}

	customResourceActions := []azureutils.AuthorizationActionInfo{
		{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/read"}, IsDataAction: true, Attributes: map[string]string{
			"Microsoft.ContainerService/managedClusters/customResources:kind":  "*",
			"Microsoft.ContainerService/managedClusters/customResources:group": "customresources.contoso.io",
		}},
		{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/write"}, IsDataAction: true, Attributes: map[string]string{
			"Microsoft.ContainerService/managedClusters/customResources:kind":  "*",
			"Microsoft.ContainerService/managedClusters/customResources:group": "customresources.contoso.io",
		}},
		{AuthorizationEntity: azureutils.AuthorizationEntity{Id: "aks/customresources/delete"}, IsDataAction: true, Attributes: map[string]string{
			"Microsoft.ContainerService/managedClusters/customResources:kind":  "*",
			"Microsoft.ContainerService/managedClusters/customResources:group": "customresources.contoso.io",
		}},
	}

	if len(got[0].Actions) != len(customResourceActions) {
		t.Errorf("Expected %d actions, got %d", len(customResourceActions), len(got))
	}

	for i := range customResourceActions {
		if !reflect.DeepEqual(got[0].Actions[i].Attributes, customResourceActions[i].Attributes) {
			t.Errorf("Expected action %v, got %v", customResourceActions[i].AuthorizationEntity.Id, got[0].Actions[i].AuthorizationEntity.Id)
		}
	}
}

func Test_prepareCheckAccessRequestBodyWithFleetMembers(t *testing.T) {
	id := uuid.New()
	fleetResourceID := "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.ContainerService/fleets/my-fleet"

	tests := []struct {
		name         string
		req          *authzv1.SubjectAccessReviewSpec
		clusterType  string
		resourceID   string
		wantResource string
		wantActions  []string
	}{
		{
			name: "fleet members without namespace",
			req: &authzv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authzv1.ResourceAttributes{
					Group:    "",
					Resource: "nodes",
					Verb:     "get",
				},
				Extra: map[string]authzv1.ExtraValue{"oid": {id.String()}},
			},
			clusterType:  fleetMembers,
			resourceID:   fleetResourceID,
			wantResource: fleetResourceID,
			wantActions:  []string{"Microsoft.ContainerService/fleets/members/nodes/read"},
		},
		{
			name: "fleet members with namespace",
			req: &authzv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authzv1.ResourceAttributes{
					Namespace: "dev",
					Group:     "",
					Resource:  "pods",
					Verb:      "create",
				},
				Extra: map[string]authzv1.ExtraValue{"oid": {id.String()}},
			},
			clusterType: fleetMembers,
			resourceID:  fleetResourceID,
			// prepareCheckAccessRequestBody only generates 'namespaces'
			// CheckAccess retrofits the resource ID subsequently after calling this function.
			wantResource: fleetResourceID + "/namespaces/dev",
			wantActions:  []string{"Microsoft.ContainerService/fleets/members/pods/write"},
		},
		{
			name: "fleet members with delete verb",
			req: &authzv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authzv1.ResourceAttributes{
					Namespace: "prod",
					Group:     "",
					Resource:  "deployments",
					Verb:      "delete",
				},
				Extra: map[string]authzv1.ExtraValue{"oid": {id.String()}},
			},
			clusterType:  fleetMembers,
			resourceID:   fleetResourceID,
			wantResource: fleetResourceID + "/namespaces/prod",
			wantActions:  []string{"Microsoft.ContainerService/fleets/members/deployments/delete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, gotErr := prepareCheckAccessRequestBody(ctx, tt.req, tt.clusterType, tt.resourceID, false, false, false)

			if gotErr != nil {
				t.Errorf("Unexpected error: %v", gotErr)
				return
			}

			if len(got) == 0 {
				t.Error("Expected non-empty result")
				return
			}

			// Verify Resource.Id
			if got[0].Resource.Id != tt.wantResource {
				t.Errorf("Resource.Id: want %q, got %q", tt.wantResource, got[0].Resource.Id)
			}

			// Verify Actions
			if len(got[0].Actions) != len(tt.wantActions) {
				t.Errorf("Actions count: want %d, got %d", len(tt.wantActions), len(got[0].Actions))
				return
			}

			for i, wantAction := range tt.wantActions {
				if got[0].Actions[i].AuthorizationEntity.Id != wantAction {
					t.Errorf("Action[%d]: want %q, got %q", i, wantAction, got[0].Actions[i].AuthorizationEntity.Id)
				}

				// Verify IsDataAction is true for fleet members
				if !got[0].Actions[i].IsDataAction {
					t.Errorf("Action[%d].IsDataAction: want true, got false", i)
				}
			}

			// Verify Subject.Attributes.ObjectId is set
			if got[0].Subject.Attributes.ObjectId != id.String() {
				t.Errorf("Subject.Attributes.ObjectId: want %q, got %q", id.String(), got[0].Subject.Attributes.ObjectId)
			}
		})
	}
}

func Test_prepareCheckAccessRequestBodyWithSubresource(t *testing.T) {
	req := &authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace:   "dev",
			Group:       "",
			Resource:    "pods",
			Subresource: "logs",
			Version:     "v1",
			Name:        "test",
			Verb:        "get",
		},
		Extra: map[string]authzv1.ExtraValue{
			"oid": {
				uuid.NewString(),
			},
		},
	}
	clusterType := aksClusterType
	createOperationsMap(clusterType)

	got, _ := prepareCheckAccessRequestBody(context.Background(), req, clusterType, resourceId, false, false, true)

	if got == nil {
		t.Errorf("Want: not nil Got: nil")
	}

	if got[0].Actions[0].AuthorizationEntity.Id != "aks/pods/read" {
		t.Errorf("Want:%v, got:%v", "aks/pods/read", got[0].Actions[0].AuthorizationEntity.Id)
	}

	value, found := got[0].Actions[0].Attributes[subresourceAttr]
	wantSubres := "logs"

	if !found {
		t.Errorf("%v attribute is not present", subresourceAttr)
	} else if value != wantSubres {
		t.Errorf("Subresource attribute - want: %v, got: %v", wantSubres, value)
	}
}

func Test_prepareCheckAccessRequestBodyWithSubresourceDisabled(t *testing.T) {
	req := &authzv1.SubjectAccessReviewSpec{
		ResourceAttributes: &authzv1.ResourceAttributes{
			Namespace:   "dev",
			Group:       "",
			Resource:    "pods",
			Subresource: "logs",
			Version:     "v1",
			Name:        "test",
			Verb:        "get",
		},
		Extra: map[string]authzv1.ExtraValue{
			"oid": {
				uuid.NewString(),
			},
		},
	}
	clusterType := aksClusterType
	createOperationsMap(clusterType)

	got, _ := prepareCheckAccessRequestBody(context.Background(), req, clusterType, resourceId, false, false, false)

	if got == nil {
		t.Errorf("Want: not nil Got: nil")
	}

	if got[0].Actions[0].AuthorizationEntity.Id != "aks/pods/read" {
		t.Errorf("Want:%v, got:%v", "aks/pods/read", got[0].Actions[0].AuthorizationEntity.Id)
	}

	if _, found := got[0].Actions[0].Attributes[subresourceAttr]; found {
		t.Errorf("%v attribute is present when it should not", subresourceAttr)
	}
}

func Test_getResultCacheKey(t *testing.T) {
	type args struct {
		subRevReq                 *authzv1.SubjectAccessReviewSpec
		allowSubresourceTypeCheck bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			aksClusterType,
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User:                  "charlie@yahoo.com",
					NonResourceAttributes: &authzv1.NonResourceAttributes{Path: "/apis/v1", Verb: "list"},
				},
				allowSubresourceTypeCheck: false,
			},
			"charlie@yahoo.com/apis/v1/read",
		},

		{
			aksClusterType,
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User:                  "echo@outlook.com",
					NonResourceAttributes: &authzv1.NonResourceAttributes{Path: "/logs", Verb: "get"},
				},
				allowSubresourceTypeCheck: false,
			},
			"echo@outlook.com/logs/read",
		},

		{
			aksClusterType,
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "alpha@bing.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "dev", Group: "", Resource: "pods",
						Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
					},
				},
				allowSubresourceTypeCheck: false,
			},
			"alpha@bing.com/dev/-/pods/delete",
		},

		{
			aksClusterType,
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "alpha@bing.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "dev", Group: "", Resource: "pods",
						Subresource: "status", Version: "v1", Name: "test", Verb: "delete",
					},
				},
				allowSubresourceTypeCheck: true,
			},
			"alpha@bing.com/dev/-/pods/delete",
		},

		{
			aksClusterType,
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "alpha@bing.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "dev", Group: "", Resource: "pods",
						Subresource: "logs", Version: "v1", Name: "test", Verb: "get",
					},
				},
				allowSubresourceTypeCheck: false,
			},
			"alpha@bing.com/dev/-/pods/read",
		},

		{
			aksClusterType,
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "alpha@bing.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "dev", Group: "", Resource: "pods",
						Subresource: "logs", Version: "v1", Name: "test", Verb: "get",
					},
				},
				allowSubresourceTypeCheck: true,
			},
			"alpha@bing.com/dev/-/pods/read/logs",
		},

		{
			"arc",
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "beta@msn.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "azure-arc",
						Group:     "authentication.k8s.io", Resource: "userextras", Subresource: "scopes", Version: "v1",
						Name: "test", Verb: "impersonate",
					},
				},
				allowSubresourceTypeCheck: false,
			},
			"beta@msn.com/azure-arc/authentication.k8s.io/userextras/impersonate/action",
		},

		{
			"arc",
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "beta@msn.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "", Group: "", Resource: "nodes",
						Subresource: "scopes", Version: "v1", Name: "", Verb: "list",
					},
				},
				allowSubresourceTypeCheck: false,
			},
			"beta@msn.com/-/-/nodes/read",
		},

		{
			"allStar",
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "beta@msn.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "", Group: "*", Resource: "*",
						Subresource: "scopes", Version: "v1", Name: "", Verb: "*",
					},
				},
				allowSubresourceTypeCheck: false,
			},
			"beta@msn.com/-/*/*/*",
		},

		{
			"allStarNSscope",
			args{
				subRevReq: &authzv1.SubjectAccessReviewSpec{
					User: "beta@msn.com",
					ResourceAttributes: &authzv1.ResourceAttributes{
						Namespace: "dev", Group: "*", Resource: "*",
						Subresource: "scopes", Version: "v1", Name: "", Verb: "*",
					},
				},
				allowSubresourceTypeCheck: false,
			},
			"beta@msn.com/dev/*/*/*",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getResultCacheKey(tt.args.subRevReq, tt.args.allowSubresourceTypeCheck); got != tt.want {
				t.Errorf("getResultCacheKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildCheckAccessURL(t *testing.T) {
	mustCreateURL := func(rawURL string) url.URL {
		t.Helper()

		parsedURL, err := url.Parse(rawURL)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return *parsedURL
	}

	const testAzureResourceID = "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName"
	testResourceIDSegCount := len(strings.Split(testAzureResourceID, "/"))

	tests := []struct {
		name          string
		baseURL       url.URL
		resourceID    string
		hasNamespace  bool
		namespacePath string
		want          string
		wantErr       bool
	}{
		// valid test cases
		{
			name:          "valid without namespace",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  false,
			namespacePath: "",
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		{
			name:          "valid with namespace",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: namespaces + "/dev",
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/namespaces/dev/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		{
			name:          "valid with managed namespace",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: managedNamespaces + "/dev",
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/managedNamespaces/dev/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		{
			name:          "valid with fleet resource ID without namespace",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.ContainerService/fleets/my-fleet",
			hasNamespace:  false,
			namespacePath: "",
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.ContainerService/fleets/my-fleet/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		{
			name:          "valid with fleet resource ID with managed namespace",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.ContainerService/fleets/my-fleet",
			hasNamespace:  true,
			namespacePath: managedNamespaces + "/dev",
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.ContainerService/fleets/my-fleet/managedNamespaces/dev/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		{
			name:          "valid with sub resource",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/subResource",
			hasNamespace:  false,
			namespacePath: namespaces + "/dev/pods",
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/subResource/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		{
			name:          "valid with previous set sub path",
			baseURL:       mustCreateURL("https://management.azure.com/test-sub-path"),
			resourceID:    "/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/subResource",
			hasNamespace:  false,
			namespacePath: namespaces + "/dev/pods",
			want:          "https://management.azure.com/test-sub-path/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/subResource/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		// invalid test cases
		// invariant 1
		{
			name:          "invalid scheme",
			baseURL:       mustCreateURL("http://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  false,
			namespacePath: "",
			wantErr:       true,
		},
		// invariant 2
		{
			name:          "empty resource ID",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    "",
			hasNamespace:  false,
			namespacePath: "",
			wantErr:       true,
		},
		// invariant 4
		{
			name:          "path traversal in namespace path",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: "../",
			wantErr:       true,
		},
		{
			name:          "path traversal in namespace path",
			baseURL:       mustCreateURL("https://management.azure.com/test-sub-path"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: "../",
			wantErr:       true,
		},
		{
			name:          "path traversal in namespace path",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: strings.Repeat("../", testResourceIDSegCount) + "/dev",
			wantErr:       true,
		},
		{
			name:          "path traversal in namespace path",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: strings.Repeat("../", testResourceIDSegCount+1) + "/dev",
			wantErr:       true,
		},
		// url encoded input
		{
			name:          "url encoded data",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: namespaces + "%2E%2E%2F%2E%2E%2Fdev",
			wantErr:       false,
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/namespaces%252E%252E%252F%252E%252E%252Fdev/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
		{
			name:          "url encoded data",
			baseURL:       mustCreateURL("https://management.azure.com"),
			resourceID:    testAzureResourceID,
			hasNamespace:  true,
			namespacePath: namespaces + "%2Fdev",
			wantErr:       false,
			want:          "https://management.azure.com/subscriptions/12345678-1234-1234-1234-123456789abc/resourceGroups/testResourceGroup/providers/Microsoft.Provider/resourceTypes/resourceName/namespaces%252Fdev/providers/Microsoft.Authorization/checkaccess?api-version=2018-09-01-preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildCheckAccessURL(tt.baseURL, tt.resourceID, tt.hasNamespace, tt.namespacePath)
			if tt.wantErr {
				assert.Errorf(t, err, "expect error, but got none. Got: %q", got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
