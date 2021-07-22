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
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/klog/v2"
)

const (
	AccessAllowedVerdict        = "Access allowed by Azure RBAC"
	AccessAllowedVerboseVerdict = "Access allowed by Azure RBAC Role Assignment %s of Role %s to user %s"
	Allowed                     = "allowed"
	AccessNotAllowedVerdict     = "User does not have access to the resource in Azure. Update role assignment to allow access."
	namespaces                  = "namespaces"
	NoOpinionVerdict            = "Azure does not have opinion for this user."
	NonAADUserNoOpVerdict       = "Azure does not have opinion for this non AAD user. If you are an AAD user, please set Extra:oid parameter for impersonated user in the kubeconfig"
	NonAADUserNotAllowedVerdict = "Access denied by Azure RBAC for non AAD users. Configure --azure.skip-authz-for-non-aad-users to enable access. If you are an AAD user, please set Extra:oid parameter for impersonated user in the kubeconfig."
)

var (
	username string
)

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

func getScope(resourceId string, attr *authzv1.ResourceAttributes) string {
	if attr != nil && attr.Namespace != "" {
		return path.Join(resourceId, namespaces, attr.Namespace)
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
		fallthrough //instead of action create will be mapped to write
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

func getResourceAndAction(subRevReq *authzv1.SubjectAccessReviewSpec) string {
	action := subRevReq.ResourceAttributes.Resource
	if subRevReq.ResourceAttributes.Resource == "pods" &&  subRevReq.ResourceAttributes.Subresource == "exec" {
		action = path.Join(subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Subresource, "action")
	} else {
		action = path.Join(subRevReq.ResourceAttributes.Resource, getActionName(subRevReq.ResourceAttributes.Verb))
    }

	return action
}

func getDataAction(subRevReq *authzv1.SubjectAccessReviewSpec, clusterType string) AuthorizationActionInfo {
	authInfo := AuthorizationActionInfo{
		IsDataAction: true}

	authInfo.AuthorizationEntity.Id = clusterType
	if subRevReq.ResourceAttributes != nil {
		if subRevReq.ResourceAttributes.Group != "" {
            authInfo.AuthorizationEntity.Id = path.Join(authInfo.AuthorizationEntity.Id, subRevReq.ResourceAttributes.Group)
        }
		
		action := getResourceAndAction(subRevReq)
		authInfo.AuthorizationEntity.Id = path.Join(authInfo.AuthorizationEntity.Id, action)
	} else if subRevReq.NonResourceAttributes != nil {
		authInfo.AuthorizationEntity.Id = path.Join(authInfo.AuthorizationEntity.Id, subRevReq.NonResourceAttributes.Path, getActionName(subRevReq.NonResourceAttributes.Verb))
	}
	return authInfo
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
		cacheKey = path.Join(cacheKey, subRevReq.ResourceAttributes.Resource, getActionName(subRevReq.ResourceAttributes.Verb))
	} else if subRevReq.NonResourceAttributes != nil {
		cacheKey = path.Join(cacheKey, subRevReq.NonResourceAttributes.Path, getActionName(subRevReq.NonResourceAttributes.Verb))
	}

	return cacheKey
}

func prepareCheckAccessRequestBody(req *authzv1.SubjectAccessReviewSpec, clusterType, resourceId string) (*CheckAccessRequest, error) {
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
	action := make([]AuthorizationActionInfo, 1)
	action[0] = getDataAction(req, clusterType)
	checkaccessreq.Actions = action
	checkaccessreq.Resource.Id = getScope(resourceId, req.ResourceAttributes)

	return &checkaccessreq, nil
}

func getNameSpaceScope(req *authzv1.SubjectAccessReviewSpec) (bool, string) {
	var namespace string = ""
	if req.ResourceAttributes != nil && req.ResourceAttributes.Namespace != "" {
		namespace = path.Join(namespaces, req.ResourceAttributes.Namespace)
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
