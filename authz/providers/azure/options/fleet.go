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

package options

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
)

const (
	fleetResoureType = "Microsoft.ContainerService/fleets"
	fleetNamePattern = `^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$`
)

var regex = regexp.MustCompile(fleetNamePattern)

func ValidateFleetID(id string) error {
	r, err := arm.ParseResourceID(id)
	if err != nil {
		return err
	}
	if !strings.EqualFold(r.ResourceType.String(), fleetResoureType) {
		return fmt.Errorf("%s is not a fleet resource", id)
	}
	if !regex.MatchString(r.Name) {
		return fmt.Errorf("%s does not have a valid fleet manager name", id)
	}
	return nil
}
