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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	auth "go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/auth/providers/azure/graph"
	"go.kubeguard.dev/guard/authz"
	authzOpts "go.kubeguard.dev/guard/authz/providers/azure/options"
	azureutils "go.kubeguard.dev/guard/util/azure"
	errutils "go.kubeguard.dev/guard/util/error"
	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/errgroup"
	v "gomodules.xyz/x/version"
	authzv1 "k8s.io/api/authorization/v1"
	"k8s.io/klog/v2"
)

const (
	managedClusters           = "Microsoft.ContainerService/managedClusters"
	fleets                    = "Microsoft.ContainerService/fleets"
	fleetMembers              = "Microsoft.ContainerService/fleets/members"
	connectedClusters         = "Microsoft.Kubernetes/connectedClusters"
	checkAccessPath           = "/providers/Microsoft.Authorization/checkaccess"
	queryParamAPIVersion      = "api-version"
	checkAccessAPIVersion     = "2018-09-01-preview"
	remainingSubReadARMHeader = "x-ms-ratelimit-remaining-subscription-reads"
	// Time delta to refresh token before expiry
	tokenExpiryDelta           = 300 * time.Second
	checkaccessContextTimeout  = 23 * time.Second
	correlationRequestIDHeader = "x-ms-correlation-request-id"
)

type AuthzInfo struct {
	AADEndpoint string
	ARMEndPoint string
}

type (
	void struct{}
)

// AccessInfo allows you to check user access from MS RBAC
type AccessInfo struct {
	headers   http.Header
	client    *http.Client
	expiresAt time.Time
	// These allow us to mock out the URL for testing
	apiURL *url.URL

	tokenProvider                          graph.TokenProvider
	clusterType                            string
	azureResourceId                        string
	armCallLimit                           int
	skipCheck                              map[string]void
	skipAuthzForNonAADUsers                bool
	allowNonResDiscoveryPathAccess         bool
	allowCustomResourceTypeCheck           bool
	allowSubresourceTypeCheck              bool
	useManagedNamespaceResourceScopeFormat bool
	useNamespaceResourceScopeFormat        bool
	httpClientRetryCount                   int
	lock                                   sync.RWMutex

	auditSAR               bool
	fleetManagerResourceId string
}

var (
	checkAccessThrottled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "guard_azure_checkaccess_throttling_failure_total",
		Help: "No of throttled checkaccess calls.",
	})

	checkAccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guard_azure_check_access_requests_total",
			Help: "Number of checkaccess request calls.",
		},
		[]string{"code"},
	)

	checkAccessFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guard_azure_checkaccess_failure_total",
			Help: "No of checkaccess failures",
		},
		[]string{"code"},
	)

	checkAccessSucceeded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "guard_azure_checkaccess_success_total",
		Help: "Number of successful checkaccess calls.",
	})

	checkAccessContextTimedOutCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guard_azure_checkaccess_context_timeout",
			Help: "No of checkacces context timeout calls",
		},
		[]string{"checkAccessBatchCount", "totalActionsCount"},
	)

	// checkAccessDuration is partitioned by the HTTP status code It uses custom
	// buckets based on the expected request duration.
	checkAccessDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "guard_azure_checkaccess_request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10, 15, 20},
		},
		[]string{"code"},
	)

	CheckAccessErrorFormat = "Error occured during authorization check. Please retry again. Error: %s"
)

func init() {
	prometheus.MustRegister(checkAccessDuration, checkAccessTotal, checkAccessFailed, checkAccessContextTimedOutCount)
}

func getClusterType(clsType string) string {
	switch clsType {
	case authzOpts.ARCAuthzMode:
		return connectedClusters
	case authzOpts.AKSAuthzMode:
		return managedClusters
	case authzOpts.FleetAuthzMode:
		return fleets
	default:
		return ""
	}
}

