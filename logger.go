// Copyright 2023 Gustavo Salomao
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

package akira

import "fmt"

// Logger is the interface responsible for logging. The Log method logs the given msg with a variable number of fields.
type Logger interface {
	Log(msg string, fields ...any)
}

// noOpLogger is a logger which does nothing.
type noOpLogger struct{}

// Log does not log any message.
func (*noOpLogger) Log(msg string, fields ...any) {
	if len(fields)%2 > 0 {
		panic(fmt.Sprintf("log with message '%s' has odd number of fields: %v", msg, len(fields)))
	}
}
