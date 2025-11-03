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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	azureutils "go.kubeguard.dev/guard/util/azure"
	errutils "go.kubeguard.dev/guard/util/error"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/klog/v2"
)

const (
	ActionBatchCount            = 200
	AccessAllowedVerdict        = "Access allowed by Azure RBAC"
	AccessAllowedVerboseVerdict = "Access allowed by Azure RBAC Role Assignment %s of Role %s to user %s"
	Allowed                     = "allowed"
	AccessNotAllowedVerdict     = "User does not have access to the resource in Azure. Update role assignment to allow access."
	NamespaceResourceFormat     = "/providers/Microsoft.KubernetesConfiguration/namespaces"
	namespaces                  = "namespaces"
	managedNamespaces           = "managedNamespaces"
	NoOpinionVerdict            = "Azure does not have opinion for this user."
	NonAADUserNoOpVerdict       = "Azure does not have opinion for this non AAD user. If you are an AAD user, please set Extra:oid parameter for impersonated user in the kubeconfig"
	NonAADUserNotAllowedVerdict = "Access denied by Azure RBAC for non AAD users. Configure --azure.skip-authz-for-non-aad-users to enable access. If you are an AAD user, please set Extra:oid parameter for impersonated user in the kubeconfig."
	PodsResource                = "pods"
	CustomResources             = "customresources"
	ReadVerb                    = "read"
	WriteVerb                   = "write"
	DeleteVerb                  = "delete"
)

var getStoredOperationsMap = azureutils.DeepCopyOperationsMap

type SubjectInfoAttributes struct {
	ObjectId string   `json:"ObjectId"`
	Groups   []string `json:"Groups,omitempty"`
}

type SubjectInfo struct {
	Attributes SubjectInfoAttributes `json:"Attributes"`
}

type CheckAccessRequest struct {
	Subject  SubjectInfo                          `json:"Subject"`
	Actions  []azureutils.AuthorizationActionInfo `json:"Actions"`
	Resource azureutils.AuthorizationEntity       `json:"Resource"`
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

const SubresourceAttrName = "Microsoft.ContainerService/managedClusters/resources:subresource"

// Allow list for subresources to include as attributes.
var subresourceAttributeAllowlist = map[string]struct{}{
	"pods/logs":                {},
	"pods/exec":                {},
	"pods/portforward":         {},
	"pods/proxy":               {},
	"pods/ephemeralcontainers": {},
	"pods/attach":              {},
	"deployments/scale":        {},
	"statefulsets/scale":       {},
	"replicasets/scale":        {},
}

func getCustomResourceOperationsMap(clusterType string) map[string]azureutils.AuthorizationActionInfo {
	return map[string]azureutils.AuthorizationActionInfo{
		ReadVerb: {
			IsDataAction: true,
			AuthorizationEntity: azureutils.AuthorizationEntity{
				Id: path.Join(clusterType, CustomResources, ReadVerb),
			},
		},
		WriteVerb: {
			IsDataAction: true,
			AuthorizationEntity: azureutils.AuthorizationEntity{
				Id: path.Join(clusterType, CustomResources, WriteVerb),
			},
		},
		DeleteVerb: {
			IsDataAction: true,
			AuthorizationEntity: azureutils.AuthorizationEntity{
				Id: path.Join(clusterType, CustomResources, DeleteVerb),
			},
		},
	}
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

	case "*": // map * to * for wildcard scenario
		return "*"
	default:
		return ""
	}
}

func getResourceAndAction(resource string, subResource string, verb string) string {
	var action string

	if resource == PodsResource && subResource == "exec" {
		action = path.Join(resource, subResource, "action")
	} else {
		action = path.Join(resource, getActionName(verb))
	}

	return action
}

