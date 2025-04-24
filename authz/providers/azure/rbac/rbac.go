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
	"github.com/pkg/errors"
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
	connectedClusters         = "Microsoft.Kubernetes/connectedClusters"
	checkAccessPath           = "/providers/Microsoft.Authorization/checkaccess"
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
	void                    struct{}
	correlationRequestIDKey string
)

// AccessInfo allows you to check user access from MS RBAC
type AccessInfo struct {
	headers   http.Header
	client    *http.Client
	expiresAt time.Time
	// These allow us to mock out the URL for testing
	apiURL *url.URL

	tokenProvider                   graph.TokenProvider
	clusterType                     string
	azureResourceId                 string
	armCallLimit                    int
	skipCheck                       map[string]void
	skipAuthzForNonAADUsers         bool
	allowNonResDiscoveryPathAccess  bool
	allowCustomResourceTypeCheck    bool
	useNamespaceResourceScopeFormat bool
	httpClientRetryCount            int
	lock                            sync.RWMutex
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
		apiURL:                          rbacURL,
		tokenProvider:                   tokenProvider,
		azureResourceId:                 opts.ResourceId,
		armCallLimit:                    opts.ARMCallLimit,
		skipAuthzForNonAADUsers:         opts.SkipAuthzForNonAADUsers,
		allowNonResDiscoveryPathAccess:  opts.AllowNonResDiscoveryPathAccess,
		allowCustomResourceTypeCheck:    opts.AllowCustomResourceTypeCheck,
		useNamespaceResourceScopeFormat: opts.UseNamespaceResourceScopeFormat,
		httpClientRetryCount:            authopts.HttpClientRetryCount,
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
			klog.Errorf("%s failed to refresh token : %s", a.tokenProvider.Name(), err.Error())
			return errors.Wrap(err, "failed to refresh rbac token")
		}

		// Set the authorization headers for future requests
		a.headers.Set("Authorization", fmt.Sprintf("Bearer %s", resp.Token))

		// Use ExpiresOn to set the expiration time
		expOn := time.Unix(int64(resp.ExpiresOn), 0)
		a.expiresAt = expOn.Add(-tokenExpiryDelta)
		klog.Infof("Token refreshed successfully at %s. Expire at set to: %s", time.Now(), a.expiresAt)
	}

	return nil
}

func (a *AccessInfo) IsTokenExpired() bool {
	return a.expiresAt.Before(time.Now())
}

func (a *AccessInfo) ShouldSkipAuthzCheckForNonAADUsers() bool {
	return a.skipAuthzForNonAADUsers
}

