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

const (
	// ReasonCodeSuccess indicates that the operation was completed with success.
	ReasonCodeSuccess ReasonCode = 0x00

	// ReasonCodeConnectionAccepted indicates that the connection was accepted.
	ReasonCodeConnectionAccepted ReasonCode = 0x00

	// ReasonCodeIdentifierRejected indicates that the client identifier is correct UTF-8 but not allowed.
	ReasonCodeIdentifierRejected ReasonCode = 0x02

	// ReasonCodeMalformedPacket indicates that data within the packet could not be parsed correctly.
	ReasonCodeMalformedPacket ReasonCode = 0x81
)

// ReasonCode is a one byte unsigned value that indicates the result of an operation, based on the MQTT specifications.
type ReasonCode byte

// String returns the human-friendly name of the reason code.
func (c ReasonCode) String() string {
	str := "unknown reason code"
	switch c {
	case ReasonCodeIdentifierRejected:
		str = "identifier rejected"
	case ReasonCodeMalformedPacket:
		str = "malformed packet"
	default:
	}
	return str
}
