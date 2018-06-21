package aws

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/appscode/guard/auth"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

const (
	OrgType         = "eks"
	v1Prefix        = "k8s-aws-v1."
	clusterIDHeader = "x-k8s-aws-id"
)

func init() {
	auth.SupportedOrgs = append(auth.SupportedOrgs, OrgType)
}

// https://github.com/heptio/aws-iam-authenticator/blob/master/pkg/token/token.go#L196
func Get(clusterID string) (string, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		AssumeRoleTokenProvider: StdinStderrTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	})
	if err != nil {
		return "", errors.Errorf("could not create session: %v", err)
	}

	stsAPI := sts.New(sess)

	request, _ := stsAPI.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	request.HTTPRequest.Header.Add(clusterIDHeader, clusterID)

	presignedURLString, err := request.Presign(60 * time.Second)
	if err != nil {
		return "", err
	}

	return v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(presignedURLString)), nil
}

func StdinStderrTokenProvider() (string, error) {
	var v string
	fmt.Fprint(os.Stderr, "Assume Role MFA token code: ")
	_, err := fmt.Scanln(&v)
	return v, err
}