func newAccessInfo(tokenProvider graph.TokenProvider, rbacURL *url.URL, opts authzOpts.Options, authopts auth.Options) (*AccessInfo, error) {
	u := &AccessInfo{
		client: httpclient.DefaultHTTPClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
			"User-Agent":   []string{fmt.Sprintf("guard-%s-%s-%s-%s", v.Version.Platform, v.Version.GoVersion, v.Version.Version, opts.AuthzMode)},
		},
		apiURL:                                 rbacURL,
		tokenProvider:                          tokenProvider,
		azureResourceId:                        opts.ResourceId,
		armCallLimit:                           opts.ARMCallLimit,
		skipAuthzForNonAADUsers:                opts.SkipAuthzForNonAADUsers,
		allowNonResDiscoveryPathAccess:         opts.AllowNonResDiscoveryPathAccess,
		allowCustomResourceTypeCheck:           opts.AllowCustomResourceTypeCheck,
		allowSubresourceTypeCheck:              opts.AllowSubresourceTypeCheck,
		useManagedNamespaceResourceScopeFormat: opts.UseManagedNamespaceResourceScopeFormat,
		useNamespaceResourceScopeFormat:        opts.UseNamespaceResourceScopeFormat,
		httpClientRetryCount:                   authopts.HttpClientRetryCount,
		auditSAR:                               opts.AuditSAR,
		fleetManagerResourceId:                 opts.FleetManagerResourceId,
	}

	u.skipCheck = make(map[string]void, len(opts.SkipAuthzCheck))
	var member void
	for _, s := range opts.SkipAuthzCheck {
		u.skipCheck[strings.ToLower(s)] = member
	}

	u.clusterType = getClusterType(opts.AuthzMode)

	u.lock = sync.RWMutex{}

	return u, nil
}

func New(opts authzOpts.Options, authopts auth.Options, authzInfo *AuthzInfo) (*AccessInfo, error) {
	rbacURL, err := url.Parse(authzInfo.ARMEndPoint)
	if err != nil {
		return nil, err
	}

	var tokenProvider graph.TokenProvider
	switch opts.AuthzMode {
	case authzOpts.ARCAuthzMode:
		// if client secret is there check use client credential provider
		if authopts.ClientSecret != "" {
			tokenProvider = graph.NewClientCredentialTokenProvider(authopts.ClientID, authopts.ClientSecret,
				fmt.Sprintf("%s%s/oauth2/v2.0/token", authzInfo.AADEndpoint, authopts.TenantID),
				fmt.Sprintf("%s.default", authzInfo.ARMEndPoint))
		} else {
			tokenProvider = graph.NewMSITokenProvider(authzInfo.ARMEndPoint, graph.MSIEndpointForARC)
		}
	case authzOpts.FleetAuthzMode:
		tokenProvider = graph.NewAKSTokenProvider(opts.AKSAuthzTokenURL, authopts.TenantID)
	case authzOpts.AKSAuthzMode:
		tokenProvider = graph.NewAKSTokenProvider(opts.AKSAuthzTokenURL, authopts.TenantID)
	}

	return newAccessInfo(tokenProvider, rbacURL, opts, authopts)
}

func (a *AccessInfo) RefreshToken(ctx context.Context) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.IsTokenExpired() {
		resp, err := a.tokenProvider.Acquire(ctx, "")
		if err != nil {
			return fmt.Errorf("Failed to refresh token (provider: %s): %w", a.tokenProvider.Name(), err)
		}

		// Set the authorization headers for future requests
		a.headers.Set("Authorization", fmt.Sprintf("Bearer %s", resp.Token))

		// Use ExpiresOn to set the expiration time
		expOn := time.Unix(int64(resp.ExpiresOn), 0)
		a.expiresAt = expOn.Add(-tokenExpiryDelta)
		klog.FromContext(ctx).Info("Token refreshed successfully", "refreshedAt", time.Now(), "expiresAt", a.expiresAt)
	}

	return nil
}

func (a *AccessInfo) IsTokenExpired() bool {
	return a.expiresAt.Before(time.Now())
}

func (a *AccessInfo) ShouldSkipAuthzCheckForNonAADUsers() bool {
	return a.skipAuthzForNonAADUsers
}

func (a *AccessInfo) GetResultFromCache(ctx context.Context, request *authzv1.SubjectAccessReviewSpec, store authz.Store) (bool, bool) {
	log := klog.FromContext(ctx)
	var result bool
	key := getResultCacheKey(request, a.allowSubresourceTypeCheck)
	log.V(10).Info("Cache search", "key", key)
	found, err := store.Get(key, &result)

	if err != nil {
		// Error contains cache statistics for troubleshooting
		log.V(5).Info("Cache get error", "key", key, "error", err)
		return false, false
	}

	if found {
		if result {
			log.V(5).Info("Cache hit: allowed", "key", key)
		} else {
			log.V(5).Info("Cache hit: denied", "key", key)
		}
	} else {
		// Cache miss - log for observability
		log.V(5).Info("Cache miss", "key", key)
	}

	return found, result
}

