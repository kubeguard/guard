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
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	AccessAllowedVerdict        = "Access allowed by Azure RBAC"
	AccessAllowedVerboseVerdict = "Access allowed by Azure RBAC Role Assignment %s of Role %s to user %s"
	Allowed                     = "allowed"
	AccessNotAllowedVerdict     = "User does not have access to the resource in Azure. Update role assignment to allow access."
	NamespaceResourceFormat     = "/providers/Microsoft.KubernetesConfiguration/namespaces"
	namespaces                  = "namespaces"
	NoOpinionVerdict            = "Azure does not have opinion for this user."
	NonAADUserNoOpVerdict       = "Azure does not have opinion for this non AAD user. If you are an AAD user, please set Extra:oid parameter for impersonated user in the kubeconfig"
	NonAADUserNotAllowedVerdict = "Access denied by Azure RBAC for non AAD users. Configure --azure.skip-authz-for-non-aad-users to enable access. If you are an AAD user, please set Extra:oid parameter for impersonated user in the kubeconfig."
)

var username string

type SubjectInfoAttributes struct {
	ObjectId string   `json:"ObjectId"`
	Groups   []string `json:"Groups,omitempty"`
}

type SubjectInfo struct {
	Attributes SubjectInfoAttributes `json:"Attributes"`
}

type AuthorizationEntity struct {
	Id string `json:"Id"`
}

type AuthorizationActionInfo struct {
	AuthorizationEntity
	IsDataAction bool `json:"IsDataAction"`
}

type CheckAccessRequest struct {
	Subject  SubjectInfo               `json:"Subject"`
	Actions  []AuthorizationActionInfo `json:"Actions"`
	Resource AuthorizationEntity       `json:"Resource"`
}

type AccessDecision struct {
	Decision string `json:"accessDecision"`
}

type RoleAssignment struct {
	Id               string `json:"id"`
	RoleDefinitionId string `json:"roleDefinitionId"`
	PrincipalId      string `json:"principalId"`
	PrincipalType    string `json:"principalType"`
	Scope            string `json:"scope"`
	Condition        string `json:"condition"`
	ConditionVersion string `json:"conditionVersion"`
	CanDelegate      bool   `json:"canDelegate"`
}

type AzureRoleAssignment struct {
	DelegatedManagedIdentityResourceId string `json:"delegatedManagedIdentityResourceId"`
	RoleAssignment
}

type Permission struct {
	Actions          []string `json:"actions,omitempty"`
	NoActions        []string `json:"noactions,omitempty"`
	DataActions      []string `json:"dataactions,omitempty"`
	NoDataActions    []string `json:"nodataactions,omitempty"`
	Condition        string   `json:"condition"`
	ConditionVersion string   `json:"conditionVersion"`
}

type Principal struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type DenyAssignment struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Permission
	Scope                   string `json:"scope"`
	DoNotApplyToChildScopes bool   `json:"doNotApplyToChildScopes"`
	Principals              []Principal
	ExcludePrincipals       []Principal
	Condition               string `json:"condition"`
	ConditionVersion        string `json:"conditionVersion"`
}

type AzureDenyAssignment struct {
	MetaData          map[string]interface{} `json:"metadata"`
	IsSystemProtected string                 `json:"isSystemProtected"`
	IsBuiltIn         bool                   `json:"isBuiltIn"`
	DenyAssignment
}

type AuthorizationDecision struct {
	Decision            string              `json:"accessDecision"`
	ActionId            string              `json:"actionId"`
	IsDataAction        bool                `json:"isDataAction"`
	AzureRoleAssignment AzureRoleAssignment `json:"roleAssignment,omitempty"`
	AzureDenyAssignment AzureDenyAssignment `json:"denyAssignment,omitempty"`
	TimeToLiveInMs      int                 `json:"timeToLiveInMs"`
}

