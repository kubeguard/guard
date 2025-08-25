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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
)

const (
	fleetResoureType = "Microsoft.ContainerService/fleets"
)

func ValidateFleetID(id string) error {
	r, err := arm.ParseResourceID(id)
	if err != nil {
		return err
	}
	if !strings.EqualFold(r.ResourceType.String(), fleetResoureType) {
		return fmt.Errorf("%s is not a fleet resource", id)
	}
	return nil
}