func (a *AccessInfo) GetResultFromCache(request *authzv1.SubjectAccessReviewSpec, store authz.Store) (bool, bool) {
	var result bool
	key := getResultCacheKey(request)
	klog.V(10).Infof("Cache search for key: %s", key)
	found, _ := store.Get(key, &result)

	if found {
		if result {
			klog.V(5).Infof("cache hit: returning allowed for key %s", key)
		} else {
			klog.V(5).Infof("cache hit: returning denied for key %s", key)
		}
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

func (a *AccessInfo) SetResultInCache(request *authzv1.SubjectAccessReviewSpec, result bool, store authz.Store) error {
	key := getResultCacheKey(request)
	klog.V(5).Infof("Cache set for key: %s, value: %t", key, result)
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

func (a *AccessInfo) CheckAccess(request *authzv1.SubjectAccessReviewSpec) (*authzv1.SubjectAccessReviewStatus, error) {
	checkAccessBodies, err := prepareCheckAccessRequestBody(request, a.clusterType, a.azureResourceId, a.useNamespaceResourceScopeFormat, a.allowCustomResourceTypeCheck)
	if err != nil {
		return nil, errors.Wrap(err, "error in preparing check access request")
	}

	checkAccessUsername := request.User

	checkAccessURL := *a.apiURL
	// Append the path for azure cluster resource id
	checkAccessURL.Path = path.Join(checkAccessURL.Path, a.azureResourceId)
	exist, nameSpaceString := getNameSpaceScope(request, a.useNamespaceResourceScopeFormat)
	if exist {
		checkAccessURL.Path = path.Join(checkAccessURL.Path, nameSpaceString)
	}

	checkAccessURL.Path = path.Join(checkAccessURL.Path, checkAccessPath)
	params := url.Values{}
	params.Add("api-version", checkAccessAPIVersion)
	checkAccessURL.RawQuery = params.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), checkaccessContextTimeout)
	defer cancel()
	eg, egCtx := errgroup.WithContext(ctx)

	ch := make(chan *authzv1.SubjectAccessReviewStatus, len(checkAccessBodies))
	if len(checkAccessBodies) > 1 {
		klog.V(5).Infof("Number of checkaccess requests to make: %d", len(checkAccessBodies))
	}
	eg.SetLimit(len(checkAccessBodies))
	for _, checkAccessBody := range checkAccessBodies {
		body := checkAccessBody
		eg.Go(func() error {
			// create a request id for every checkaccess request
			requestUUID := uuid.New()
			reqContext := context.WithValue(egCtx, correlationRequestIDKey(correlationRequestIDHeader), []string{requestUUID.String()})
			reqContext = azureutils.WithRetryableHttpClient(reqContext, a.httpClientRetryCount)
			err := a.sendCheckAccessRequest(reqContext, checkAccessUsername, checkAccessURL, body, ch)
			if err != nil {
				code := http.StatusInternalServerError
				if v, ok := err.(errutils.HttpStatusCode); ok {
					code = v.Code()
				}
				err = errutils.WithCode(errors.Errorf("Error: %s. Correlation ID: %s", err, requestUUID.String()), code)
				return err
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			klog.V(5).Infof("Checkaccess requests have timed out. Error: %v", ctx.Err())
			actionsCount := 0
			for i := 0; i < len(checkAccessBodies); i += 1 {
				actionsCount = actionsCount + len(checkAccessBodies[i].Actions)
			}
			checkAccessContextTimedOutCount.WithLabelValues(azureutils.ConvertIntToString(len(checkAccessBodies)), azureutils.ConvertIntToString(actionsCount)).Inc()
			close(ch)
			return nil, errutils.WithCode(errors.Wrap(ctx.Err(), "Checkaccess requests have timed out."), http.StatusInternalServerError)
		} else {
			close(ch)
			// print error we get from sendcheckAccessRequest
			klog.Error(err)
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

func (a *AccessInfo) sendCheckAccessRequest(ctx context.Context, checkAccessUsername string, checkAccessURL url.URL, checkAccessBody *CheckAccessRequest, ch chan *authzv1.SubjectAccessReviewStatus) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(checkAccessBody); err != nil {
		return errutils.WithCode(errors.Wrap(err, "error encoding check access request"), http.StatusInternalServerError)
	}

	if klog.V(10).Enabled() {
		binaryData, _ := json.MarshalIndent(checkAccessBody, "", "    ")
		klog.V(10).Infof("checkAccessURI:%s", checkAccessURL.String())
		klog.V(10).Infof("binary data:%s", binaryData)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, checkAccessURL.String(), buf)
	if err != nil {
		return errutils.WithCode(errors.Wrap(err, "error creating check access request"), http.StatusInternalServerError)
	}

	a.setReqHeaders(req)
	// set x-ms-correlation-request-id for the checkaccess request
	correlationID := ctx.Value(correlationRequestIDKey(correlationRequestIDHeader)).([]string)
	req.Header[correlationRequestIDHeader] = correlationID
	internalServerCode := azureutils.ConvertIntToString(http.StatusInternalServerError)
	// start time to calculate checkaccess duration
	start := time.Now()
	klog.V(5).Infof("Sending checkAccess request with correlationID: %s", correlationID[0])
	client := azureutils.LoadClientWithContext(ctx, a.client)
	resp, err := client.Do(req)
	duration := time.Since(start).Seconds()
	if err != nil {
		checkAccessTotal.WithLabelValues(internalServerCode).Inc()
		checkAccessDuration.WithLabelValues(internalServerCode).Observe(duration)
		return errutils.WithCode(errors.Wrap(err, "error in check access request execution."), http.StatusInternalServerError)
	}

	defer resp.Body.Close()
	respStatusCode := azureutils.ConvertIntToString(resp.StatusCode)
	checkAccessTotal.WithLabelValues(respStatusCode).Inc()
	checkAccessDuration.WithLabelValues(respStatusCode).Observe(duration)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		checkAccessTotal.WithLabelValues(internalServerCode).Inc()
		checkAccessDuration.WithLabelValues(internalServerCode).Observe(duration)
		return errutils.WithCode(errors.Wrap(err, "error in reading response body"), http.StatusInternalServerError)
	}

	klog.V(7).Infof("checkaccess response: %s, Configured ARM call limit: %d", string(data), a.armCallLimit)
	if resp.StatusCode != http.StatusOK {
		klog.Errorf("error in check access response. error code: %d, response: %s, correlationID: %s", resp.StatusCode, string(data), correlationID[0])
		// metrics for calls with StatusCode >= 300
		if resp.StatusCode >= http.StatusMultipleChoices {
			if resp.StatusCode == http.StatusTooManyRequests {
				klog.V(10).Infoln("Closing idle TCP connections.")
				a.client.CloseIdleConnections()
				checkAccessThrottled.Inc()
			}

			checkAccessFailed.WithLabelValues(respStatusCode).Inc()
		}

		return errutils.WithCode(errors.Errorf("request %s failed with status code: %d and response: %s", req.URL.Path, resp.StatusCode, string(data)), resp.StatusCode)
	} else {
		remaining := resp.Header.Get(remainingSubReadARMHeader)
		klog.Infof("Checkaccess Request has succeeded, CorrelationID is %s. Remaining request count in ARM instance:%s", correlationID[0], remaining)
		count, _ := strconv.Atoi(remaining)
		if count < a.armCallLimit {
			if klog.V(10).Enabled() {
				klog.V(10).Infoln("Closing idle TCP connections.")
			}
			// Usually ARM connections are cached by destination ip and port
			// By closing the idle connection, a new request will use different port which
			// will connect to different ARM instance of the region to ensure there is no ARM throttling
			a.client.CloseIdleConnections()
		}
		checkAccessSucceeded.Inc()
	}

	// Decode response and prepare k8s response
	status, err := ConvertCheckAccessResponse(checkAccessUsername, data)
	if err != nil {
		return err
	}

	ch <- status
	return nil
}