func getScope(resourceId string, attr *authzv1.ResourceAttributes, useNamespaceResourceScopeFormat bool) string {
	if attr != nil && attr.Namespace != "" {
		if useNamespaceResourceScopeFormat {
			return path.Join(resourceId, NamespaceResourceFormat, attr.Namespace)
		} else {
			return path.Join(resourceId, namespaces, attr.Namespace)
		}
	}
	return resourceId
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func getValidSecurityGroups(groups []string) []string {
	var finalGroups []string
	for _, element := range groups {
		if isValidUUID(element) {
			finalGroups = append(finalGroups, element)
		}
	}
	return finalGroups
}

func getActionName(verb string) string {
	/* Kubernetes supprots some special verbs for which we need to return data action /verb/action.
	Following is the list of special verbs and their API group/resources in Kubernetes:
	use:policy/podsecuritypolicies
	bind:rbac.authorization.k8s.io/roles,clusterroles
	escalate:rbac.authorization.k8s.io/roles,clusterroles
	impersonate:core/users,groups,serviceaccounts
	impersonate:authentication.k8s.io/userextras

	https://kubernetes.io/docs/reference/access-authn-authz/authorization/#determine-the-request-verb
	*/
	switch verb {
	case "get":
		fallthrough
	case "list":
		fallthrough
	case "watch":
		return "read"

	case "bind":
		return "bind/action"
	case "escalate":
		return "escalate/action"
	case "use":
		return "use/action"
	case "impersonate":
		return "impersonate/action"

	case "create":
		fallthrough // instead of action create will be mapped to write
	case "patch":
		fallthrough
	case "update":
		return "write"

	case "delete":
		fallthrough
	case "deletecollection": // TODO: verify scenario
		return "delete"
	default:
		return ""
	}
}

func getResourceAndAction(resource string, subResource string, verb string) string {
	var action string

	if resource == "pods" && subResource == "exec" {
		action = path.Join(resource, subResource, "action")
	} else {
		action = path.Join(resource, getActionName(verb))
	}

	return action
}

func getDataActions(subRevReq *authzv1.SubjectAccessReviewSpec, clusterType string, apiresourcesList []*metav1.APIResourceList) []AuthorizationActionInfo {

	var authInfoList []AuthorizationActionInfo
	filterOnApigroup := func(resourceList *metav1.APIResourceList) bool {
		group := strings.Split(resourceList.GroupVersion, "/")[0]
		return group == subRevReq.ResourceAttributes.Group
	}

	if subRevReq.ResourceAttributes != nil {

		if subRevReq.ResourceAttributes.Resource != "*" && subRevReq.ResourceAttributes.Verb != "*" {
			authInfoSingle := AuthorizationActionInfo{
				IsDataAction: true,
			}

			authInfoSingle.AuthorizationEntity.Id = clusterType

			if subRevReq.ResourceAttributes.Group != "" {
				authInfoSingle.AuthorizationEntity.Id = path.Join(authInfoSingle.AuthorizationEntity.Id, subRevReq.ResourceAttributes.Group)
			}

			action := getResourceAndAction(subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Subresource, subRevReq.ResourceAttributes.Verb)
			authInfoSingle.AuthorizationEntity.Id = path.Join(authInfoSingle.AuthorizationEntity.Id, action)
			authInfoList = append(authInfoList, authInfoSingle)
		} else if subRevReq.ResourceAttributes.Resource == "*" {
			var filteredResources []*metav1.APIResourceList
			if subRevReq.ResourceAttributes.Group == "*" {
				// fetch resources under all apigroups
				filteredResources = apiresourcesList
			} else if subRevReq.ResourceAttributes.Group != "" {
				// fetch resources under specified apigroup
				filteredResources = filterResources(apiresourcesList, filterOnApigroup)
			} else {
				// if Group is not there that means it is the core apigroup
				onlyCoreResources := func(resourceList *metav1.APIResourceList) bool { return resourceList.GroupVersion == "v1" }
				filteredResources = filterResources(apiresourcesList, onlyCoreResources)
			}

			// if Namespace is empty or is populated, we need to create the list only for namespace scoped resources
			var finalFilteredResources []*metav1.APIResourceList
			if subRevReq.ResourceAttributes.Namespace == "" || subRevReq.ResourceAttributes.Namespace != "" {
				for _, apiResList := range filteredResources {
					var resourceList []metav1.APIResource
					for _, resource := range apiResList.APIResources {
						if resource.Namespaced {
							resourceList = append(resourceList, resource)
						}
					}
					apiResList.APIResources = resourceList
					if len(apiResList.APIResources) > 0 {
						finalFilteredResources = append(finalFilteredResources, apiResList)
					}
				}
			}

			klog.V(5).Infof("Final filtered resource : %v", finalFilteredResources)

			// create list of Data Actions
			authInfoList = createAuthorizationActionInfoList(finalFilteredResources, subRevReq.ResourceAttributes.Verb, clusterType)
		} else {
			// this case will only come if resource is not * and verb is *
			var finalFilteredResource []*metav1.APIResourceList
			filterResource := subRevReq.ResourceAttributes.Resource
			filteredApiResourceList := filterResources(apiresourcesList, filterOnApigroup)
			klog.V(5).Infof("Filtered resources on group : %v, %d", filteredApiResourceList, len(finalFilteredResource))
			for _, filteredApiResource := range filteredApiResourceList {
				for _, resource := range filteredApiResource.APIResources {
					if resource.Name == filterResource && len(finalFilteredResource) != 1 {
						klog.V(5).Infof("group %s, version %s", resource.Group, resource.Version)
						singleApiResourceList := &metav1.APIResourceList{
							GroupVersion: filteredApiResource.GroupVersion,
							APIResources: []metav1.APIResource{
								resource,
							},
						}
						finalFilteredResource = append(finalFilteredResource, singleApiResourceList)
					}
				}
			}

			klog.V(5).Infof("Final filtered resource : %v", finalFilteredResource)

			authInfoList = createAuthorizationActionInfoList(finalFilteredResource, subRevReq.ResourceAttributes.Verb, clusterType)
		}
	} else if subRevReq.NonResourceAttributes != nil {
		authInfoSingle := AuthorizationActionInfo{
			IsDataAction: true,
		}
		authInfoSingle.AuthorizationEntity.Id = path.Join(clusterType, subRevReq.NonResourceAttributes.Path, getActionName(subRevReq.NonResourceAttributes.Verb))
		authInfoList = append(authInfoList, authInfoSingle)
	}
	return authInfoList
}

func filterResources(apiresourcesList []*metav1.APIResourceList, criteria func(*metav1.APIResourceList) bool) (filteredResources []*metav1.APIResourceList) {
	for _, res := range apiresourcesList {
		if criteria(res) {
			filteredResources = append(filteredResources, res)
		}
	}
	return
}

func createAuthorizationActionInfoList(apiresourceList []*metav1.APIResourceList, filterVerb string, clusterType string) []AuthorizationActionInfo {
	var authInfos []AuthorizationActionInfo
	for _, apiResList := range apiresourceList {
		group := ""
		if apiResList.GroupVersion != "" && apiResList.GroupVersion != "v1" {
			group = strings.Split(apiResList.GroupVersion, "/")[0]
		}

		for _, resource := range apiResList.APIResources {

			authInfo := AuthorizationActionInfo{
				IsDataAction: true,
			}

			authInfo.AuthorizationEntity.Id = clusterType

			if group != "" {
				authInfo.AuthorizationEntity.Id = path.Join(authInfo.AuthorizationEntity.Id, group)
			}

			resourceName := resource.Name
			subResourceName := ""
			if strings.Contains(resource.Name, "/") {
				resourceName = strings.Split(resource.Name, "/")[0]
				subResourceName = strings.Split(resource.Name, "/")[1]
			}

			if filterVerb != "*" {
				action := getResourceAndAction(resourceName, subResourceName, filterVerb)
				authInfo.AuthorizationEntity.Id = path.Join(authInfo.AuthorizationEntity.Id, action)
				found := searchInAuthInfo(authInfos, authInfo.AuthorizationEntity.Id)

				if found == -1 {
					authInfos = append(authInfos, authInfo)
				}
			} else {
				// create data actions for all the verbs
				for _, verb := range resource.Verbs {
					authInfoSingle := authInfo
					action := getResourceAndAction(resourceName, subResourceName, verb)
					authInfoSingle.AuthorizationEntity.Id = path.Join(authInfoSingle.AuthorizationEntity.Id, action)
					found := searchInAuthInfo(authInfos, authInfoSingle.AuthorizationEntity.Id)
					klog.V(5).Infof("found string: %v %s", found, authInfoSingle.AuthorizationEntity.Id)

					if found == -1 {
						authInfos = append(authInfos, authInfoSingle)
					}

				}
			}
		}
	}

	return authInfos
}

func searchInAuthInfo(authInfos []AuthorizationActionInfo, searchAction string) int {
	found := slices.IndexFunc(authInfos, func(a AuthorizationActionInfo) bool { return a.AuthorizationEntity.Id == searchAction })

	return found
}

func defaultDir(s string) string {
	if s != "" {
		return s
	}
	return "-" // invalid for a namespace
}

func getResultCacheKey(subRevReq *authzv1.SubjectAccessReviewSpec) string {
	cacheKey := subRevReq.User

	if subRevReq.ResourceAttributes != nil {
		cacheKey = path.Join(cacheKey, defaultDir(subRevReq.ResourceAttributes.Namespace))
		cacheKey = path.Join(cacheKey, defaultDir(subRevReq.ResourceAttributes.Group))
		action := getResourceAndAction(subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Subresource, subRevReq.ResourceAttributes.Verb)
		cacheKey = path.Join(cacheKey, action)
	} else if subRevReq.NonResourceAttributes != nil {
		cacheKey = path.Join(cacheKey, subRevReq.NonResourceAttributes.Path, getActionName(subRevReq.NonResourceAttributes.Verb))
	}

	return cacheKey
}

func prepareCheckAccessRequestBody(req *authzv1.SubjectAccessReviewSpec, clusterType string, apiresourcesList []*metav1.APIResourceList, resourceId string, useNamespaceResourceScopeFormat bool) (*CheckAccessRequest, error) {
	/* This is how sample SubjectAccessReview request will look like
		{
			"kind": "SubjectAccessReview",
		    	"apiVersion": "authorization.k8s.io/v1beta1",
		    	"metadata": {
		        	"creationTimestamp": null
		    	},
		    	"spec": {
		        	"resourceAttributes": {
		            		"namespace": "default",
			            	"verb": "get",
					"group": "extensions",
					"version": "v1beta1",
					"resource": "deployments",
					"name": "obo-deploy"
		        	},
				"user": "user@contoso.com",
				"extra": {
					"oid": [
		    			"62103f2e-051d-48cc-af47-b1ff3deec630"
				]
		        	}
		    	},
		    	"status": {
		        	"allowed": false
		    	}
		}

		For check access it will be converted into following request for arc cluster:
		{
			"Subject": {
				"Attributes": {
	                                "ObjectId": "62103f2e-051d-48cc-af47-b1ff3deec630"
				}
			},
			"Actions": [
				{
					"Id": "Microsoft.Kubernetes/connectedClusters/extensions/deployments/read",
					"IsDataAction": true
				}
			],
			"Resource": {
				"Id": "<resourceId>/namespaces/<namespace name>"
			}
		}
	*/
	checkaccessreq := CheckAccessRequest{}
	var userOid string
	if oid, ok := req.Extra["oid"]; ok {
		val := oid.String()
		userOid = val[1 : len(val)-1]
	} else {
		return nil, errors.New("oid info not sent from authentication module")
	}

	if isValidUUID(userOid) {
		checkaccessreq.Subject.Attributes.ObjectId = userOid
	} else {
		return nil, errors.New("oid info sent from authentication module is not valid")
	}

	groups := getValidSecurityGroups(req.Groups)
	checkaccessreq.Subject.Attributes.Groups = groups

	username = req.User
	var actions []AuthorizationActionInfo
	actions = getDataActions(req, clusterType, apiresourcesList)
	checkaccessreq.Actions = actions
	checkaccessreq.Resource.Id = getScope(resourceId, req.ResourceAttributes, useNamespaceResourceScopeFormat)

	return &checkaccessreq, nil
}

func getNameSpaceScope(req *authzv1.SubjectAccessReviewSpec, useNamespaceResourceScopeFormat bool) (bool, string) {
	var namespace string = ""
	if req.ResourceAttributes != nil && req.ResourceAttributes.Namespace != "" {
		if useNamespaceResourceScopeFormat {
			namespace = path.Join(NamespaceResourceFormat, req.ResourceAttributes.Namespace)
		} else {
			namespace = path.Join(namespaces, req.ResourceAttributes.Namespace)
		}
		return true, namespace
	}
	return false, namespace
}

func ConvertCheckAccessResponse(body []byte) (*authzv1.SubjectAccessReviewStatus, error) {
	var (
		response []AuthorizationDecision
		allowed  bool
		denied   bool
		verdict  string
	)

	err := json.Unmarshal(body, &response)
	if err != nil {
		klog.V(10).Infof("Failed to parse checkacccess response. Error:%s", err.Error())
		return nil, errors.Wrap(err, "Error in unmarshalling check access response.")
	}

	if strings.ToLower(response[0].Decision) == Allowed {
		allowed = true
		verdict = fmt.Sprintf(AccessAllowedVerboseVerdict, response[0].AzureRoleAssignment.Id, response[0].AzureRoleAssignment.RoleDefinitionId, username)
	} else {
		allowed = false
		denied = true
		verdict = AccessNotAllowedVerdict
	}

	return &authzv1.SubjectAccessReviewStatus{Allowed: allowed, Reason: verdict, Denied: denied}, nil
}
