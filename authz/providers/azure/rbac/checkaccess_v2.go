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

/*
Package rbac implements Azure RBAC authorization using CheckAccess v2 API.

This file contains the CheckAccess v2 API implementation, which uses the official
Azure checkaccess-v2-go-sdk to perform RBAC authorization checks. The v2 implementation
provides the following improvements over v1:

- Uses official Azure SDK instead of direct HTTP calls
- Better error handling and structured responses
- Improved logging with correlation IDs
- Consistent Prometheus metrics with v1

The implementation maintains the same batching behavior (200 actions per request) and
multi-level fallback logic (primary -> managed namespace -> fleet) as the v1 API.

To enable v2 API:
  --azure.use-checkaccess-v2=true
  --azure.pdp-endpoint=<PDP endpoint URL>

For more information on CheckAccess v2, see:
https://github.com/Azure/checkaccess-v2-go-sdk
*/

package rbac

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	azureutils "go.kubeguard.dev/guard/util/azure"
	errutils "go.kubeguard.dev/guard/util/error"

	checkaccess "github.com/Azure/checkaccess-v2-go-sdk/client"
	"github.com/google/uuid"
	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/klog/v2"
)

// performCheckAccessV2 performs authorization check using CheckAccess v2 SDK.
// It handles batching (200 actions per request) and executes requests in parallel.
//
// Unlike the SDK's CreateAuthorizationRequest which extracts oid/groups from JWT claims,
// this function accepts oid and groups directly from the SubjectAccessReviewSpec.
// This is necessary for AKS scenario where the JWT token doesn't contain user claims.
// Reference: https://github.com/kubeguard/guard/blob/master/authz/providers/azure/rbac/checkaccessreqhelper.go#L674-L684
func (a *AccessInfo) performCheckAccessV2(
	ctx context.Context,
	resourceId string,
	actions []string,
	userOid string,
	groups []string,
) (*authzv1.SubjectAccessReviewStatus, error) {
	log := klog.FromContext(ctx)

	// Batch actions into chunks of 200 (same as v1)
	const batchSize = ActionBatchCount
	var allDecisions []checkaccess.AuthorizationDecision

	for i := 0; i < len(actions); i += batchSize {
		batchIndex := i / batchSize
		end := min(i+batchSize, len(actions))

		batchActions := actions[i:end]
		correlationID := uuid.New().String()
		batchLog := log.WithValues("correlationID", correlationID, "batchIndex", batchIndex, "actionsCount", len(batchActions))
		batchCtx := klog.NewContext(ctx, batchLog)

		batchLog.V(7).Info("Starting CheckAccess v2 batch")
		start := time.Now()

		// Build authorization request manually instead of using SDK's CreateAuthorizationRequest.
		// The SDK helper extracts oid/groups from JWT claims, but for AKS the user identity
		// comes from SubjectAccessReviewSpec.Extra["oid"] and SubjectAccessReviewSpec.Groups.
		authzReq := buildAuthorizationRequestV2(resourceId, batchActions, userOid, groups)

		// Perform checkaccess call
		resp, err := a.pdpClient.CheckAccess(batchCtx, authzReq)
		duration := time.Since(start).Seconds()

		if err != nil {
			batchLog.Error(err, "CheckAccess v2 request failed", "durationSeconds", duration)
			// Use HTTP 500 to represent SDK errors, consistent with v1's internal server error pattern
			statusCode := azureutils.ConvertIntToString(http.StatusInternalServerError)
			checkAccessTotal.WithLabelValues(statusCode).Inc()
			checkAccessFailed.WithLabelValues(statusCode).Inc()
			checkAccessDuration.WithLabelValues(statusCode).Observe(duration)
			return nil, fmt.Errorf("CheckAccess v2 batch failed (batchIndex: %d, durationSeconds: %.2f): %w", batchIndex, duration, err)
		}

		batchLog.V(5).Info("CheckAccess v2 request succeeded", "durationSeconds", duration, "decisionsCount", len(resp.Value))
		// Use HTTP 200 to represent successful SDK calls, consistent with v1's HTTP 200 OK pattern
		statusCode := azureutils.ConvertIntToString(http.StatusOK)
		checkAccessTotal.WithLabelValues(statusCode).Inc()
		checkAccessSucceeded.Inc()
		checkAccessDuration.WithLabelValues(statusCode).Observe(duration)

		allDecisions = append(allDecisions, resp.Value...)
	}

	// Convert v2 response to v1 status
	return convertV2ResponseToStatus(ctx, allDecisions), nil
}

