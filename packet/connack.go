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

package packet

import (
	"errors"
	"fmt"
)

var validReasonCodes = [][]ReasonCode{
	// V3.1.
	{
		ReasonCodeSuccess, ReasonCodeV3UnacceptableProtocolVersion, ReasonCodeV3IdentifierRejected,
		ReasonCodeV3ServerUnavailable, ReasonCodeV3BadUsernameOrPassword, ReasonCodeV3NotAuthorized,
	},
	// V3.1.1.
	{
		ReasonCodeSuccess, ReasonCodeV3UnacceptableProtocolVersion, ReasonCodeV3IdentifierRejected,
		ReasonCodeV3ServerUnavailable, ReasonCodeV3BadUsernameOrPassword, ReasonCodeV3NotAuthorized,
	},
	// V5.0.
	{
		ReasonCodeSuccess, ReasonCodeUnspecifiedError, ReasonCodeMalformedPacket, ReasonCodeProtocolError,
		ReasonCodeImplementationSpecificError, ReasonCodeUnsupportedProtocolVersion, ReasonCodeClientIDNotValid,
		ReasonCodeBadUsernameOrPassword, ReasonCodeNotAuthorized, ReasonCodeServerUnavailable, ReasonCodeServerBusy,
		ReasonCodeBanned, ReasonCodeBadAuthenticationMethod, ReasonCodeTopicNameInvalid, ReasonCodePacketTooLarge,
		ReasonCodeQuotaExceeded, ReasonCodePayloadFormatInvalid, ReasonCodeRetainNotSupported,
		ReasonCodeQoSNotSupported, ReasonCodeUseAnotherServer, ReasonCodeServerMoved, ReasonCodeConnectionRateExceeded,
	},
}

// ConnAck represents the CONNACK Packet from MQTT specifications.
type ConnAck struct {
	// Properties contains the properties of the CONNACK packet.
	Properties *ConnAckProperties `json:"properties"`

	// Version represents the MQTT version.
	Version Version `json:"version"`

	// Code represents the reason code based on the MQTT specifications.
	Code ReasonCode `json:"code"`

	// SessionPresent indicates if there is already a session associated with the Client ID.
	SessionPresent bool `json:"session_present"`
}

// Type returns the packet type.
func (p *ConnAck) Type() Type {
	return TypeConnAck
}

// Size returns the CONNACK packet size.
func (p *ConnAck) Size() int {
	size := p.remainingLength()
	header := FixedHeader{PacketType: TypeConnAck, RemainingLength: size}

	size += header.size()
	return size
}

// Encode encodes the CONNACK Packet into buf and returns the number of bytes encoded. The buffer must have the
// length greater than or equals to the packet size, otherwise this method returns an error.
func (p *ConnAck) Encode(buf []byte) (n int, err error) {
	if len(buf) < p.Size() {
		return 0, errors.New("buffer too small")
	}

	err = p.Validate()
	if err != nil {
		return 0, err
	}

	header := FixedHeader{PacketType: TypeConnAck, RemainingLength: p.remainingLength()}
	n = header.encode(buf)

	var flags byte
	if p.SessionPresent && p.Code == ReasonCodeSuccess && p.Version != MQTT31 {
		flags = 1
	}

	buf[n] = flags
	n++

	buf[n] = byte(p.Code)
	n++

	if p.Version == MQTT50 {
		var size int

		size, err = p.Properties.encode(buf[n:])
		n += size
	}

	return n, err
}

