package kubeconfig

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

func AddAuthInfo(username string, authInfo *clientcmdapi.AuthInfo) error {
	var konfig *clientcmdapi.Config
	if _, err := os.Stat(Path()); err == nil {
		// ~/.kube/config exists
		konfig, err = clientcmd.LoadFromFile(Path())
		if err != nil {
			return err
		}

		bakFile := Path() + ".bak." + time.Now().Format("2006-01-02T15-04")
		err = ioutil.CopyFile(bakFile, Path())
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
			AuthInfos: map[string]*clientcmdapi.AuthInfo{},
		}
	}

	konfig.AuthInfos[username] = authInfo

	err := os.MkdirAll(filepath.Dir(Path()), 0755)
	if err != nil {
		return err
	}
	err = clientcmd.WriteToFile(*konfig, Path())
	if err != nil {
		return err
	}
	term.Successln("Configuration has been written to", Path())
	return nil
}

func Path() string {
	return filepath.Join(homedir.HomeDir(), ".kube", "config")
}
