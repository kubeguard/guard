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
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.kubeguard.dev/guard/auth/providers/azure/graph"
	errutils "go.kubeguard.dev/guard/util/error"
	"go.kubeguard.dev/guard/util/httpclient"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-retryablehttp"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
	v "gomodules.xyz/x/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ManagedClusters             = "Microsoft.ContainerService/managedClusters"
	Fleets                      = "Microsoft.ContainerService/fleets"
	ConnectedClusters           = "Microsoft.Kubernetes/connectedClusters"
	OperationsEndpointFormatARC = "%s/providers/Microsoft.Kubernetes/operations?api-version=2021-10-01"
	OperationsEndpointFormatAKS = "%s/providers/Microsoft.ContainerService/operations?api-version=2018-10-31"
	retrieveServerApiError      = "unable to retrieve the complete list of server APIs"
	apiserviceUnreachableError  = "the server is currently unable to handle the request"
)

var (
	operationsMapLock = sync.RWMutex{}
	operationsMap     = NewOperationsMap()
	settings          *DiscoverResourcesSettings
)

var (
	discoverResourcesApiServerCallDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "guard_apiresources_request_duration_seconds",
			Help:    "A histogram of latencies for apiserver requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10, 15, 20},
		})

	discoverResourcesAzureCallDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "guard_azure_get_operations_request_duration_seconds",
			Help:    "A histogram of latencies for azure get operations requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10, 15, 20},
		})

	DiscoverResourcesTotalDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "guard_discover_resources_request_duration_seconds",
			Help:    "A histogram of latencies for azure get operations requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10, 15, 20},
		})

	counterDiscoverResources = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guard_discover_requests_requests_total",
			Help: "A counter for discover resources.",
		},
		[]string{"code"},
	)

	counterGetOperationsResources = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "guard_azure_get_operations_requests_total",
			Help: "A counter for get operations call in discover resources.",
		},
		[]string{"code"},
	)
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type DiscoverResourcesSettings struct {
	clusterType        string
	environment        azure.Environment
	operationsEndpoint string
	aksLoginURL        string
	kubeconfigFilePath string
	tenantID           string
	clientID           string
	clientSecret       string
}

type Display struct {
	Provider    string `json:"provider"`
	Resource    string `json:"resource"`
	Operation   string `json:"operation"`
	Description string `json:"description"`
}

type Operation struct {
	Name         string  `json:"name"`
	Display      Display `json:"display"`
	IsDataAction *bool   `json:"isDataAction,omitempty"`
}

type OperationList struct {
	Value    []Operation `json:"value"`
	NextLink string      `json:"nextLink"`
}

type Resource struct {
	Id         string
	Namespaced bool
	Name       string
	Group      string
	Verb       string
}

type AuthorizationEntity struct {
	Id string `json:"Id"`
}

type AuthorizationActionInfo struct {
	AuthorizationEntity
	IsDataAction bool              `json:"IsDataAction"`
	Attributes   map[string]string `json:"Attributes"`
}

type DataAction struct {
	ActionInfo           AuthorizationActionInfo
	IsNamespacedResource bool
}

type VerbAndActionsMap map[string]DataAction

func NewVerbAndActionsMap() VerbAndActionsMap {
	return make(map[string]DataAction)
}

type ResourceAndVerbMap map[string]VerbAndActionsMap

func NewResourceAndVerbMap() ResourceAndVerbMap {
	return make(map[string]VerbAndActionsMap)
}

type OperationsMap map[string]ResourceAndVerbMap

func NewOperationsMap() OperationsMap {
	return make(map[string]ResourceAndVerbMap)
}

func DeepCopyOperationsMap() OperationsMap {
	operationsMapLock.RLock()
	defer operationsMapLock.RUnlock()
	if len(operationsMap) == 0 {
		return NewOperationsMap()
	}

	copyMap := NewOperationsMap()
	for k, v := range operationsMap {
		copyMap[k] = v
	}

	return copyMap
}

func (o OperationsMap) String() string {
	opMapString, _ := json.Marshal(o)
	return string(opMapString)
}