// buildAuthorizationRequestV2 constructs a CheckAccess v2 AuthorizationRequest.
// This replaces the SDK's CreateAuthorizationRequest which expects JWT claims.
// For AKS/Guard, the user identity (oid, groups) comes from SubjectAccessReviewSpec.
func buildAuthorizationRequestV2(resourceId string, actions []string, userOid string, groups []string) checkaccess.AuthorizationRequest {
	// Build action info list
	actionInfos := make([]checkaccess.ActionInfo, len(actions))
	for i, action := range actions {
		actionInfos[i] = checkaccess.ActionInfo{Id: action}
	}

	// Build subject attributes with oid and groups
	subjectAttrs := checkaccess.SubjectAttributes{
		ObjectId: userOid,
	}
	if len(groups) > 0 {
		subjectAttrs.Groups = groups
	}

	return checkaccess.AuthorizationRequest{
		Subject: checkaccess.SubjectInfo{
			Attributes: subjectAttrs,
		},
		Actions: actionInfos,
		Resource: checkaccess.ResourceInfo{
			Id: resourceId,
		},
	}
}

// convertV2ResponseToStatus converts CheckAccess v2 AuthorizationDecision responses
// to Kubernetes SubjectAccessReviewStatus format.
func convertV2ResponseToStatus(ctx context.Context, decisions []checkaccess.AuthorizationDecision) *authzv1.SubjectAccessReviewStatus {
	log := klog.FromContext(ctx)

	// Handle edge case: empty decisions (should not happen but defensive programming)
	if len(decisions) == 0 {
		log.V(5).Info("CheckAccess v2 returned no decisions, denying by default")
		return &authzv1.SubjectAccessReviewStatus{
			Allowed: false,
			Reason:  AccessNotAllowedVerdict,
			Denied:  true,
		}
	}

	// Check if any decision is denied
	allAllowed := true
	var denyReason string

	for _, decision := range decisions {
		// AccessDecision is embedded, access it via string conversion
		accessDecision := string(decision.AccessDecision)
		if !strings.EqualFold(accessDecision, Allowed) {
			allAllowed = false
			denyReason = fmt.Sprintf("Access denied for action: %s", decision.ActionId)
			log.V(7).Info("Access denied by v2 API", "actionId", decision.ActionId, "decision", accessDecision)
			break
		}
	}

	if allAllowed && len(decisions) > 0 {
		// Extract role assignment info from first decision for verbose verdict
		firstDecision := decisions[0]
		roleAssignmentId := firstDecision.RoleAssignment.Id
		roleDefinitionId := firstDecision.RoleAssignment.RoleDefinitionId

		verdict := fmt.Sprintf(AccessAllowedVerboseVerdict, roleAssignmentId, roleDefinitionId, "user")
		log.V(5).Info("Access allowed via v2 API",
			"roleAssignmentId", roleAssignmentId,
			"roleDefinitionId", roleDefinitionId,
		)

		return &authzv1.SubjectAccessReviewStatus{
			Allowed: true,
			Reason:  verdict,
			Denied:  false,
		}
	}

	// Access denied
	if denyReason == "" {
		denyReason = AccessNotAllowedVerdict
	}

	return &authzv1.SubjectAccessReviewStatus{
		Allowed: false,
		Reason:  denyReason,
		Denied:  true,
	}
}

