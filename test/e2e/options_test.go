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
package e2e_test

import (
	"flag"
	"path/filepath"

	"github.com/appscode/go/flags"
	logs "github.com/appscode/go/log/golog"

	"github.com/golang/glog"
	"k8s.io/client-go/util/homedir"
)

type E2EOptions struct {
	KubeContext string
	KubeConfig  string
}

var (
	options = &E2EOptions{
		KubeConfig:  filepath.Join(homedir.HomeDir(), ".kube", "config"),
		KubeContext: "minikube",
	}
)

func init() {
	flag.StringVar(&options.KubeConfig, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&options.KubeContext, "kube-context", "", "Name of kube context")
	enableLogging()
}

func enableLogging() {
	defer func() {
		logs.InitLogs()
		defer logs.FlushLogs()
	}()
	err := flag.Set("logtostderr", "true")
	if err != nil {
		glog.Fatal(err)
	}
	logLevelFlag := flag.Lookup("v")
	if logLevelFlag != nil {
		if len(logLevelFlag.Value.String()) > 0 && logLevelFlag.Value.String() != "0" {
			return
		}
	}
	flags.SetLogLevel(2)
}
