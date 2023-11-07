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

// Listener interface used to notify objects of Lambda lifecycle events.
type Listener interface {
	// FunctionInvoked is called after the extension receives a "Next" notification.
	FunctionInvoked()
	// FunctionFinished is called after the extension is notified that the function has completed, but before the environment is frozen.
	// The environment is only frozen once all listeners have returned.
	FunctionFinished()
	// EnvironmentShutdown is called when the extension is notified that the environment is about to shut down.
	// Shutting down of the collector components only happens after all listeners have returned.
	EnvironmentShutdown()
}

type Notifier interface {
	AddListener(listener Listener)
}

var (
	notifier Notifier
)

func SetNotifier(n Notifier) {
	notifier = n
}

func GetNotifier() Notifier {
	return notifier
}