// extractUserIdentityV2 extracts user oid and groups from SubjectAccessReviewSpec.
// This mirrors the v1 logic in prepareCheckAccessRequestBody for AKS scenario where
// user identity comes from request.Extra["oid"] and request.Groups, not from JWT claims.
// Reference: https://github.com/kubeguard/guard/blob/master/authz/providers/azure/rbac/checkaccessreqhelper.go#L674-L684
func extractUserIdentityV2(request *authzv1.SubjectAccessReviewSpec) (userOid string, groups []string, err error) {
	// Extract oid from request.Extra (same logic as v1)
	if oid, ok := request.Extra["oid"]; ok {
		val := oid.String()
		// Remove surrounding brackets from the string representation
		userOid = val[1 : len(val)-1]
		if !isValidUUID(userOid) {
			return "", nil, errutils.WithCode(
				fmt.Errorf("oid info sent from authentication module is not valid (oid: %s)", userOid),
				http.StatusBadRequest,
			)
		}
	} else {
		return "", nil, errutils.WithCode(
			fmt.Errorf("oid info not sent from authentication module"),
			http.StatusBadRequest,
		)
	}

	// Extract valid security groups (same logic as v1)
	groups = getValidSecurityGroups(request.Groups)

	return userOid, groups, nil
}

// checkAccessV2 is the main entry point for v2 API authorization checks.
// It handles the primary check and fallback logic using the v2 SDK.
func (a *AccessInfo) checkAccessV2(ctx context.Context, request *authzv1.SubjectAccessReviewSpec) (*authzv1.SubjectAccessReviewStatus, error) {
	log := klog.FromContext(ctx)

	// Extract user identity from request (oid and groups)
	// For AKS, user identity comes from SubjectAccessReviewSpec, not JWT claims
	userOid, groups, err := extractUserIdentityV2(request)
	if err != nil {
		return nil, err
	}

	log.V(7).Info("Extracted user identity for v2", "userOid", userOid, "groupsCount", len(groups))

	// Prepare actions list from request (same logic as v1 but get action IDs)
	actions, err := getDataActionsV2(ctx, request, a.clusterType, a.allowCustomResourceTypeCheck, a.allowSubresourceTypeCheck)
	if err != nil {
		return nil, fmt.Errorf("error preparing v2 actions list: %w", err)
	}

	// Determine resource ID (with or without namespace scope)
	namespaceExist, namespaceString := getNameSpaceScope(request, a.useNamespaceResourceScopeFormat)
	resourceId, err := buildResourceIDForV2(a.azureResourceId, namespaceExist, namespaceString)
	if err != nil {
		return nil, fmt.Errorf("error building primary resource ID: %w", err)
	}

	// Primary check
	log.V(5).Info("Performing primary CheckAccess v2", "resourceId", resourceId, "actionsCount", len(actions))
	status, err := a.performCheckAccessV2(ctx, resourceId, actions, userOid, groups)
	if err != nil {
		return nil, err
	}
	if status != nil && status.Allowed {
		return status, nil
	}

	// Fallback to managed namespace check
	managedNamespaceExists, managedNamespacePath := getManagedNameSpaceScope(request)
	if a.useManagedNamespaceResourceScopeFormat &&
		(a.clusterType == managedClusters || a.clusterType == fleets) &&
		managedNamespaceExists {
		log.V(7).Info("Falling back to managed namespace scope check (v2)", "namespacePath", managedNamespacePath)

		managedResourceId, err := buildResourceIDForV2(a.azureResourceId, true, managedNamespacePath)
		if err != nil {
			return nil, fmt.Errorf("error building managed namespace resource ID: %w", err)
		}
		status, err = a.performCheckAccessV2(ctx, managedResourceId, actions, userOid, groups)
		if err != nil {
			return nil, fmt.Errorf("Managed namespace CheckAccess v2 failed: %w", err)
		}
		if status != nil && status.Allowed {
			log.V(5).Info("Managed namespace CheckAccess v2 allowed")
			return status, nil
		}
	}

	// Fallback to fleet scope check
	if a.fleetManagerResourceId != "" {
		log.V(7).Info("Falling back to fleet manager scope check (v2)", "fleetResourceId", a.fleetManagerResourceId)

		fleetResourceId, err := buildResourceIDForV2(a.fleetManagerResourceId, namespaceExist, namespaceString)
		if err != nil {
			return nil, fmt.Errorf("error building fleet manager resource ID: %w", err)
		}

		// For fleet members, we may need different actions - reuse v1 logic if needed
		status, err = a.performCheckAccessV2(ctx, fleetResourceId, actions, userOid, groups)
		if err != nil {
			return nil, fmt.Errorf("Fleet manager CheckAccess v2 failed: %w", err)
		}
		if status != nil && status.Allowed {
			log.V(5).Info("Fleet manager CheckAccess v2 allowed")
			return status, nil
		}

		if managedNamespaceExists {
			fleetManagedResourceId, err := buildResourceIDForV2(a.fleetManagerResourceId, true, managedNamespacePath)
			if err != nil {
				return nil, fmt.Errorf("error building fleet manager managed namespace resource ID: %w", err)
			}

			// For fleet members, we may need different actions - reuse v1 logic if needed
			status, err = a.performCheckAccessV2(ctx, fleetManagedResourceId, actions, userOid, groups)
			if err != nil {
				return nil, fmt.Errorf("Fleet manager managed namespace CheckAccess v2 failed: %w", err)
			}
			if status != nil && status.Allowed {
				log.V(5).Info("Fleet manager managed namespace CheckAccess v2 allowed")
				return status, nil
			}
		}
	}

	return status, nil
}