func getDataActions(ctx context.Context, subRevReq *authzv1.SubjectAccessReviewSpec, clusterType string, allowCustomResourceTypeCheck bool, allowSubresourceTypeCheck bool) ([]azureutils.AuthorizationActionInfo, error) {
	var authInfoList []azureutils.AuthorizationActionInfo
	var err error

	if subRevReq.ResourceAttributes != nil {
		storedOperationsMap := getStoredOperationsMap()

		// Create logger with feature context for custom resource type check
		featureLog := klog.FromContext(ctx).WithValues("feature", "CustomResourceTypeCheck")
		featureCtx := klog.NewContext(ctx, featureLog)

		isCustomerResourceTypeCheckAvailable := allowCustomResourceTypeCheck && storedOperationsMap != nil && len(storedOperationsMap) != 0
		if !isCustomerResourceTypeCheckAvailable {
			featureLog.V(5).Info("Feature is not available", "allowCustomResourceTypeCheck", allowCustomResourceTypeCheck, "operationsMapAvailable", storedOperationsMap != nil && len(storedOperationsMap) != 0)
		}

		if subRevReq.ResourceAttributes.Resource != "*" && subRevReq.ResourceAttributes.Group != "*" && subRevReq.ResourceAttributes.Verb != "*" {
			/*
				This sections handles the following scenarios:

				| Scenario                      | Namespace is empty (Cluster scope call)      | Namespace is not empty (NS scope)        |
				------------------------------- ---------------------------------------------- ------------------------------------------
				| Verb, Res and Group are not * | Normal single resource call at cluster scope | Normal single resource call  at ns scope |

			*/
			authInfoSingle := azureutils.AuthorizationActionInfo{
				IsDataAction: true,
			}

			if isCustomerResourceTypeCheckAvailable && isCustomResourceAccessRequest(subRevReq, storedOperationsMap) {
				/*
					In this case both Res and Group are not *, but there is no matching DataAction present on the storedOperationsMap.
					The resource is presumed to be a CR and <clusterType>/customresources/<action> DataAction will be used for check access.
				*/
				featureLog.V(5).Info("Resource identified as custom resource")
				return getAuthInfoListForCustomResource(featureCtx, subRevReq, clusterType)
			}

			authInfoSingle.AuthorizationEntity.Id = clusterType

			if subRevReq.ResourceAttributes.Group != "" {
				authInfoSingle.AuthorizationEntity.Id = path.Join(authInfoSingle.AuthorizationEntity.Id, subRevReq.ResourceAttributes.Group)
			}

			action := getResourceAndAction(subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Subresource, subRevReq.ResourceAttributes.Verb)
			authInfoSingle.AuthorizationEntity.Id = path.Join(authInfoSingle.AuthorizationEntity.Id, action)
			if allowSubresourceTypeCheck {
				err = setAuthInfoSubresourceAttributes(&authInfoSingle, subRevReq)
				if err != nil {
					return nil, errors.Errorf("Error while setting subresource attributes: %s", err.Error())
				}
			}
			authInfoList = append(authInfoList, authInfoSingle)

		} else {
			if storedOperationsMap == nil || (storedOperationsMap != nil && len(storedOperationsMap) == 0) {
				return nil, errors.Errorf("Wildcard support for Resource/Verb/Group is not enabled for request Group: %s, Resource: %s, Verb: %s", subRevReq.ResourceAttributes.Group, subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Verb)
			}

			authInfoList, err = getAuthInfoListForWildcard(featureCtx, subRevReq, storedOperationsMap, clusterType, isCustomerResourceTypeCheckAvailable, allowSubresourceTypeCheck)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("Error which creating actions for checkaccess for Group: %s, Resource: %s, Verb: %s", subRevReq.ResourceAttributes.Group, subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Verb))
			}

		}
	} else if subRevReq.NonResourceAttributes != nil {
		authInfoSingle := azureutils.AuthorizationActionInfo{
			IsDataAction: true,
		}
		authInfoSingle.AuthorizationEntity.Id = path.Join(clusterType, subRevReq.NonResourceAttributes.Path, getActionName(subRevReq.NonResourceAttributes.Verb))
		authInfoList = append(authInfoList, authInfoSingle)
	}
	return authInfoList, nil
}