func (a *AccessInfo) SkipAuthzCheck(request *authzv1.SubjectAccessReviewSpec) bool {
	if a.clusterType == connectedClusters {
		_, ok := a.skipCheck[strings.ToLower(request.User)]
		return ok
	}
	return false
}

func (a *AccessInfo) SetResultInCache(ctx context.Context, request *authzv1.SubjectAccessReviewSpec, result bool, store authz.Store) error {
	log := klog.FromContext(ctx)
	key := getResultCacheKey(request, a.allowSubresourceTypeCheck)
	log.V(10).Info("Cache set", "key", key, "value", result)
	return store.Set(key, result)
}

func (a *AccessInfo) AllowNonResPathDiscoveryAccess(request *authzv1.SubjectAccessReviewSpec) bool {
	if request.NonResourceAttributes != nil && a.allowNonResDiscoveryPathAccess && strings.EqualFold(request.NonResourceAttributes.Verb, "get") {
		path := strings.ToLower(request.NonResourceAttributes.Path)
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/openapi") || strings.HasPrefix(path, "/version") || strings.HasPrefix(path, "/healthz") {
			return true
		}
	}
	return false
}

func (a *AccessInfo) setReqHeaders(req *http.Request) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	// Set the auth headers for the request
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	for k, value := range a.headers {
		req.Header[k] = value
	}
}

