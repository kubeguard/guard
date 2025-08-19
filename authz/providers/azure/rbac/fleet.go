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

package rbac

import (
	"path"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

const (
	fleetResourceIDFile    = "fleet-resource-id"
	fleetResourceIDPattern = `(?i)^/subscriptions/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/resourceGroups/[-\w.()]{1,90}/providers/Microsoft\.ContainerService/fleets/[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$`
)

var fleetResrcExp = regexp.MustCompile(fleetResourceIDPattern)

func (a *AccessInfo) getFleetManagerResourceID() (bool, string) {
	if a.runtimeConfigPath == "" {
		return false, ""
	}
	fleetIDFilePath := path.Join(a.runtimeConfigPath, fleetResourceIDFile)
	configuredID, err := a.runtimeConfigReader(fleetIDFilePath)
	if err != nil {
		klog.Errorf("unable to read fleet resource Id from file %s. Error: %s", fleetIDFilePath, err.Error())
		return false, ""
	}
	id := strings.TrimSpace(string(configuredID))
	if ok := fleetResrcExp.MatchString(id); !ok {
		klog.Errorf("%s did not contain a fleet resource Id.", fleetResourceIDFile)
		return false, ""
	}
	return true, id
}
