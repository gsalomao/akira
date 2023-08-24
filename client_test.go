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

import "testing"

func TestClientStateString(t *testing.T) {
	testCases := []struct {
		name  string
		state ClientState
	}{
		{"Disconnected", ClientDisconnected},
		{"Authenticating", ClientAuthenticating},
		{"Connected", ClientConnected},
		{"Re-authenticating", ClientReAuthenticating},
		{"Invalid", ClientState(99)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name := tc.state.String()
			if name != tc.name {
				t.Errorf("Unexpected name\nwant: %s\ngot:  %s", tc.name, name)
			}
		})
	}
}
