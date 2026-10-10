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
	. "github.com/onsi/gomega"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert/certstore"
	"gomodules.xyz/x/crypto/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kmodules.xyz/client-go/tools/clientcmd"
)

type Framework struct {
	KubeClient kubernetes.Interface
	RestConfig *rest.Config
	namespace  string
	CertStore  *certstore.CertStore
}

func New(kubeConfigPath, kubeContext string) *Framework {
	restConfig, err := clientcmd.BuildConfigFromContext(kubeConfigPath, kubeContext)
	Expect(err).NotTo(HaveOccurred())

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	Expect(err).NotTo(HaveOccurred())

	store, err := certstore.New(blobfs.NewInMemoryFS(), "pki")
	Expect(err).NotTo(HaveOccurred())

	err = store.InitCA()
	Expect(err).NotTo(HaveOccurred())

	return &Framework{
		KubeClient: kubeClient,
		RestConfig: restConfig,
		namespace:  rand.WithUniqSuffix("test-guard"),
		CertStore:  store,
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       "guard",
	}
}

type Invocation struct {
	*Framework
	app string
}
