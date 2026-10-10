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

package framework

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const portForwardReadyTimeout = 30 * time.Second

// PortForwardSession represents a running client-go pod port-forward session.
type PortForwardSession struct {
	LocalPort uint16

	stopCh chan struct{}
	doneCh chan error
	once   sync.Once
}

// Close stops the port-forward session and waits for the forwarding goroutine.
func (s *PortForwardSession) Close() error {
	if s == nil {
		return nil
	}

	var err error
	s.once.Do(func() {
		close(s.stopCh)
		err = <-s.doneCh
	})

	return err
}

// PortForwardFirstPod forwards a random local port to the requested remote pod
// port on the first Running pod matching the provided label selector.
func (f *Framework) PortForwardFirstPod(
	ctx context.Context,
	namespace, labelSelector string,
	remotePort uint16,
) (*PortForwardSession, error) {
	pod, err := f.getFirstRunningPod(ctx, namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	return f.PortForwardPod(ctx, namespace, pod.Name, remotePort)
}

// PortForwardPod opens an in-process SPDY port-forward session to the provided pod.
func (f *Framework) PortForwardPod(
	ctx context.Context,
	namespace, podName string,
	remotePort uint16,
) (*PortForwardSession, error) {
	if f.RestConfig == nil {
		return nil, fmt.Errorf("REST config is not initialized")
	}

	transport, upgrader, err := spdy.RoundTripperFor(f.RestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create port-forward transport: %w", err)
	}

	apiServerURL, err := url.Parse(f.RestConfig.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse REST host %q: %w", f.RestConfig.Host, err)
	}

	portForwardURL := &url.URL{
		Scheme: apiServerURL.Scheme,
		Host:   apiServerURL.Host,
		Path:   fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName),
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, portForwardURL)
	readyCh := make(chan struct{})
	stopCh := make(chan struct{})
	doneCh := make(chan error, 1)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	forwarder, err := portforward.NewOnAddresses(
		dialer,
		[]string{"127.0.0.1"},
		[]string{fmt.Sprintf("0:%d", remotePort)},
		stopCh,
		readyCh,
		stdout,
		stderr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create port-forward session: %w", err)
	}

	go func() {
		err := forwarder.ForwardPorts()
		if err != nil {
			if msg := strings.TrimSpace(stderr.String()); msg != "" {
				err = fmt.Errorf("%w: %s", err, msg)
			}
			doneCh <- err
			return
		}
		doneCh <- nil
	}()

	select {
	case <-readyCh:
		ports, err := forwarder.GetPorts()
		if err != nil {
			close(stopCh)
			<-doneCh
			return nil, fmt.Errorf("failed to read forwarded ports: %w", err)
		}
		if len(ports) == 0 {
			close(stopCh)
			<-doneCh
			return nil, fmt.Errorf("port-forward did not expose a local port")
		}
		return &PortForwardSession{LocalPort: ports[0].Local, stopCh: stopCh, doneCh: doneCh}, nil
	case err := <-doneCh:
		return nil, fmt.Errorf("port-forward failed before becoming ready: %w", err)
	case <-ctx.Done():
		close(stopCh)
		<-doneCh
		return nil, fmt.Errorf("port-forward canceled: %w", ctx.Err())
	case <-time.After(portForwardReadyTimeout):
		close(stopCh)
		<-doneCh
		return nil, fmt.Errorf("timed out waiting for pod port-forward to become ready")
	}
}

func (f *Framework) getFirstRunningPod(ctx context.Context, namespace, labelSelector string) (*corev1.Pod, error) {
	pods, err := f.KubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		if pods.Items[i].Status.Phase == corev1.PodRunning {
			return &pods.Items[i], nil
		}
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for label selector %q", labelSelector)
	}
	return nil, fmt.Errorf("no running pods found for label selector %q", labelSelector)
}