func getAuthInfoListForWildcard(ctx context.Context, subRevReq *authzv1.SubjectAccessReviewSpec, storedOperationsMap azureutils.OperationsMap, clusterType string, isCustomerResourceTypeCheckAvailable bool, allowSubresourceTypeCheck bool) ([]azureutils.AuthorizationActionInfo, error) {
	var authInfoList []azureutils.AuthorizationActionInfo
	var err error
	finalFilteredOperations := azureutils.NewOperationsMap()
	if subRevReq.ResourceAttributes.Resource == "*" {
		/*
			This sections handles the following scenarios:

			| Scenario              | Namespace is empty (Cluster scope call)                    | Namespace is not empty (NS scope)             |
			------------------------ ------------------------------------------------------------  ----------------------------------------------
			| Verb-*, Res-*, Group-*| All cluster and ns res with all verbs at clusterscope | All ns resources at ns scope |

			| Res-*, Group-*        | All cluster and ns res with specified verb at clusterscope | All ns res with specified verb at ns scope    |

			| Verb-*, Res-*         | All cluster and ns res with all verbs under                | All ns res with all verbs under specified     |
			|                       | specified apigroup at clusterscope                         | apigroup at nsscope                           |

			| Resource - *          | All CS and NS Resources under specifed apigroup with       | All NS Resources under specifed apigroup with |
			|                       | specified verb at cluster scope                            | specified verb at ns scope                    |
		*/
		filteredOperations := azureutils.NewOperationsMap()
		if subRevReq.ResourceAttributes.Group == "*" {
			// all resources under all apigroups
			filteredOperations = storedOperationsMap
		} else if isCustomerResourceTypeCheckAvailable && isCustomResourceAccessRequest(subRevReq, storedOperationsMap) {
			/*
				In this case Group is not *, but there are no matching DataActions present on the storedOperationsMap.
				The resource is presumed to be a CR and <clusterType>/customresources/<action> DataAction will be used for check access.
			*/
			featureLog.V(5).Info("Resource identified as custom resource (wildcard resource)")
			return getAuthInfoListForCustomResource(featureCtx, subRevReq, clusterType)
		} else if subRevReq.ResourceAttributes.Group != "" {
			// all resources under specified apigroup
			if value, found := storedOperationsMap[subRevReq.ResourceAttributes.Group]; found {
				filteredOperations[subRevReq.ResourceAttributes.Group] = value
			} else {
				return nil, errors.Errorf("No resources found for group %s", subRevReq.ResourceAttributes.Group)
			}
		} else {
			// if Group is not there that means it is the core apigroup
			if value, found := storedOperationsMap["v1"]; found {
				filteredOperations["v1"] = value
			} else {
				return nil, errors.New("No resources found for the core group")
			}
		}

		// if Namespace is populated, we need to create the list only for namespace scoped resources
		if subRevReq.ResourceAttributes.Namespace != "" {
			for group, resMap := range filteredOperations {
				for resourceName, verbValues := range resMap {
					for _, dataAction := range verbValues {
						if dataAction.IsNamespacedResource {
							finalFilteredOperations = initializeMapForGroupAndResource(finalFilteredOperations, group, resourceName)
							finalFilteredOperations[group][resourceName] = verbValues
							break
						}
					}
				}
			}
		} else {
			// both cluster scoped and namespace scoped resource list
			finalFilteredOperations = filteredOperations
		}
	} else {
		/*
			   This sections handles the following scenarios:

				| Scenario        | Namespace is empty (Cluster scope call)          | Namespace is not empty (NS scope)                    |
				------------------ --------------------------------------------------  -----------------------------------------------------
				| Verb-*, Group-* | Resource under all apigroups and with all verbs  | Resource under specifed apigroups and with all verbs |
								  | at clusterscope                                  | at ns scope                                          |

				| Verb - *        | Resource under specifed apigroups and with all   | Resource under specifed apigroups and with all verbs |
								  | verbs at cluster scope                           | at ns scope                                          |

				| Group - *       | Resource under all apigroups with specified verb | Resource under all apigroups with specified verb     |
				|                 | at cluster scope                                 |  at ns scope                                         |
		*/
		if subRevReq.ResourceAttributes.Group == "*" {
			// #1 and #3
			for group, resMap := range storedOperationsMap {
				if verbMap, found := resMap[subRevReq.ResourceAttributes.Resource]; found {
					finalFilteredOperations = initializeMapForGroupAndResource(finalFilteredOperations, group, subRevReq.ResourceAttributes.Resource)
					finalFilteredOperations[group][subRevReq.ResourceAttributes.Resource] = verbMap
				}
			}
		} else if isCustomerResourceTypeCheckAvailable && isCustomResourceAccessRequest(subRevReq, storedOperationsMap) {
			/*
				In this case both Res and Group are not *, but there are no matching DataActions present on the storedOperationsMap.
				The resource is presumed to be a CR and <clusterType>/customresources/<action> DataAction will be used for check access.
			*/
			featureLog.V(5).Info("Resource identified as custom resource (wildcard verb)")
			return getAuthInfoListForCustomResource(featureCtx, subRevReq, clusterType)
		} else { // #2
			group := "v1" // core api group key
			if subRevReq.ResourceAttributes.Group != "" {
				group = subRevReq.ResourceAttributes.Group
			}

			if resMap, found := storedOperationsMap[group]; found {
				if verbMap, found := resMap[subRevReq.ResourceAttributes.Resource]; found {
					finalFilteredOperations = initializeMapForGroupAndResource(finalFilteredOperations, group, subRevReq.ResourceAttributes.Resource)
					finalFilteredOperations[group][subRevReq.ResourceAttributes.Resource] = verbMap
				}
			} else {
				return nil, errors.Errorf("No resources found for group %s and resource %s", subRevReq.ResourceAttributes.Group, subRevReq.ResourceAttributes.Resource)
			}

		}
	}

	klog.V(7).Infof("List of filtered operations: %s", finalFilteredOperations)

	// create list of Data Actions
	authInfoList, err = createAuthorizationActionInfoList(finalFilteredOperations, subRevReq.ResourceAttributes.Verb)
	if err != nil {
		return nil, err
	}

	if allowSubresourceTypeCheck {
		for i := range authInfoList {
			err = setAuthInfoSubresourceAttributes(&authInfoList[i], subRevReq)
			if err != nil {
				return nil, errors.Errorf("Error while setting subresource attributes: %s", err.Error())
			}
		}
	}
	return authInfoList, nil
}

