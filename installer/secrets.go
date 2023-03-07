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

package installer

import (
	"os"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func newProxySecret(namespace string, httpsProxy string, httpProxy string, noProxy string) runtime.Object {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-proxy",
			Namespace: namespace,
			Labels:    labels,
		},
		Type: core.SecretTypeOpaque,
		Data: map[string][]byte{
			"HTTP_PROXY":  []byte(httpProxy),
			"HTTPS_PROXY": []byte(httpsProxy),
			"NO_PROXY":    []byte(noProxy),
		},
	}
}

func newProxyCertSecret(namespace string, proxyCert string) (runtime.Object, error) {
	cert, err := os.ReadFile(proxyCert)
	if err != nil {
		return nil, err
	}

	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guard-proxy-cert",
			Namespace: namespace,
			Labels:    labels,
		},
		Type: core.SecretTypeOpaque,
		Data: map[string][]byte{
			"proxy-cert.crt": cert,
		},
	}, nil
}
