// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lambdalifecycle

import "os"

type InitType int

const (
	OnDemand InitType = iota
	ProvisionedConcurrency
	SnapStart
	LambdaManagedInstances
	Unknown InitType = -1
)

func (t InitType) String() string {
	switch t {
	case OnDemand:
		return "on-demand"
	case ProvisionedConcurrency:
		return "provisioned-concurrency"
	case SnapStart:
		return "snap-start"
	case LambdaManagedInstances:
		return "lambda-managed-instances"
	default:
		return "unknown"
	}
}

func ParseInitType(s string) InitType {
	switch s {
	case "on-demand":
		return OnDemand
	case "provisioned-concurrency":
		return ProvisionedConcurrency
	case "snap-start":
		return SnapStart
	case "lambda-managed-instances":
		return LambdaManagedInstances
	default:
		return Unknown
	}
}

func InitTypeFromEnv(envVar string) InitType {
	return ParseInitType(os.Getenv(envVar))
}