func getAuthInfoListForCustomResource(ctx context.Context, subRevReq *authzv1.SubjectAccessReviewSpec, clusterType string) ([]azureutils.AuthorizationActionInfo, error) {
	log := klog.FromContext(ctx)
	var authInfoList []azureutils.AuthorizationActionInfo
	if subRevReq.ResourceAttributes.Verb == "*" {
		log.V(7).Info("Creating actions for all custom resource verbs")
		for _, action := range getCustomResourceOperationsMap(clusterType) {
			authInfoList = append(authInfoList, action)
		}
	} else {
		action := getActionName(subRevReq.ResourceAttributes.Verb)
		authInfoSingle, found := getCustomResourceOperationsMap(clusterType)[action]
		if !found {
			return nil, errors.Errorf("No actions found for verb: %s, action: %s", subRevReq.ResourceAttributes.Verb, action)
		}
		log.V(7).Info("Creating action for custom resource", "action", action)
		authInfoList = append(authInfoList, authInfoSingle)
	}

	for i := range authInfoList {
		err := setAuthInfoResourceAttributes(&authInfoList[i], subRevReq)
		if err != nil {
			return nil, errors.Errorf("Error while setting resource attributes: %s", err.Error())
		}
	}

	return authInfoList, nil
}

func setAuthInfoResourceAttributes(action *azureutils.AuthorizationActionInfo, subRevReq *authzv1.SubjectAccessReviewSpec) error {
	if subRevReq.ResourceAttributes == nil {
		return errors.New("Resource attributes are empty")
	}
	if subRevReq.ResourceAttributes.Group == "" {
		return errors.New("Group is empty")
	}
	action.Attributes = make(map[string]string)
	action.Attributes["Microsoft.ContainerService/managedClusters/customResources:group"] = subRevReq.ResourceAttributes.Group
	if subRevReq.ResourceAttributes.Resource != "" {
		action.Attributes["Microsoft.ContainerService/managedClusters/customResources:kind"] = subRevReq.ResourceAttributes.Resource
	}
	return nil
}