func ConvertIntToString(number int) string {
	return strconv.Itoa(number)
}

func SetDiscoverResourcesSettings(clusterType string, environment string, loginURL string, kubeconfigFilePath string, tenantID string, clientID string, clientSecret string) error {
	if settings == nil {
		settings = &DiscoverResourcesSettings{
			clusterType:        clusterType,
			aksLoginURL:        loginURL,
			kubeconfigFilePath: kubeconfigFilePath,
			tenantID:           tenantID,
			clientID:           clientID,
			clientSecret:       clientSecret,
		}

		env := azure.PublicCloud
		var err error
		if environment != "" {
			env, err = azure.EnvironmentFromName(environment)
			if err != nil {
				return errors.Wrap(err, "Failed to parse environment for Azure.")
			}
		}

		settings.environment = env

		switch clusterType {
		case ConnectedClusters:
			settings.operationsEndpoint = fmt.Sprintf(OperationsEndpointFormatARC, settings.environment.ResourceManagerEndpoint)
		case ManagedClusters:
			settings.operationsEndpoint = fmt.Sprintf(OperationsEndpointFormatAKS, settings.environment.ResourceManagerEndpoint)
		case Fleets:
			settings.operationsEndpoint = fmt.Sprintf(OperationsEndpointFormatAKS, settings.environment.ResourceManagerEndpoint)
		default:
			return errors.Errorf("Failed to create endpoint for Get Operations call. Cluster type %s is not supported.", clusterType)
		}

		return nil
	}
	return nil
}

// ReconcileDiscoverResources reconciles the operationsMap
func ReconcileDiscoverResources(ctx context.Context, wg *sync.WaitGroup, loopDuration time.Duration) {
	defer wg.Done()
	for {
		select {
		case <-time.After(loopDuration):
			discoverResourcesListStart := time.Now()
			err := DiscoverResources(ctx)
			discoverResourcesDuration := time.Since(discoverResourcesListStart).Seconds()
			if err != nil {
				code := http.StatusInternalServerError
				if v, ok := err.(errutils.HttpStatusCode); ok {
					code = v.Code()
				}
				counterDiscoverResources.WithLabelValues(ConvertIntToString(code)).Inc()
				DiscoverResourcesTotalDuration.Observe(discoverResourcesDuration)
				klog.Errorf("Failed to reconcile map of data actions. Error:%s", err)
			} else {
				counterDiscoverResources.WithLabelValues(ConvertIntToString(http.StatusOK)).Inc()
				DiscoverResourcesTotalDuration.Observe(discoverResourcesDuration)
			}
		case <-ctx.Done():
			return
		}
	}
}

/*
DiscoverResources does the following:
1. Fetches list of ApiResources from the apiserver
2. Fetches list of Data Actions via Get Operations call on Azure
3. creates OperationsMap which is a map of "group": { "resource": { "verb": DataAction{} } } }
This map is used to create list of AuthorizationActionInfos when we get a SAR request where Resource/Verb/Group is *
*/
func DiscoverResources(ctx context.Context) error {
	apiResourcesListStart := time.Now()
	apiResourcesList, err := fetchApiResources()
	apiResourcesListDuration := time.Since(apiResourcesListStart).Seconds()

	if err != nil {
		return errors.Wrap(err, "Failed to fetch list of api-resources from apiserver.")
	}

	discoverResourcesApiServerCallDuration.Observe(apiResourcesListDuration)

	getOperationsStart := time.Now()
	operationsList, err := fetchDataActionsList(ctx)
	getOperationsDuration := time.Since(getOperationsStart).Seconds()

	if err != nil {
		return errors.Wrap(err, "Failed to fetch operations from Azure.")
	}

	discoverResourcesAzureCallDuration.Observe(getOperationsDuration)

	createOperationsMap(apiResourcesList, operationsList)

	return nil
}