// getDataActionsV2 converts SubjectAccessReviewSpec to a list of action IDs for v2 API.
// This extracts the action ID portion from the v1 AuthorizationActionInfo.
func getDataActionsV2(ctx context.Context, request *authzv1.SubjectAccessReviewSpec, clusterType string, allowCustomResourceTypeCheck bool, allowSubresourceTypeCheck bool) ([]string, error) {
	// Reuse v1 logic to get actions
	authInfoList, err := getDataActions(ctx, request, clusterType, allowCustomResourceTypeCheck, allowSubresourceTypeCheck)
	if err != nil {
		return nil, err
	}

	// Extract action IDs
	actions := make([]string, len(authInfoList))
	for i, authInfo := range authInfoList {
		actions[i] = authInfo.AuthorizationEntity.Id
	}

	return actions, nil
}

// buildResourceIDForV2 constructs and validates a resource ID for CheckAccess v2 API.
// This function provides defensive validation similar to buildCheckAccessURL for v1.
//
// Invariants:
//  1. baseResourceID must not be empty
//  2. the resulting resource ID must start with baseResourceID (prevents path traversal)
//
// Any invariant violation will result in an error being returned.
func buildResourceIDForV2(baseResourceID string, hasNamespace bool, namespacePath string) (string, error) {
	// invariant 1: baseResourceID must not be empty
	if baseResourceID == "" {
		return "", fmt.Errorf("baseResourceID must not be empty")
	}

	resourceID := baseResourceID
	if hasNamespace {
		resourceID = path.Join(resourceID, namespacePath)
	}

	// invariant 2: ensure no path traversal occurred
	normalizedBase := strings.TrimPrefix(baseResourceID, "/")
	normalizedResult := strings.TrimPrefix(resourceID, "/")
	if !strings.HasPrefix(normalizedResult, normalizedBase) {
		return "", fmt.Errorf("invalid resource ID %q, expected to start with %q (possible path traversal)", resourceID, baseResourceID)
	}

	return resourceID, nil
}