func shouldHandleSubresource(resource string, subresource string) bool {
	_, shouldHandle := subresourceAttributeAllowlist[resource+"/"+subresource]
	return shouldHandle
}

func setAuthInfoSubresourceAttributes(action *azureutils.AuthorizationActionInfo, subRevReq *authzv1.SubjectAccessReviewSpec) error {
	if subRevReq.ResourceAttributes == nil {
		return errors.New("Resource attributes are empty")
	}

	if shouldHandleSubresource(subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Subresource) {
		if action.Attributes == nil {
			action.Attributes = make(map[string]string)
		}
		action.Attributes[SubresourceAttrName] = subRevReq.ResourceAttributes.Subresource
	}
	return nil
}

func isCustomResourceAccessRequest(subRevReq *authzv1.SubjectAccessReviewSpec, operationsMap azureutils.OperationsMap) bool {
	if subRevReq.ResourceAttributes == nil {
		return false
	}
	if subRevReq.ResourceAttributes.Group == "" || subRevReq.ResourceAttributes.Group == "*" {
		return false
	}
	if subRevReq.ResourceAttributes.Resource == "" || subRevReq.ResourceAttributes.Resource == "*" {
		_, found := operationsMap[subRevReq.ResourceAttributes.Group]
		return !found
	}

	_, found := operationsMap[subRevReq.ResourceAttributes.Group][subRevReq.ResourceAttributes.Resource]
	return !found
}

func initializeMapForGroupAndResource(filteredOperations azureutils.OperationsMap, group string, resourceName string) azureutils.OperationsMap {
	if _, found := filteredOperations[group]; !found {
		filteredOperations[group] = azureutils.NewResourceAndVerbMap()
	}

	if _, found := filteredOperations[group][resourceName]; !found {
		filteredOperations[group][resourceName] = azureutils.NewVerbAndActionsMap()
	}

	return filteredOperations
}

func createAuthorizationActionInfoList(filteredOperations azureutils.OperationsMap, filterVerb string) ([]azureutils.AuthorizationActionInfo, error) {
	if len(filteredOperations) == 0 {
		return nil, errors.New("No operations were found for the request.")
	}

	var authInfos []azureutils.AuthorizationActionInfo

	if filterVerb != "*" {
		verb := getActionName(filterVerb)
		for _, resMap := range filteredOperations {
			for _, verbsMap := range resMap {
				if dataAction, found := verbsMap[verb]; found {
					authInfos = append(authInfos, dataAction.ActionInfo)
				}
			}
		}

		if len(authInfos) == 0 {
			return nil, errors.Errorf("No operations were found for the verb: %s.", filterVerb)
		}
	} else {
		for _, resMap := range filteredOperations {
			for _, verbsMap := range resMap {
				for _, dataAction := range verbsMap {
					authInfos = append(authInfos, dataAction.ActionInfo)
				}
			}
		}
	}

	if klog.V(5).Enabled() {
		printAuthInfos, _ := json.Marshal(authInfos)

		klog.Infof("List of authorization action info for checkaccess: %s", string(printAuthInfos))
	}

	return authInfos, nil
}