func createOperationsMap(apiResourcesList []*metav1.APIResourceList, operationsList []Operation) {
	operationsMapLock.Lock()
	defer operationsMapLock.Unlock()

	for _, resList := range apiResourcesList {
		if len(resList.APIResources) == 0 {
			continue
		}

		group := "v1" // core api group
		if resList.GroupVersion != "" && resList.GroupVersion != "v1" {
			group = strings.Split(resList.GroupVersion, "/")[0]
		}

		for _, apiResource := range resList.APIResources {
			if strings.Contains(apiResource.Name, "/") {
				continue
			}

			actionId := settings.clusterType
			if group != "v1" {
				actionId = path.Join(actionId, group)
			}

			resourceName := apiResource.Name

			actionId = path.Join(actionId, resourceName)

			for _, operation := range operationsList {
				if strings.Contains(operation.Name, actionId) {
					opNameArr := strings.Split(operation.Name, "/")

					/* The strings.contains check will return true for groups that have same prefix. For example:
					    Will return true for "Microsoft.ContainerService/managedCluster/events.k8s.io/events/.."
						and Microsoft.ContainerService/managedCluster/mc/events/.."  when:
						group = v1
						resource = events
						actionID = Microsoft.ContainerService/managedCluster/events/.."
						Without the below validation , the dataactions for events in events.k8s.io will get added in v1 map as well which
						will return the wrong data actions later in checkaccess
						So we need extra validation to check whether the group / resource are equal.
					*/
					if group != "v1" {
						// extra validation to make sure groups are the same
						if group != opNameArr[2] {
							continue
						}
					} else {
						// make sure resources are the same for core apigroup
						if resourceName != opNameArr[2] {
							continue
						}
					}

					verb := opNameArr[len(opNameArr)-1]
					if verb == "action" {
						verb = path.Join(opNameArr[len(opNameArr)-2], opNameArr[len(opNameArr)-1])
					}

					da := DataAction{
						ActionInfo: AuthorizationActionInfo{
							IsDataAction: true,
						},
						IsNamespacedResource: apiResource.Namespaced,
					}
					da.ActionInfo.AuthorizationEntity.Id = operation.Name
					if _, found := operationsMap[group]; !found {
						operationsMap[group] = NewResourceAndVerbMap()
					}

					if _, found := operationsMap[group][resourceName]; !found {
						operationsMap[group][resourceName] = NewVerbAndActionsMap()
					}

					operationsMap[group][resourceName][verb] = da
				}
			}
		}
	}

	klog.V(5).Infof("Operations Map created for resources: %s", operationsMap)
}

func fetchApiResources() ([]*metav1.APIResourceList, error) {
	// creates the in-cluster config
	klog.V(5).Infof("Fetching list of APIResources from the apiserver.")

	var cfg *rest.Config
	var err error
	if settings.kubeconfigFilePath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", settings.kubeconfigFilePath)
	} else {
		cfg, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error building kubeconfig")
	}

	kubeclientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error building kubernetes clientset")
	}

	apiresourcesList, err := kubeclientset.Discovery().ServerPreferredResources()

	if klog.V(5).Enabled() {
		printApiresourcesList, _ := json.Marshal(apiresourcesList)

		klog.Infof("List of ApiResources fetched from apiserver: %s", string(printApiresourcesList))
	}

	if err != nil {
		// ignoring unreachable apiservice error.
		if strings.Contains(err.Error(), retrieveServerApiError) && strings.Contains(err.Error(), apiserviceUnreachableError) {
			klog.Infof("Error while fetching apiservices from apiserver: %s", err.Error())
		} else {
			return nil, err
		}
	}

	return apiresourcesList, nil
}

