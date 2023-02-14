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

package azure

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func Test_createOperationsMap(t *testing.T) {
	type args struct {
		apiResourcesList []*metav1.APIResourceList
		operationsList   []Operation
		clusterType      string
	}

	tests := []struct {
		name string
		args args
		want OperationsMap
	}{
		{
			"all operations should be present",
			args{
				apiResourcesList: []*metav1.APIResourceList{
					{
						GroupVersion: "v1",
						APIResources: []metav1.APIResource{
							{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
							{Name: "nodes", Namespaced: false, Kind: "Node", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
					{
						GroupVersion: "apps/v1",
						APIResources: []metav1.APIResource{
							{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
				},
				operationsList: []Operation{
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/apps/deployments/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "deployments", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/apps/deployments/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "deployments", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/apps/deployments/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "deployments", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/nodes/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "nodes", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/nodes/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "nodes", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},

					{
						Name:         "Microsoft.Kubernetes/connectedClusters/nodes/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "nodes", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/exec/action",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
				},
				clusterType: "Microsoft.Kubernetes/connectedClusters",
			},
			OperationsMap{
				"apps": ResourceAndVerbMap{
					"deployments": VerbAndActionsMap{
						"read":   DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/apps/deployments/read"}, IsDataAction: true}, IsNamespacedResource: true},
						"write":  DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/apps/deployments/write"}, IsDataAction: true}, IsNamespacedResource: true},
						"delete": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/apps/deployments/delete"}, IsDataAction: true}, IsNamespacedResource: true},
					},
				},
				"v1": ResourceAndVerbMap{
					"nodes": VerbAndActionsMap{
						"read":   DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/nodes/read"}, IsDataAction: true}, IsNamespacedResource: false},
						"write":  DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/nodes/write"}, IsDataAction: true}, IsNamespacedResource: false},
						"delete": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/nodes/delete"}, IsDataAction: true}, IsNamespacedResource: false},
					},
					"pods": VerbAndActionsMap{
						"read":        DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/read"}, IsDataAction: true}, IsNamespacedResource: true},
						"write":       DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/write"}, IsDataAction: true}, IsNamespacedResource: true},
						"delete":      DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/delete"}, IsDataAction: true}, IsNamespacedResource: true},
						"exec/action": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/exec/action"}, IsDataAction: true}, IsNamespacedResource: true},
					},
				},
			},
		},

		{
			"only groups there in Azure should be present",
			args{
				apiResourcesList: []*metav1.APIResourceList{
					{
						GroupVersion: "v1",
						APIResources: []metav1.APIResource{
							{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
					{
						GroupVersion: "somerandomgroup/v1",
						APIResources: []metav1.APIResource{
							{Name: "somerandomres", Namespaced: true, Kind: "Somerandomres", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
				},
				operationsList: []Operation{
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/exec/action",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
				},
				clusterType: "Microsoft.Kubernetes/connectedClusters",
			},
			OperationsMap{
				"v1": ResourceAndVerbMap{
					"pods": VerbAndActionsMap{
						"read":        DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/read"}, IsDataAction: true}, IsNamespacedResource: true},
						"write":       DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/write"}, IsDataAction: true}, IsNamespacedResource: true},
						"delete":      DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/delete"}, IsDataAction: true}, IsNamespacedResource: true},
						"exec/action": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/exec/action"}, IsDataAction: true}, IsNamespacedResource: true},
					},
				},
			},
		},

		{
			"only resources there in Azure should be present",
			args{
				apiResourcesList: []*metav1.APIResourceList{
					{
						GroupVersion: "v1",
						APIResources: []metav1.APIResource{
							{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
					{
						GroupVersion: "apps/v1",
						APIResources: []metav1.APIResource{
							{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
							{Name: "somerandomres", Namespaced: true, Kind: "Somerandomres", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
				},
				operationsList: []Operation{
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/apps/deployments/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "deployments", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/apps/deployments/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "deployments", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/apps/deployments/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "deployments", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/exec/action",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/pods/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "pods", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
				},
				clusterType: "Microsoft.Kubernetes/connectedClusters",
			},
			OperationsMap{
				"apps": ResourceAndVerbMap{
					"deployments": VerbAndActionsMap{
						"read":   DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/apps/deployments/read"}, IsDataAction: true}, IsNamespacedResource: true},
						"write":  DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/apps/deployments/write"}, IsDataAction: true}, IsNamespacedResource: true},
						"delete": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/apps/deployments/delete"}, IsDataAction: true}, IsNamespacedResource: true},
					},
				},
				"v1": ResourceAndVerbMap{
					"pods": VerbAndActionsMap{
						"read":        DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/read"}, IsDataAction: true}, IsNamespacedResource: true},
						"write":       DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/write"}, IsDataAction: true}, IsNamespacedResource: true},
						"delete":      DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/delete"}, IsDataAction: true}, IsNamespacedResource: true},
						"exec/action": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/pods/exec/action"}, IsDataAction: true}, IsNamespacedResource: true},
					},
				},
			},
		},

		{
			"same resource names in multiple groups should be present only in their groups",
			args{
				apiResourcesList: []*metav1.APIResourceList{
					{
						GroupVersion: "v1",
						APIResources: []metav1.APIResource{
							{Name: "events", Namespaced: true, Kind: "Event", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
					{
						GroupVersion: "events.k8s.io/v1",
						APIResources: []metav1.APIResource{
							{Name: "events", Namespaced: true, Kind: "Event", Verbs: []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"}},
						},
					},
				},
				operationsList: []Operation{
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/events/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "events", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/events.k8s.io/events/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "events", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/events/read",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "events", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/events/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "events", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/events.k8s.io/events/delete",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "events", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
					{
						Name:         "Microsoft.Kubernetes/connectedClusters/events.k8s.io/events/write",
						Display:      Display{Provider: "Microsoft.Kubernetes", Resource: "events", Operation: "some desc", Description: "some desc"},
						IsDataAction: pointer.Bool(true),
					},
				},
				clusterType: "Microsoft.Kubernetes/connectedClusters",
			},
			OperationsMap{
				"events.k8s.io": ResourceAndVerbMap{
					"events": VerbAndActionsMap{
						"read":   DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/events.k8s.io/events/read"}, IsDataAction: true}, IsNamespacedResource: true},
						"write":  DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/events.k8s.io/events/write"}, IsDataAction: true}, IsNamespacedResource: true},
						"delete": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/events.k8s.io/events/delete"}, IsDataAction: true}, IsNamespacedResource: true},
					},
				},
				"v1": ResourceAndVerbMap{
					"events": VerbAndActionsMap{
						"read":   DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/events/read"}, IsDataAction: true}, IsNamespacedResource: true},
						"write":  DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/events/write"}, IsDataAction: true}, IsNamespacedResource: true},
						"delete": DataAction{ActionInfo: AuthorizationActionInfo{AuthorizationEntity: AuthorizationEntity{Id: "Microsoft.Kubernetes/connectedClusters/events/delete"}, IsDataAction: true}, IsNamespacedResource: true},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		operationsMap = NewOperationsMap()
		_ = SetDiscoverResourcesSettings(tt.args.clusterType, "", "", "", "", "", "")
		createOperationsMap(tt.args.apiResourcesList, tt.args.operationsList)
		got := GetOperationsMap()
		if len(got) != len(tt.want) {
			t.Errorf("[createOperationsMap()]Map lengths are not equal. = %v, want %v", got, tt.want)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("createOperationsMap() = %v, want %v", got, tt.want)
		}
	}
}