func defaultDir(s string) string {
	if s != "" {
		return s
	}
	return "-" // invalid for a namespace
}

func getResultCacheKey(subRevReq *authzv1.SubjectAccessReviewSpec, allowSubresourceTypeCheck bool) string {
	cacheKey := subRevReq.User

	if subRevReq.ResourceAttributes != nil {
		cacheKey = path.Join(cacheKey, defaultDir(subRevReq.ResourceAttributes.Namespace))
		cacheKey = path.Join(cacheKey, defaultDir(subRevReq.ResourceAttributes.Group))
		action := getResourceAndAction(subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Subresource, subRevReq.ResourceAttributes.Verb)
		cacheKey = path.Join(cacheKey, action)

		// Cache results for subresources of interest separately
		if allowSubresourceTypeCheck {
			if shouldHandleSubresource(subRevReq.ResourceAttributes.Resource, subRevReq.ResourceAttributes.Subresource) {
				cacheKey = path.Join(cacheKey, subRevReq.ResourceAttributes.Subresource)
			}
		}
	} else if subRevReq.NonResourceAttributes != nil {
		cacheKey = path.Join(cacheKey, subRevReq.NonResourceAttributes.Path, getActionName(subRevReq.NonResourceAttributes.Verb))
	}

	return cacheKey
}

func prepareCheckAccessRequestBody(ctx context.Context, req *authzv1.SubjectAccessReviewSpec, clusterType string, resourceId string, useNamespaceResourceScopeFormat bool, allowCustomResourceTypeCheck bool, allowSubresourceTypeCheck bool) ([]*CheckAccessRequest, error) {
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
					"Attributes": {
						"Microsoft.ContainerService/managedClusters/customResources:kind": "SecretProviderClass",
						"Microsoft.ContainerService/managedClusters/customResources:group": "extensions",
					}
				}
			],
			"Resource": {
				"Id": "<resourceId>/namespaces/<namespace name>"
			}
		}
	*/

	var userOid string
	if oid, ok := req.Extra["oid"]; ok {
		val := oid.String()
		userOid = val[1 : len(val)-1]
		if !isValidUUID(userOid) {
			return nil, errutils.WithCode(errors.New("oid info sent from authentication module is not valid"), http.StatusBadRequest)
		}
	} else {
		return nil, errutils.WithCode(errors.New("oid info not sent from authentication module"), http.StatusBadRequest)
	}
	groups := getValidSecurityGroups(req.Groups)
	actions, err := getDataActions(ctx, req, clusterType, allowCustomResourceTypeCheck, allowSubresourceTypeCheck)
	if err != nil {
		return nil, errutils.WithCode(errors.Wrap(err, "Error while creating list of dataactions for check access call"), http.StatusInternalServerError)
	}
	var checkAccessReqs []*CheckAccessRequest
	for i := 0; i < len(actions); i += ActionBatchCount {
		j := i + ActionBatchCount
		if j > len(actions) {
			j = len(actions)
		}

		checkaccessreq := CheckAccessRequest{}
		checkaccessreq.Subject.Attributes.Groups = groups
		checkaccessreq.Subject.Attributes.ObjectId = userOid
		checkaccessreq.Actions = actions[i:j]
		checkaccessreq.Resource.Id = getScope(resourceId, req.ResourceAttributes, useNamespaceResourceScopeFormat)
		checkAccessReqs = append(checkAccessReqs, &checkaccessreq)
	}

	return checkAccessReqs, nil
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