// Validate validates if the CONNACK Packet is valid.
func (p *ConnAck) Validate() error {
	if p.Version < MQTT31 || p.Version > MQTT50 {
		return fmt.Errorf("%w: invalid version", ErrMalformedPacket)
	}

	var found bool
	for _, code := range validReasonCodes[p.Version-MQTT31] {
		if p.Code == code {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("%w: invalid reason code", ErrMalformedPacket)
	}

	if p.Version == MQTT50 {
		err := p.Properties.Validate()
		if err != nil {
			return fmt.Errorf("%w: invalid properties: %s ", ErrMalformedPacket, err.Error())
		}
	}

	return nil
}

func (p *ConnAck) remainingLength() int {
	// Start with 2 bytes for the ConnAck flags and Reason Code.
	remainingLength := 2

	if p.Version == MQTT50 {
		n := p.Properties.size()
		n += sizeVarInteger(n)
		remainingLength += n
	}

	return remainingLength
}

// ConnAckProperties contains the properties of the CONNACK packet.
type ConnAckProperties struct {
	// Flags indicates which properties are present.
	Flags PropertyFlags `json:"flags"`

	// UserProperties is a list of user properties.
	UserProperties []UserProperty `json:"user_properties"`

	// AssignedClientID represents the client ID assigned by the server in case of the client connected with the
	// server without specifying a client ID.
	AssignedClientID []byte `json:"assigned_client_id"`

	// ReasonString represents the reason associated with the response.
	ReasonString []byte `json:"reason_string"`

	// ResponseInfo contains a string that can be used to as the basis for creating a Response Topic.
	ResponseInfo []byte `json:"response_info"`

	// ServerReference contains a string indicating another server the client can use.
	ServerReference []byte `json:"server_reference"`

	// AuthenticationMethod contains the name of the authentication method.
	AuthenticationMethod []byte `json:"authentication_method"`

	// AuthenticationData contains the authentication data.
	AuthenticationData []byte `json:"authentication_data"`

	// SessionExpiryInterval represents the time, in seconds, which the server must store the Session State after the
	// network connection is closed.
	SessionExpiryInterval uint32 `json:"session_expiry_interval"`

	// MaximumPacketSize represents the maximum packet size, in bytes, the client is willing to accept.
	MaximumPacketSize uint32 `json:"maximum_packet_size"`

	// ReceiveMaximum represents the maximum number of inflight messages with QoS > 0.
	ReceiveMaximum uint16 `json:"receive_maximum"`

	// TopicAliasMaximum represents the highest number of Topic Alias that the client accepts.
	TopicAliasMaximum uint16 `json:"topic_alias_maximum"`

	// ServerKeepAlive represents the Keep Alive, in seconds, assigned by the server, and to be used by the client.
	ServerKeepAlive uint16 `json:"server_keep_alive"`

	// MaximumQoS represents the maximum QoS supported by the server.
	MaximumQoS byte `json:"maximum_qos"`

	// RetainAvailable indicates whether the server supports retained messages or not.
	RetainAvailable bool `json:"retain_available"`

	// WildcardSubscriptionAvailable indicates whether the server supports Wildcard Subscriptions or not.
	WildcardSubscriptionAvailable bool `json:"wildcard_subscription_available"`

	// SubscriptionIDAvailable indicates whether the server supports Subscription Identifiers or not.
	SubscriptionIDAvailable bool `json:"subscription_id_available"`

	// SharedSubscriptionAvailable indicates whether the server supports Shared Subscriptions or not.
	SharedSubscriptionAvailable bool `json:"shared_subscription_available"`
}

// Has returns whether the property is present or not.
func (p *ConnAckProperties) Has(id PropertyID) bool {
	if p != nil {
		return p.Flags.Has(id)
	}
	return false
}

// Set sets the property indicating that it's present.
func (p *ConnAckProperties) Set(id PropertyID) {
	if p != nil {
		p.Flags = p.Flags.Set(id)
	}
}

// Validate validates if the properties are valid.
func (p *ConnAckProperties) Validate() error {
	if p == nil {
		return nil
	}

	if err := validatePropString(p.Flags, PropertyAssignedClientID, p.AssignedClientID); err != nil {
		return errors.New("missing assigned client identifier")
	}
	if err := validatePropString(p.Flags, PropertyReasonString, p.ReasonString); err != nil {
		return errors.New("missing reason string")
	}
	if err := validatePropString(p.Flags, PropertyResponseInfo, p.ResponseInfo); err != nil {
		return errors.New("missing response info")
	}
	if err := validatePropString(p.Flags, PropertyServerReference, p.ServerReference); err != nil {
		return errors.New("missing server reference")
	}
	if err := validatePropString(p.Flags, PropertyAuthenticationMethod, p.AuthenticationMethod); err != nil {
		return errors.New("missing authentication method")
	}
	if err := validatePropString(p.Flags, PropertyAuthenticationData, p.AuthenticationData); err != nil {
		return errors.New("missing authentication data")
	}
	if err := validatePropUserProperty(p.Flags, p.UserProperties); err != nil {
		return err
	}
	if p.Flags.Has(PropertyMaximumQoS) && p.MaximumQoS > byte(QoS2) {
		return errors.New("invalid maximum qos")
	}
	return nil
}

func (p *ConnAckProperties) size() int {
	var size int

	if p != nil {
		size += sizePropSessionExpiryInterval(p.Flags)
		size += sizePropReceiveMaximum(p.Flags)
		size += sizePropMaxQoS(p.Flags)
		size += sizePropRetainAvailable(p.Flags)
		size += sizePropMaxPacketSize(p.Flags)
		size += sizePropAssignedClientID(p.Flags, p.AssignedClientID)
		size += sizePropTopicAliasMaximum(p.Flags)
		size += sizePropReasonString(p.Flags, p.ReasonString)
		size += sizePropUserProperties(p.Flags, p.UserProperties)
		size += sizePropWildcardSubscriptionAvailable(p.Flags)
		size += sizePropSubscriptionIDAvailable(p.Flags)
		size += sizePropSharedSubscriptionAvailable(p.Flags)
		size += sizePropServerKeepAlive(p.Flags)
		size += sizePropResponseInfo(p.Flags, p.ResponseInfo)
		size += sizePropServerReference(p.Flags, p.ServerReference)
		size += sizePropAuthenticationMethod(p.Flags, p.AuthenticationMethod)
		size += sizePropAuthenticationData(p.Flags, p.AuthenticationData)
	}

	return size
}

func (p *ConnAckProperties) encode(buf []byte) (n int, err error) {
	n = encodeVarInteger(buf, p.size())

	if p == nil {
		return n, nil
	}

	var size int

	size = encodePropUint(buf[n:], p.Flags, PropertySessionExpiryInterval, p.SessionExpiryInterval)
	n += size

	size = encodePropUint(buf[n:], p.Flags, PropertyReceiveMaximum, p.ReceiveMaximum)
	n += size

	size = encodePropUint(buf[n:], p.Flags, PropertyMaximumQoS, p.MaximumQoS)
	n += size

	size = encodePropBool(buf[n:], p.Flags, PropertyRetainAvailable, p.RetainAvailable)
	n += size

	size = encodePropUint(buf[n:], p.Flags, PropertyMaximumPacketSize, p.MaximumPacketSize)
	n += size

	size, err = encodePropString(buf[n:], p.Flags, PropertyAssignedClientID, p.AssignedClientID)
	n += size
	if err != nil {
		return n, err
	}

	size = encodePropUint(buf[n:], p.Flags, PropertyTopicAliasMaximum, p.TopicAliasMaximum)
	n += size

	size, err = encodePropString(buf[n:], p.Flags, PropertyReasonString, p.ReasonString)
	n += size
	if err != nil {
		return n, err
	}

	size, err = encodePropUserProperties(buf[n:], p.Flags, p.UserProperties, err)
	n += size
	if err != nil {
		return n, err
	}

	size = encodePropBool(buf[n:], p.Flags, PropertyWildcardSubscriptionAvailable, p.WildcardSubscriptionAvailable)
	n += size

	size = encodePropBool(buf[n:], p.Flags, PropertySubscriptionIDAvailable, p.SubscriptionIDAvailable)
	n += size

	size = encodePropBool(buf[n:], p.Flags, PropertySharedSubscriptionAvailable, p.SharedSubscriptionAvailable)
	n += size

	size = encodePropUint(buf[n:], p.Flags, PropertyServerKeepAlive, p.ServerKeepAlive)
	n += size

	size, err = encodePropString(buf[n:], p.Flags, PropertyResponseInfo, p.ResponseInfo)
	n += size
	if err != nil {
		return n, err
	}

	size, err = encodePropString(buf[n:], p.Flags, PropertyServerReference, p.ServerReference)
	n += size
	if err != nil {
		return n, err
	}

	size, err = encodePropString(buf[n:], p.Flags, PropertyAuthenticationMethod, p.AuthenticationMethod)
	n += size
	if err != nil {
		return n, err
	}

	size, err = encodePropString(buf[n:], p.Flags, PropertyAuthenticationData, p.AuthenticationData)
	n += size

	return n, err
}
