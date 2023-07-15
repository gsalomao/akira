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

// Reason Codes for MQTT V5.0.
const (
	// ReasonCodeSuccess indicates that the connection was accepted.
	ReasonCodeSuccess ReasonCode = 0x00

	// ReasonCodeNormalDisconnection indicates to close the connection normally and do not send the Will Message.
	ReasonCodeNormalDisconnection ReasonCode = 0x00

	// ReasonCodeGrantedQoS0 indicates that the subscription was accepted and the maximum QoS sent will be QoS 0.
	ReasonCodeGrantedQoS0 ReasonCode = 0x00

	// ReasonCodeGrantedQoS1 indicates that the subscription was accepted and the maximum QoS sent will be QoS 1.
	ReasonCodeGrantedQoS1 ReasonCode = 0x01

	// ReasonCodeGrantedQoS2 indicates that the subscription was accepted and any received QoS will be sent.
	ReasonCodeGrantedQoS2 ReasonCode = 0x02

	// ReasonCodeDisconnectWithWillMessage indicates to close the connection, but the server shall send the Will
	// Message.
	ReasonCodeDisconnectWithWillMessage ReasonCode = 0x04

	// ReasonCodeNoMatchingSubscribers indicates that the message was accepted but there are no subscribers.
	ReasonCodeNoMatchingSubscribers ReasonCode = 0x10

	// ReasonCodeNoSubscriptionExisted indicates that no matching Topic Filter is being used.
	ReasonCodeNoSubscriptionExisted ReasonCode = 0x11

	// ReasonCodeContinueAuthentication indicates to continue the authentication with another step.
	ReasonCodeContinueAuthentication ReasonCode = 0x18

	// ReasonCodeReAuthenticate indicates to initiate a re-authentication.
	ReasonCodeReAuthenticate ReasonCode = 0x19

	// ReasonCodeUnspecifiedError indicates that the server does not wish to reveal the reason for the failure, or none
	// of the other reason codes apply.
	ReasonCodeUnspecifiedError ReasonCode = 0x80

	// ReasonCodeMalformedPacket indicates that data within the Connect packet could not be correctly parsed.
	ReasonCodeMalformedPacket ReasonCode = 0x81

	// ReasonCodeProtocolError indicates that data in the Connect packet does not conform to the MQTT specification.
	ReasonCodeProtocolError ReasonCode = 0x82

	// ReasonCodeImplementationSpecificError indicates that the Connect is valid but is not accepted by this server.
	ReasonCodeImplementationSpecificError ReasonCode = 0x83

	// ReasonCodeUnsupportedProtocolVersion indicates that does not support the version of the MQTT protocol
	// requested by the client.
	ReasonCodeUnsupportedProtocolVersion ReasonCode = 0x84

	// ReasonCodeInvalidClientID indicates that the client identifier is a valid string but is not allowed.
	ReasonCodeInvalidClientID ReasonCode = 0x85

	// ReasonCodeBadUsernameOrPassword indicates that the server does not accept the username or password specified by
	// the client.
	ReasonCodeBadUsernameOrPassword ReasonCode = 0x86

	// ReasonCodeNotAuthorized indicates that the client is not authorized to connect.
	ReasonCodeNotAuthorized ReasonCode = 0x87

	// ReasonCodeServerUnavailable indicates that the server is not available.
	ReasonCodeServerUnavailable ReasonCode = 0x88

	// ReasonCodeServerBusy indicates that the server is busy at the moment.
	ReasonCodeServerBusy ReasonCode = 0x89

	// ReasonCodeBanned indicates that the client has been banned by administrative action.
	ReasonCodeBanned ReasonCode = 0x8A

	// ReasonCodeServerShuttingDown indicates that the server is shutting down.
	ReasonCodeServerShuttingDown ReasonCode = 0x8B

	// ReasonCodeBadAuthenticationMethod indicates that the authentication method is not supported or does not match
	// the authentication method currently in use.
	ReasonCodeBadAuthenticationMethod ReasonCode = 0x8C

	// ReasonCodeKeepAliveTimeout indicates that connection is closed because no packet has been received for 1.5
	// times the Keepalive time.
	ReasonCodeKeepAliveTimeout ReasonCode = 0x8D

	// ReasonCodeSessionTakeOver indicates that another connection using the same client identified has connected.
	ReasonCodeSessionTakeOver ReasonCode = 0x8E

	// ReasonCodeTopicFilterInvalid indicates that the topic filter is correctly formed but is not allowed.
	ReasonCodeTopicFilterInvalid ReasonCode = 0x8F

	// ReasonCodeTopicNameInvalid indicates that the Will Topic Name is not malformed, but is not accepted.
	ReasonCodeTopicNameInvalid ReasonCode = 0x90

	// ReasonCodePacketIDInUse indicates that the specified packet ID is already in use.
	ReasonCodePacketIDInUse ReasonCode = 0x91

	// ReasonCodePacketIDNotFound indicates that the packet ID was not found.
	ReasonCodePacketIDNotFound ReasonCode = 0x92

	// ReasonCodeReceiveMaximumExceeded indicates that has been received more than Receive Maximum publication for
	// which it has not sent PUBACK or PUBCOMP.
	ReasonCodeReceiveMaximumExceeded ReasonCode = 0x93

	// ReasonCodeTopicAliasInvalid indicates that has received a PUBLISH Packet containing a topic alias which is
	// greater than the maximum topic alias sent in the CONNECT or CONNACK packet.
	ReasonCodeTopicAliasInvalid ReasonCode = 0x94

	// ReasonCodePacketTooLarge indicates that the Connect packet exceeded the maximum permissible size.
	ReasonCodePacketTooLarge ReasonCode = 0x95

	// ReasonCodeMessageRateTooHigh indicates that the received data rate is too high.
	ReasonCodeMessageRateTooHigh ReasonCode = 0x96

	// ReasonCodePacketQuotaExceeded indicates that an implementation or administrative imposed limit has been exceeded.
	ReasonCodePacketQuotaExceeded ReasonCode = 0x97

	// ReasonCodeAdministrativeAction indicates that the connection is closed by administrative action.
	ReasonCodeAdministrativeAction ReasonCode = 0x98

	// ReasonCodePayloadFormatInvalid indicates that the Will Payload does not match the specified Payload Format
	// Indicator.
	ReasonCodePayloadFormatInvalid ReasonCode = 0x99

	// ReasonCodeRetainNotSupported indicates that the server does not support retained messages, and Will Retain was
	// set to 1.
	ReasonCodeRetainNotSupported ReasonCode = 0x9A

	// ReasonCodeQoSNotSupported indicates that the server does not support the QoS set in Will QoS.
	ReasonCodeQoSNotSupported ReasonCode = 0x9B

	// ReasonCodeUseAnotherServer indicates that the client should temporarily use another server.
	ReasonCodeUseAnotherServer ReasonCode = 0x9C

	// ReasonCodeServerMoved indicates that the client should permanently use another server.
	ReasonCodeServerMoved ReasonCode = 0x9D

	// ReasonCodeSharedSubscriptionsNotSupported indicates that the server does not support Shared Subscriptions.
	ReasonCodeSharedSubscriptionsNotSupported ReasonCode = 0x9E

	// ReasonCodeConnectionRateExceeded indicates that the connection rate limit has been exceeded.
	ReasonCodeConnectionRateExceeded ReasonCode = 0x9F

	// ReasonCodeMaximumConnectTime indicates that the maximum connection time authorized for this connection has been
	// exceeded.
	ReasonCodeMaximumConnectTime ReasonCode = 0xA0

	// ReasonCodeSubscriptionIDNotSupported indicates that the server does not support Subscription Identifiers.
	ReasonCodeSubscriptionIDNotSupported ReasonCode = 0xA1

	// ReasonCodeWildcardSubscriptionsNotSupported indicates that the server does not support Wildcard Subscriptions.
	ReasonCodeWildcardSubscriptionsNotSupported ReasonCode = 0xA2
)

// Reason Codes for MQTT V3.x.
const (
	// ReasonCodeV3UnacceptableProtocolVersion indicates that the server does not support the level of the MQTT
	// protocol requested by the Client.
	ReasonCodeV3UnacceptableProtocolVersion ReasonCode = 0x01

	// ReasonCodeV3IdentifierRejected indicates that the client identifier is a correct but not allowed.
	ReasonCodeV3IdentifierRejected ReasonCode = 0x02

	// ReasonCodeV3ServerUnavailable indicates that the network connection has been made but the MQTT service is
	// unavailable.
	ReasonCodeV3ServerUnavailable ReasonCode = 0x03

	// ReasonCodeV3BadUsernameOrPassword indicates that the data in the username or password is malformed.
	ReasonCodeV3BadUsernameOrPassword ReasonCode = 0x04

	// ReasonCodeV3NotAuthorized indicates that the client is not authorized to connect.
	ReasonCodeV3NotAuthorized ReasonCode = 0x05
)

// ReasonCode is a one byte unsigned value that indicates the result of an operation.
type ReasonCode byte
