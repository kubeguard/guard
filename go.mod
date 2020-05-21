module github.com/appscode/guard

go 1.12

require (
	github.com/Azure/go-autorest/autorest v0.9.0
	github.com/appscode/go v0.0.0-20200323182826-54e98e09185a
	github.com/appscode/pat v0.0.0-20170521084856-48ff78925b79
	github.com/aws/aws-sdk-go v1.31.2
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-github/v25 v25.1.3
	github.com/google/gofuzz v1.1.0
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/json-iterator/go v1.1.8
	github.com/moul/http2curl v1.0.0
	github.com/nmcclain/asn1-ber v0.0.0-20170104154839-2661553a0484 // indirect
	github.com/nmcclain/ldap v0.0.0-20191021200707-3b3b69a7e9e3
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.0.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	github.com/xanzy/go-gitlab v0.31.0
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gomodules.xyz/cert v1.0.3
	google.golang.org/api v0.4.0
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/goidentity.v1 v1.0.0 // indirect
	gopkg.in/jcmturner/gokrb5.v4 v4.1.2
	gopkg.in/square/go-jose.v2 v2.2.2
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v0.18.3
	kmodules.xyz/client-go v0.0.0-20200521065424-173e32c78a20
)

replace (
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200520235721-10b58e57a423
	k8s.io/apiserver => github.com/kmodules/apiserver v0.18.4-0.20200521000930-14c5f6df9625
	k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200521033432-49d3646051ad
)
