package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/appscode/go/ioutil"
	"github.com/appscode/go/term"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

func AddAuthInfoToKubeConfig(email string, authInfo *clientcmdapi.AuthInfo) error {
	var konfig *clientcmdapi.Config
	if _, err := os.Stat(KubeConfigPath()); err == nil {
		// ~/.kube/config exists
		konfig, err = clientcmd.LoadFromFile(KubeConfigPath())
		if err != nil {
			return err
		}

		bakFile := KubeConfigPath() + ".bak." + time.Now().Format("2006-01-02T15-04")
		err = ioutil.CopyFile(bakFile, KubeConfigPath())
		if err != nil {
			return err
		}
		term.Infoln(fmt.Sprintf("Current Kubeconfig is backed up as %s.", bakFile))
	} else {
		konfig = &clientcmdapi.Config{
			APIVersion: "v1",
			Kind:       "Config",
			Preferences: clientcmdapi.Preferences{
				Colors: true,
			},
		}
	}

	konfig.AuthInfos[email] = authInfo

	err := os.MkdirAll(filepath.Dir(KubeConfigPath()), 0755)
	if err != nil {
		return err
	}
	err = clientcmd.WriteToFile(*konfig, KubeConfigPath())
	if err != nil {
		return err
	}
	term.Successln("Configuration has been written to", KubeConfigPath())
	return nil
}

func KubeConfigPath() string {
	return homedir.HomeDir() + "/.kube/config"
}