func (a *AccessInfo) performCheckAccess(
	parentCtx context.Context,
	checkAccessURL string,
	checkAccessBodies []*CheckAccessRequest,
	checkAccessUsername string,
) (*authzv1.SubjectAccessReviewStatus, error) {
	ctx, cancel := context.WithTimeout(parentCtx, checkaccessContextTimeout)
	defer cancel()
	eg, egCtx := errgroup.WithContext(ctx)

	ch := make(chan *authzv1.SubjectAccessReviewStatus, len(checkAccessBodies))
	if len(checkAccessBodies) > 1 {
		klog.FromContext(parentCtx).V(5).Info("Multiple checkaccess requests to execute", "batchCount", len(checkAccessBodies))
	}
	eg.SetLimit(len(checkAccessBodies))
	for batchIndex, checkAccessBody := range checkAccessBodies {
		body := checkAccessBody
		index := batchIndex
		eg.Go(func() error {
			// create a correlation ID for every checkaccess request
			correlationID := uuid.New().String()

			// Create logger with correlation ID and batch context
			log := klog.FromContext(parentCtx).WithValues("correlationID", correlationID, "batchIndex", index)
			reqContext := klog.NewContext(egCtx, log)
			reqContext = azureutils.WithRetryableHttpClient(reqContext, a.httpClientRetryCount)

			log.V(7).Info("Starting checkaccess batch", "actionsCount", len(body.Actions))
			err := a.sendCheckAccessRequest(reqContext, checkAccessUsername, checkAccessURL, body, ch)
			if err != nil {
				code := http.StatusInternalServerError
				if v, ok := err.(errutils.HttpStatusCode); ok {
					code = v.Code()
				}
				err = errutils.WithCode(fmt.Errorf("Checkaccess batch failed (batchIndex: %d, statusCode: %d): %w", index, code, err), code)
				return err
			}
			log.V(7).Info("Checkaccess batch completed successfully")
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			actionsCount := 0
			for i := 0; i < len(checkAccessBodies); i += 1 {
				actionsCount = actionsCount + len(checkAccessBodies[i].Actions)
			}
			checkAccessContextTimedOutCount.WithLabelValues(azureutils.ConvertIntToString(len(checkAccessBodies)), azureutils.ConvertIntToString(actionsCount)).Inc()
			close(ch)
			return nil, errutils.WithCode(fmt.Errorf("Checkaccess requests timed out (batchCount: %d, totalActionsCount: %d, timeoutSeconds: %.0f): %w", len(checkAccessBodies), actionsCount, checkaccessContextTimeout.Seconds(), ctx.Err()), http.StatusInternalServerError)
		} else {
			close(ch)
			// error already contains context from child goroutines
			return nil, err
		}
	}
	close(ch)

	var finalStatus *authzv1.SubjectAccessReviewStatus
	for status := range ch {
		if status.Denied {
			finalStatus = status
			break
		}

		finalStatus = status
	}
	return finalStatus, nil
}

// auditSARIfNeeded logs the SubjectAccessReview request if auditing is enabled.
func (a *AccessInfo) auditSARIfNeeded(ctx context.Context, request *authzv1.SubjectAccessReviewSpec) {
	if !a.auditSAR {
		return
	}

	// NOTE: aligning with the same log level used in the sendCheckAccessRequest.
	// so we will only add at most one more log per request
	logger := klog.FromContext(ctx).V(5)

	if request == nil {
		logger.Info("SubjectAccessReview request is nil")
		return
	}

	if request.ResourceAttributes == nil {
		logger.Info("SubjectAccessReview details", "ResourceAttributes", "<nil>")
	} else {
		logger.Info("SubjectAccessReview details", "ResourceAttributes", request.ResourceAttributes)
	}

	if request.NonResourceAttributes != nil {
		logger.Info("SubjectAccessReview non-resource attributes",
			"path", request.NonResourceAttributes.Path,
			"verb", request.NonResourceAttributes.Verb,
		)
	}
}

func (a *AccessInfo) CheckAccess(ctx context.Context, request *authzv1.SubjectAccessReviewSpec) (*authzv1.SubjectAccessReviewStatus, error) {
	a.auditSARIfNeeded(ctx, request)

	checkAccessBodies, err := prepareCheckAccessRequestBody(ctx, request, a.clusterType, a.azureResourceId, a.useNamespaceResourceScopeFormat, a.allowCustomResourceTypeCheck, a.allowSubresourceTypeCheck)
	if err != nil {
		return nil, fmt.Errorf("error in preparing check access request: %w", err)
	}

	checkAccessUsername := request.User

	// Build primary check access URL
	exist, nameSpaceString := getNameSpaceScope(request, a.useNamespaceResourceScopeFormat)
	checkAccessURL, err := buildCheckAccessURL(*a.apiURL, a.azureResourceId, exist, nameSpaceString)
	if err != nil {
		return nil, fmt.Errorf("error in building check access URL: %w", err)
	}

	log := klog.FromContext(ctx)
	log.V(5).Info("Performing primary check access", "batchCount", len(checkAccessBodies))
	status, err := a.performCheckAccess(ctx, checkAccessURL, checkAccessBodies, checkAccessUsername)
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
		log.V(7).Info("Falling back to managed namespace scope check", "namespacePath", managedNamespacePath)
		// Build managed namespace URL
		managedNamespaceURL, err := buildCheckAccessURL(*a.apiURL, a.azureResourceId, true, managedNamespacePath)
		if err != nil {
			return nil, fmt.Errorf("error in building managed namespace check access URL: %w", err)
		}

		// Update resource IDs for managed namespace
		for _, b := range checkAccessBodies {
			b.Resource.Id = path.Join(a.azureResourceId, managedNamespacePath)
		}

		status, err = a.performCheckAccess(ctx, managedNamespaceURL, checkAccessBodies, checkAccessUsername)
		if err != nil {
			return nil, fmt.Errorf("Managed namespace check access failed: %w", err)
		}
		if status != nil && status.Allowed {
			log.V(5).Info("Managed namespace check access allowed")
			return status, nil
		}
	}

	// Fallback to fleet scope check when the managedCluster has joined a fleet
	if a.fleetManagerResourceId != "" {
		log.V(7).Info("Falling back to fleet manager scope check", "fleetResourceId", a.fleetManagerResourceId)
		fleetURL, err := buildCheckAccessURL(*a.apiURL, a.fleetManagerResourceId, managedNamespaceExists, managedNamespacePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to build fleet manager check access URL: %w", err)
		}
		bodiesForFleetRBAC, err := prepareCheckAccessRequestBody(ctx, request, fleetMembers, a.fleetManagerResourceId, false, a.allowCustomResourceTypeCheck, a.allowSubresourceTypeCheck)
		if err != nil {
			return nil, fmt.Errorf("Failed to prepare check access request for fleet manager: %w", err)
		}
		if managedNamespaceExists {
			for _, b := range bodiesForFleetRBAC {
				b.Resource.Id = path.Join(a.fleetManagerResourceId, managedNamespacePath)
			}
		}
		status, err = a.performCheckAccess(ctx, fleetURL, bodiesForFleetRBAC, checkAccessUsername)
		if err != nil {
			err = fmt.Errorf("Fleet manager check access failed: %w", err)
		} else if status != nil && status.Allowed {
			log.V(5).Info("Fleet manager check access allowed")
		}
		return status, err
	}
	return status, nil
}

