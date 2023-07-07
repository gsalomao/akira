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

// ErrV3ClientIDRejected indicates that the CONNECT packet was rejected due to the client ID.
var ErrV3ClientIDRejected = Error{Code: ReasonCodeIdentifierRejected, Reason: "client id rejected"}

// MQTT malformed packet error.
var (
	ErrMalformedInteger = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "integer",
	}
	ErrMalformedString = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "string",
	}
	ErrMalformedVarInteger = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "variable integer",
	}
	ErrMalformedBinary = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "binary",
	}
	ErrMalformedPacketType = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "packet type",
	}
	ErrMalformedPacketLength = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "packet length",
	}
	ErrMalformedFlags = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "flags",
	}
	ErrMalformedProtocolName = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "protocol name",
	}
	ErrMalformedProtocolVersion = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "protocol version",
	}
	ErrMalformedConnectFlags = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "connect flags",
	}
	ErrMalformedKeepAlive = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "keep alive",
	}
	ErrMalformedClientID = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "client identifier",
	}
	ErrMalformedWillTopic = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "will topic",
	}
	ErrMalformedWillPayload = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "will payload",
	}
	ErrMalformedUsername = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "username",
	}
	ErrMalformedPassword = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "password",
	}
	ErrMalformedPropertyLength = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property length",
	}
	ErrMalformedPropertyInvalid = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "invalid property",
	}
	ErrMalformedPropertyConnect = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "connect property",
	}
	ErrMalformedPropertyWill = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "will property",
	}
	ErrMalformedPropertyReceiveMaximum = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: receive maximum",
	}
	ErrMalformedPropertyMaxPacketSize = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: max packet size",
	}
	ErrMalformedPropertyUserProperty = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: user property",
	}
	ErrMalformedPropertyContentType = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: content type",
	}
	ErrMalformedPropertyResponseTopic = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: response topic",
	}
	ErrMalformedPropertyCorrelationData = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: correlation data",
	}
	ErrMalformedPropertySessionExpiryInterval = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: session invalid interval",
	}
	ErrMalformedPropertyTopicAliasMaximum = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: topic alias maximum",
	}
	ErrMalformedPropertyRequestResponseInfo = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: request response info",
	}
	ErrMalformedPropertyRequestProblemInfo = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: request problem info",
	}
	ErrMalformedPropertyAuthenticationMethod = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: authenticating method",
	}
	ErrMalformedPropertyAuthenticationData = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: authenticating data",
	}
	ErrMalformedPropertyWillDelayInterval = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: will delay interval",
	}
	ErrMalformedPropertyPayloadFormatIndicator = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: payload format indicator",
	}
	ErrMalformedPropertyMessageExpiryInterval = Error{
		Code:   ReasonCodeMalformedPacket,
		Reason: "property: message expiry interval",
	}
)

// Error represents the errors related to the MQTT protocol.
type Error struct {
	// Reason is string with a human-friendly message about the error.
	Reason string

	// Code represents the error codes based on the MQTT specifications.
	Code ReasonCode
}

// Error returns a human-friendly message.
func (e Error) Error() string {
	return fmt.Sprintf("%s: %s (code=0x%.2x)", e.Code.String(), e.Reason, byte(e.Code))
}