func fetchDataActionsList(ctx context.Context) ([]Operation, error) {
	req, err := http.NewRequest(http.MethodGet, settings.operationsEndpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create request for Get Operations call.")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("guard-%s-%s-%s", v.Version.Platform, v.Version.GoVersion, v.Version.Version))

	var token string
	if settings.clusterType == ConnectedClusters {
		tokenProvider := graph.NewClientCredentialTokenProvider(settings.clientID, settings.clientSecret,
			fmt.Sprintf("%s%s/oauth2/v2.0/token", settings.environment.ActiveDirectoryEndpoint, settings.tenantID),
			fmt.Sprintf("%s/.default", settings.environment.ResourceManagerEndpoint))

		authResp, erro := tokenProvider.Acquire(ctx, "")
		if erro != nil {
			return nil, errors.Wrap(erro, "Error getting authorization headers for Get Operations call.")
		}

		token = authResp.Token
	} else { // AKS and Fleet
		tokenProvider := graph.NewAKSTokenProvider(settings.aksLoginURL, settings.tenantID)

		authResp, err := tokenProvider.Acquire(ctx, "")
		if err != nil {
			return nil, errors.Wrap(err, "Error getting authorization headers for Get Operations call.")
		}

		token = authResp.Token
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := httpclient.DefaultHTTPClient

	resp, err := client.Do(req)
	if err != nil {
		counterGetOperationsResources.WithLabelValues(ConvertIntToString(http.StatusInternalServerError)).Inc()
		return nil, errors.Wrap(err, "Failed to send request for Get Operations call.")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		counterGetOperationsResources.WithLabelValues(ConvertIntToString(http.StatusInternalServerError)).Inc()
		return nil, errors.Wrap(err, "Error in reading response body")
	}

	counterGetOperationsResources.WithLabelValues(ConvertIntToString(resp.StatusCode)).Inc()
	if resp.StatusCode != http.StatusOK {
		return nil, errutils.WithCode(errors.Errorf("Request failed with status code: %d and response: %s", resp.StatusCode, string(data)), resp.StatusCode)
	}

	operationsList := OperationList{}
	err = json.Unmarshal(data, &operationsList)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode response")
	}

	var finalOperations []Operation
	for _, op := range operationsList.Value {
		if *op.IsDataAction && strings.Contains(op.Name, settings.clusterType) {
			finalOperations = append(finalOperations, op)
		}
	}

	if klog.V(5).Enabled() {
		printFinalOperations, _ := json.Marshal(finalOperations)

		klog.Infof("List of Operations fetched from Azure %s", string(printFinalOperations))
	}

	return finalOperations, nil
}

// contextAwareLogger is a custom logger that uses klog's context-aware logging
type contextAwareLogger struct {
	ctx context.Context
}

func (l *contextAwareLogger) Printf(format string, v ...interface{}) {
	// Use klog's context-aware logging which automatically includes request ID and correlation ID
	klog.FromContext(l.ctx).Info(fmt.Sprintf(format, v...))
}

// MakeRetryableHttpClient creates an HTTP client which attempts the request
// (1 + retryCount) times and has a 3 second timeout per attempt.
func MakeRetryableHttpClient(retryCount int, ctx context.Context) retryablehttp.Client {
	// Copy the default HTTP client so we can set a timeout.
	// (It uses the same transport since the pointer gets copied)
	httpClient := *httpclient.DefaultHTTPClient
	httpClient.Timeout = 3 * time.Second

	// Attempt the request up to 3 times
	return retryablehttp.Client{
		HTTPClient:   &httpClient,
		RetryWaitMin: 500 * time.Millisecond,
		RetryWaitMax: 2 * time.Second,
		RetryMax:     retryCount, // initial + retryCount retries = (1 + retryCount) attempts
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
		Logger:       &contextAwareLogger{ctx: ctx},
	}
}

// WithRetryableHttpClient sets the oauth2.HTTPClient key of the context to an
// *http.Client made from makeRetryableHttpClient.
// Some of the libraries we use will take the client out of the context via
// oauth2.HTTPClient and use it, so this way we can add retries to external code.
func WithRetryableHttpClient(ctx context.Context, retryCount int) context.Context {
	retryClient := MakeRetryableHttpClient(retryCount, ctx)
	return context.WithValue(ctx, oauth2.HTTPClient, retryClient.StandardClient())
}

func LoadClientWithContext(ctx context.Context, defaultClient *http.Client) *http.Client {
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		return c
	}

	return defaultClient
}

func init() {
	prometheus.MustRegister(DiscoverResourcesTotalDuration, discoverResourcesAzureCallDuration, discoverResourcesApiServerCallDuration, counterDiscoverResources, counterGetOperationsResources)
}