func (a *AccessInfo) sendCheckAccessRequest(ctx context.Context, checkAccessUsername string, checkAccessURL string, checkAccessBody *CheckAccessRequest, ch chan *authzv1.SubjectAccessReviewStatus) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(checkAccessBody); err != nil {
		return errutils.WithCode(fmt.Errorf("Failed to encode check access request: %w", err), http.StatusInternalServerError)
	}

	log := klog.FromContext(ctx)
	if log.V(10).Enabled() {
		binaryData, _ := json.MarshalIndent(checkAccessBody, "", "    ")
		log.V(10).Info("CheckAccess request details", "url", checkAccessURL, "body", string(binaryData))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, checkAccessURL, buf)
	if err != nil {
		return errutils.WithCode(fmt.Errorf("Failed to create check access request: %w", err), http.StatusInternalServerError)
	}

	a.setReqHeaders(req)

	// Get logger from context (already has correlationID and batchIndex from parent)
	log = klog.FromContext(ctx)

	internalServerCode := azureutils.ConvertIntToString(http.StatusInternalServerError)
	// start time to calculate checkaccess duration
	start := time.Now()
	log.V(5).Info("Sending checkAccess request to Azure")
	client := azureutils.LoadClientWithContext(ctx, a.client)
	resp, err := client.Do(req)
	duration := time.Since(start).Seconds()
	if err != nil {
		checkAccessTotal.WithLabelValues(internalServerCode).Inc()
		checkAccessDuration.WithLabelValues(internalServerCode).Observe(duration)
		return errutils.WithCode(fmt.Errorf("CheckAccess request execution failed (durationSeconds: %.2f): %w", duration, err), http.StatusInternalServerError)
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	respStatusCode := azureutils.ConvertIntToString(resp.StatusCode)
	checkAccessTotal.WithLabelValues(respStatusCode).Inc()
	checkAccessDuration.WithLabelValues(respStatusCode).Observe(duration)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		checkAccessTotal.WithLabelValues(internalServerCode).Inc()
		checkAccessDuration.WithLabelValues(internalServerCode).Observe(duration)
		return errutils.WithCode(fmt.Errorf("Failed to read response body: %w", err), http.StatusInternalServerError)
	}

	log.V(7).Info("CheckAccess response received", "statusCode", resp.StatusCode, "durationSeconds", duration, "responseBody", string(data), "armCallLimit", a.armCallLimit)

	// We can expect the response to be a 200 OK for ARM proxy resources or 404 Not Found for ARM tracked resources due to resource deletion
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		// metrics for calls with StatusCode >= 300
		if resp.StatusCode >= http.StatusMultipleChoices {
			if resp.StatusCode == http.StatusTooManyRequests {
				log.Info("Azure throttling detected (HTTP 429), closing idle TCP connections to switch ARM instance")
				a.client.CloseIdleConnections()
				checkAccessThrottled.Inc()
			}

			checkAccessFailed.WithLabelValues(respStatusCode).Inc()
		}

		return errutils.WithCode(fmt.Errorf("CheckAccess request failed (statusCode: %d, durationSeconds: %.2f): request %s failed with response: %s", resp.StatusCode, duration, req.URL.Path, string(data)), resp.StatusCode)
	}

	remaining := resp.Header.Get(remainingSubReadARMHeader)
	log.Info("CheckAccess request succeeded", "remainingARMCalls", remaining, "durationSeconds", duration)
	count, _ := strconv.Atoi(remaining)
	if count < a.armCallLimit {
		log.Info("ARM call limit threshold reached, closing idle TCP connections to switch ARM instance", "remainingCalls", count, "threshold", a.armCallLimit)
		// Usually ARM connections are cached by destination ip and port
		// By closing the idle connection, a new request will use different port which
		// will connect to different ARM instance of the region to ensure there is no ARM throttling
		a.client.CloseIdleConnections()
	}
	checkAccessSucceeded.Inc()

	var status *authzv1.SubjectAccessReviewStatus
	if resp.StatusCode == http.StatusNotFound {
		log.V(5).Info("CheckAccess returned 404, tracked resource deleted, returning default not found decision")
		status = defaultNotFoundDecision()
	} else {
		// Decode response and prepare k8s response
		status, err = ConvertCheckAccessResponse(ctx, checkAccessUsername, data)
		if err != nil {
			return fmt.Errorf("Failed to convert check access response: %w", err)
		}
	}

	ch <- status
	return nil
}
