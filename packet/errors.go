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

var (
	// ErrUnspecifiedError indicates that the server does not wish to reveal the reason for the failure, or none of the
	// other reasons apply.
	ErrUnspecifiedError = Error{"unspecified error", ReasonCodeUnspecifiedError}

	// ErrMalformedPacket indicates that data within the packet could not be correctly parsed.
	ErrMalformedPacket = Error{"malformed packet", ReasonCodeMalformedPacket}

	// ErrProtocolError indicates that data in the packet does not conform with the MQTT specification, or an unexpected
	// or out of order packet was received.
	ErrProtocolError = Error{"protocol error", ReasonCodeProtocolError}

	// ErrImplementationSpecificError indicates that the packet is valid but is not accepted by the server.
	ErrImplementationSpecificError = Error{"implementation specific error", ReasonCodeImplementationSpecificError}

	// ErrUnsupportedProtocolVersion indicates that the server does not support the version of the MQTT protocol
	// requested by the client.
	ErrUnsupportedProtocolVersion = Error{"unsupported protocol version", ReasonCodeUnsupportedProtocolVersion}

	// ErrClientIDNotValid indicates that the client identifier is a valid string but is not allowed by the server.
	ErrClientIDNotValid = Error{"client identifier not valid", ReasonCodeClientIDNotValid}

	// ErrBadUsernameOrPassword indicates that the server does not accept the username or password specified by the
	// client.
	ErrBadUsernameOrPassword = Error{"bad username or password", ReasonCodeBadUsernameOrPassword}

	// ErrNotAuthorized indicates that the client is not authorized to perform the operation.
	ErrNotAuthorized = Error{"not authorized", ReasonCodeNotAuthorized}

	// ErrServerUnavailable indicates that the server is not available at the moment.
	ErrServerUnavailable = Error{"server unavailable", ReasonCodeServerUnavailable}

	// ErrServerBusy indicates that the server is busy at the moment.
	ErrServerBusy = Error{"server busy", ReasonCodeServerBusy}

	// ErrBanned indicates that the client has been banned by administrative action.
	ErrBanned = Error{"banned", ReasonCodeBanned}

	// ErrServerShuttingDown indicates that the server is shutting down.
	ErrServerShuttingDown = Error{"server shutting down", ReasonCodeServerShuttingDown}

	// ErrBadAuthenticationMethod indicates that the authentication method is not supported or does not match the
	// authentication method currently in use.
	ErrBadAuthenticationMethod = Error{"bad authentication method", ReasonCodeBadAuthenticationMethod}

	// ErrKeepAliveTimeout indicates that connection is closed because no packet has been received for 1.5 times the
	// keep alive time.
	ErrKeepAliveTimeout = Error{"keep alive timeout", ReasonCodeKeepAliveTimeout}

	// ErrSessionTakenOver indicates that another client using the same client identified has connected.
	ErrSessionTakenOver = Error{"session taken over", ReasonCodeSessionTakenOver}

	// ErrTopicFilterInvalid indicates that the topic filter is correctly formed but is not allowed.
	ErrTopicFilterInvalid = Error{"topic filter invalid", ReasonCodeTopicFilterInvalid}

	// ErrTopicNameInvalid indicates that the topic name is not malformed, but it is not accepted by the server.
	ErrTopicNameInvalid = Error{"topic name invalid", ReasonCodeTopicNameInvalid}

	// ErrPacketIDInUse indicates that the specified packet identifier is already in use.
	ErrPacketIDInUse = Error{"packet identifier in use", ReasonCodePacketIDInUse}

	// ErrPacketIDNotFound indicates that the packet identifier is unknown.
	ErrPacketIDNotFound = Error{"packet identifier not found", ReasonCodePacketIDNotFound}

	// ErrReceiveMaximumExceeded indicates that server has received more than Receive Maximum publication for which it
	// has not sent PUBACK or PUBCOMP.
	ErrReceiveMaximumExceeded = Error{"receive maximum exceeded", ReasonCodeReceiveMaximumExceeded}

	// ErrTopicAliasInvalid indicates that the server has received a PUBLISH Packet containing a topic alias which is
	// greater than the maximum topic alias sent in the CONNECT or CONNACK packet.
	ErrTopicAliasInvalid = Error{"topic alias invalid", ReasonCodeTopicAliasInvalid}

	// ErrPacketTooLarge indicates that the packet exceeded the maximum permissible size.
	ErrPacketTooLarge = Error{"packet too large", ReasonCodePacketTooLarge}

	// ErrMessageRateTooHigh indicates that the received data rate is too high.
	ErrMessageRateTooHigh = Error{"message rate too high", ReasonCodeMessageRateTooHigh}

	// ErrQuotaExceeded indicates that an implementation or administrative imposed limit has been exceeded.
	ErrQuotaExceeded = Error{"quota exceeded", ReasonCodeQuotaExceeded}

	// ErrAdministrativeAction indicates that the connection is closed due to an administrative action.
	ErrAdministrativeAction = Error{"administrative action", ReasonCodeAdministrativeAction}

	// ErrPayloadFormatInvalid indicates that the payload format does not match the specified Payload Format Indicator.
	ErrPayloadFormatInvalid = Error{"payload format invalid", ReasonCodePayloadFormatInvalid}

	// ErrRetainNotSupported indicates that the server does not support retained messages.
	ErrRetainNotSupported = Error{"retain not supported", ReasonCodeRetainNotSupported}

	// ErrQoSNotSupported indicates that the server does not support the specified QoS.
	ErrQoSNotSupported = Error{"qos not supported", ReasonCodeQoSNotSupported}

	// ErrUseAnotherServer indicates that the client should temporarily use another server.
	ErrUseAnotherServer = Error{"use another server", ReasonCodeUseAnotherServer}

	// ErrServerMoved indicates that the client should permanently use another server.
	ErrServerMoved = Error{"server moved", ReasonCodeServerMoved}

	// ErrSharedSubscriptionNotSupported indicates that the server does not support Shared Subscriptions.
	ErrSharedSubscriptionNotSupported = Error{
		"shared subscription not supported",
		ReasonCodeSharedSubscriptionsNotSupported,
	}

	// ErrConnectionRateExceeded indicates that the connection rate limit has been exceeded.
	ErrConnectionRateExceeded = Error{"connection rate exceeded", ReasonCodeConnectionRateExceeded}

	// ErrMaximumConnectTime indicates that the maximum connection time authorized for this connection has been
	// exceeded.
	ErrMaximumConnectTime = Error{"maximum connect time", ReasonCodeMaximumConnectTime}

	// ErrSubscriptionIDNotSupported indicates that the server does not support Subscription Identifiers.
	ErrSubscriptionIDNotSupported = Error{
		"subscription identifiers not supported",
		ReasonCodeSubscriptionIDNotSupported,
	}

	// ErrWildcardSubscriptionsNotSupported indicates that the server does not support Wildcard Subscriptions.
	ErrWildcardSubscriptionsNotSupported = Error{
		"wildcard subscriptions not supported",
		ReasonCodeWildcardSubscriptionsNotSupported,
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
	return e.Reason
}