func getManagedNameSpaceScope(req *authzv1.SubjectAccessReviewSpec) (bool, string) {
	var namespace string = ""
	if req.ResourceAttributes != nil && req.ResourceAttributes.Namespace != "" {
		namespace = path.Join(managedNamespaces, req.ResourceAttributes.Namespace)
		return true, namespace
	}
	return false, namespace
}

func ConvertCheckAccessResponse(username string, body []byte) (*authzv1.SubjectAccessReviewStatus, error) {
	var (
		response []AuthorizationDecision
		allowed  bool
		denied   bool
		verdict  string
	)

	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, errutils.WithCode(errors.Wrap(err, "Error in unmarshalling check access response."), http.StatusInternalServerError)
	}

	deniedResultFound := slices.IndexFunc(response, func(a AuthorizationDecision) bool { return strings.ToLower(a.Decision) != Allowed })

	if deniedResultFound == -1 { // no denied result found
		allowed = true
		verdict = fmt.Sprintf(AccessAllowedVerboseVerdict, response[0].AzureRoleAssignment.Id, response[0].AzureRoleAssignment.RoleDefinitionId, username)
	} else {
		allowed = false
		denied = true
		verdict = AccessNotAllowedVerdict
	}

	return &authzv1.SubjectAccessReviewStatus{Allowed: allowed, Reason: verdict, Denied: denied}, nil
}

func defaultNotFoundDecision() *authzv1.SubjectAccessReviewStatus {
	return &authzv1.SubjectAccessReviewStatus{
		Allowed: false,
		Denied:  true,
		Reason:  AccessNotAllowedVerdict,
	}
}

// buildCheckAccessURL constructs the Azure check access URL in string form.
//
// The input parameters hold the following invariants:
//
// 1. base: the scheme must be https, and the host should be the Azure management API endpoint.
// 2. azureResourceID: the parent Azure resource ID for the check access request, must not be empty
//
// The returned URL holds the following invariants:
//
// 3. the host and scheme of the URL are the same as the base URL
// 4. the path should has the prefix of azureResourceID value
// 5. the path should have the suffix of the checkAccessPath
//
// Any invariant violation will result in an error being returned.
func buildCheckAccessURL(
	base url.URL,
	azureResourceID string,
	hasNamespace bool,
	namespacePath string,
) (string, error) {
	rv := base // shallow copy of the base URL to allow checking before returning

	rv.Path = path.Join(rv.Path, azureResourceID)
	azureResourceIDPrefix := rv.Path
	if hasNamespace {
		rv.Path = path.Join(rv.Path, namespacePath)
	}
	rv.Path = path.Join(rv.Path, checkAccessPath)

	params := url.Values{}
	params.Add(queryParamAPIVersion, checkAccessAPIVersion)
	rv.RawQuery = params.Encode()

	// invariant checks

	// invariant 1
	if !strings.EqualFold(rv.Scheme, "https") {
		return "", fmt.Errorf("invalid scheme %q, expected https", rv.Scheme)
	}
	// invariant 2
	if azureResourceID == "" {
		return "", fmt.Errorf("azureResourceID must not be empty")
	}
	// invariant 3
	if rv.Host != base.Host || rv.Scheme != base.Scheme {
		err := fmt.Errorf(
			"invalid URL, expected host %q and scheme %q, got host %q and scheme %q",
			base.Host, base.Scheme, rv.Host, rv.Scheme,
		)
		return "", err
	}
	// invariant 4
	if !strings.HasPrefix(strings.TrimPrefix(rv.Path, "/"), strings.TrimPrefix(azureResourceIDPrefix, "/")) {
		err := fmt.Errorf(
			"invalid URL path %q, expected to start with %q",
			rv.Path, azureResourceIDPrefix,
		)
		return "", err
	}
	// invariant 5
	if !strings.HasSuffix(rv.Path, checkAccessPath) {
		err := fmt.Errorf(
			"invalid URL path %q, expected to end with %q",
			rv.Path, checkAccessPath,
		)
		return "", err
	}

	return rv.String(), nil
}
