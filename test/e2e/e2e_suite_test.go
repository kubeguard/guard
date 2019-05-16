package e2e_test

import (
	"testing"
	"time"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/guard/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"kmodules.xyz/client-go/tools/clientcmd"
)

const (
	TIMEOUT = 20 * time.Minute
)

var (
	root *framework.Framework
)

func TestE2e(t *testing.T) {
	logs.InitLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	clientConfig, err := clientcmd.BuildConfigFromContext(options.KubeConfig, options.KubeContext)
	Expect(err).NotTo(HaveOccurred())

	client, err := kubernetes.NewForConfig(clientConfig)
	Expect(err).NotTo(HaveOccurred())

	root = framework.New(client)
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())
	By("Using test namespace " + root.Namespace())

	// clean up before test starts
	_, err = root.GetService("guard", root.Namespace())
	if err == nil {
		Expect(root.DeleteService("guard", root.Namespace())).NotTo(HaveOccurred())
	}

	_, err = root.GetDeployment("guard", root.Namespace())
	if err == nil {
		Expect(root.DeleteDeployment("guard", root.Namespace())).NotTo(HaveOccurred())
	}

	_, err = root.GetClusterRole("guard")
	if err == nil {
		Expect(root.DeleteClusterRole("guard")).NotTo(HaveOccurred())
	}

	_, err = root.GetClusterRoleBinding("guard")
	if err == nil {
		Expect(root.DeleteClusterRoleBinding("guard")).NotTo(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	root.DeleteNamespace()
})
