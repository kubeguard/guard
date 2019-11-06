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
package eks

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

const (
	OrgType         = "eks"
	v1Prefix        = "k8s-aws-v1."
	clusterIDHeader = "x-k8s-aws-id"
)

// "EKS" provider is not supported like other ones.
//func init() {
//	auth.SupportedOrgs = append(auth.SupportedOrgs, OrgType)
//}

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
