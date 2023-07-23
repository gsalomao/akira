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

import "fmt"

const (
	connectFlagReserved     = 0x01
	connectFlagCleanSession = 0x02
	connectFlagWillFlag     = 0x04
	connectFlagWillQoS      = 0x18
	connectFlagWillRetain   = 0x20
	connectFlagPasswordFlag = 0x40
	connectFlagUsernameFlag = 0x80
)

const (
	connectFlagShiftReserved = +iota
	connectFlagShiftCleanSession
	connectFlagShiftWillFlag
	connectFlagShiftWillQoS
	connectFlagShiftWillRetain
	connectFlagShiftPasswordFlag
	connectFlagShiftUsernameFlag
)

var protocolNames = []string{"MQIsdp", "MQTT", "MQTT"}

// Connect represents the CONNECT Packet from MQTT specifications.
type Connect struct {
	// ClientID represents the client identifier.
	ClientID []byte `json:"client_id"`

	// WillTopic represents the topic which the Will Payload will be published.
	WillTopic []byte `json:"will_topic"`

	// WillPayload represents the Will Payload to be published.
	WillPayload []byte `json:"will_payload"`

	// Username represents the username which the server must use for authentication and authorization.
	Username []byte `json:"username"`

	// Password represents the password which the server must use for authentication and authorization.
	Password []byte `json:"password"`

	// Properties contains the properties of the CONNECT packet.
	Properties *ConnectProperties `json:"properties"`

	// WillProperties contains the Will properties.
	WillProperties *WillProperties `json:"will_properties"`

	// KeepAlive is a time interval, measured in seconds, that is permitted to elapse between the point at which
	// the client finishes transmitting one control packet and the point it starts sending the next.
	KeepAlive uint16 `json:"keep_alive"`

	// Version represents the MQTT version.
	Version Version `json:"version"`

	// Flags represents the Connect flags.
	Flags ConnectFlags `json:"flags"`
}

// Type returns the packet type.
func (p *Connect) Type() Type {
	return TypeConnect
}

// Size returns the CONNECT packet size.
func (p *Connect) Size() int {
	// The protocolNames is an array and the protocol version starts with number 3 for MQTT V3.1. Based on that, the
	// protocol version is decremented by the MQTT V3.1 version number to get a value ranging from 0..2 for MQTT V3.1,
	// V3.1.1, and V5.0.
	size := sizeString(protocolNames[p.Version-MQTT31])

	// Add +1 byte for the protocol version, +1 byte for the Connect flags, and +2 bytes for the Keep Alive.
	size += 4

	if p.Version == MQTT50 {
		n := p.Properties.size()
		n += sizeVarInteger(n)
		size += n
	}

	size += sizeBinary(p.ClientID)

	if p.Flags.WillFlag() {
		if p.Version == MQTT50 {
			n := p.WillProperties.size()
			n += sizeVarInteger(n)
			size += n
		}

		size += sizeBinary(p.WillTopic)
		size += sizeBinary(p.WillPayload)
	}

	if p.Flags.Username() {
		size += sizeBinary(p.Username)
	}

	if p.Flags.Password() {
		size += sizeBinary(p.Password)
	}

	header := FixedHeader{PacketType: TypeConnect, RemainingLength: size}
	size += header.size()

	return size
}

// Decode decodes the CONNECT Packet from buf and header. This method returns the number of bytes read
// from buf and the error, if it fails to read the packet correctly.
func (p *Connect) Decode(buf []byte, h FixedHeader) (n int, err error) {
	err = p.validateHeader(buf, h)
	if err != nil {
		return 0, err
	}

	var cnt int

	cnt, err = p.decodeVersion(buf)
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodeFlags(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodeKeepAlive(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodeProperties(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodeClientID(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodeWillProperties(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodeWill(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodeUsername(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	cnt, err = p.decodePassword(buf[n:])
	if err != nil {
		return 0, err
	}
	n += cnt

	return n, nil
}

func (p *Connect) validateHeader(buf []byte, h FixedHeader) error {
	if h.PacketType != TypeConnect {
		return fmt.Errorf("%w: invalid packet type", ErrMalformedPacket)
	}
	if h.Flags != 0 {
		return fmt.Errorf("%w: invalid control flags", ErrMalformedPacket)
	}
	if len(buf) != h.RemainingLength {
		return fmt.Errorf("%w: buffer length does not match remaining length", ErrMalformedPacket)
	}
	return nil
}

func (p *Connect) decodeVersion(buf []byte) (int, error) {
	name, n, err := decodeString(buf)
	if err != nil {
		return n, fmt.Errorf("%w: mssing protocol name", ErrMalformedPacket)
	}
	if n == len(buf) {
		return n, fmt.Errorf("%w: missing protocol version", ErrMalformedPacket)
	}

	p.Version = Version(buf[n])
	n++

	if p.Version < MQTT31 || p.Version > MQTT50 {
		return n, fmt.Errorf("%w: invalid protocol version", ErrMalformedPacket)
	}
	if string(name) != protocolNames[p.Version-MQTT31] {
		return n, fmt.Errorf("%w: invalid protocol name", ErrMalformedPacket)
	}
	return n, nil
}

func (p *Connect) decodeFlags(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("%w: missing connect flags", ErrMalformedPacket)
	}

	p.Flags = ConnectFlags(buf[0])
	n := 1

	if p.Flags.Reserved() {
		return n, fmt.Errorf("%w: invalid connect flags - reserved", ErrMalformedPacket)
	}

	willFlags := p.Flags.WillFlag()
	willQos := p.Flags.WillQoS()

	if !willFlags && willQos != QoS0 {
		return n, fmt.Errorf("%w: invalid connect flags - will qos without will flag", ErrMalformedPacket)
	}
	if willQos > QoS2 {
		return n, fmt.Errorf("%w: invalid connect flags - will qos", ErrMalformedPacket)
	}
	if !willFlags && p.Flags.WillRetain() {
		return n, fmt.Errorf("%w: invalid connect flags - will retain without will flag", ErrMalformedPacket)
	}
	if p.Version != MQTT50 && p.Flags.Password() && !p.Flags.Username() {
		return n, fmt.Errorf("%w: invalid connect flags - password without username", ErrMalformedPacket)
	}
	return n, nil
}

func (p *Connect) decodeKeepAlive(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("%w: missing keep alive", ErrMalformedPacket)
	}

	var err error

	p.KeepAlive, err = decodeUint[uint16](buf)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid keep alive", ErrMalformedPacket)
	}
	return 2, nil
}

func (p *Connect) decodeProperties(buf []byte) (int, error) {
	if p.Version != MQTT50 {
		return 0, nil
	}

	var n int
	var err error

	p.Properties, n, err = decodeProperties[ConnectProperties](buf)
	if err != nil {
		return n, fmt.Errorf("connect properties: %w", err)
	}
	return n, nil
}

func (p *Connect) decodeClientID(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, fmt.Errorf("%w: missing client identifier", ErrMalformedPacket)
	}

	id, n, err := decodeString(buf)
	if err != nil {
		return n, fmt.Errorf("%w: invalid client identifier", ErrMalformedPacket)
	}
	if len(id) == 0 && p.Version == MQTT31 {
		return n, fmt.Errorf("%w: missing client identifier", ErrMalformedPacket)
	}

	p.ClientID = make([]byte, len(id))
	copy(p.ClientID, id)

	return n, nil
}

func (p *Connect) decodeWillProperties(buf []byte) (int, error) {
	if !p.Flags.WillFlag() || p.Version != MQTT50 {
		return 0, nil
	}

	var n int
	var err error

	p.WillProperties, n, err = decodeProperties[WillProperties](buf)
	if err != nil {
		return n, fmt.Errorf("will properties: %w", err)
	}
	return n, nil
}

func (p *Connect) decodeWill(buf []byte) (int, error) {
	if !p.Flags.WillFlag() {
		return 0, nil
	}
	if len(buf) == 0 {
		return 0, fmt.Errorf("%w: missing will topic", ErrMalformedPacket)
	}

	topic, n, err := decodeString(buf)
	if err != nil {
		return n, fmt.Errorf("%w: invalid will topic", ErrMalformedPacket)
	}
	if !isValidTopicName(string(topic)) {
		return n, fmt.Errorf("%w: invalid will topic", ErrMalformedPacket)
	}

	var payload []byte
	var size int

	if len(buf[n:]) == 0 {
		return 0, fmt.Errorf("%w: missing will payload", ErrMalformedPacket)
	}

	payload, size, err = decodeString(buf[n:])
	n += size
	if err != nil {
		return n, fmt.Errorf("%w: invalid will payload", ErrMalformedPacket)
	}

	p.WillTopic = make([]byte, len(topic))
	copy(p.WillTopic, topic)

	p.WillPayload = make([]byte, len(payload))
	copy(p.WillPayload, payload)

	return n, nil
}

func (p *Connect) decodeUsername(buf []byte) (int, error) {
	if !p.Flags.Username() {
		return 0, nil
	}
	if len(buf) == 0 {
		return 0, fmt.Errorf("%w: missing username", ErrMalformedPacket)
	}

	username, n, err := decodeString(buf)
	if err != nil {
		return n, fmt.Errorf("%w: invalid username", ErrMalformedPacket)
	}

	p.Username = make([]byte, len(username))
	copy(p.Username, username)

	return n, nil
}

func (p *Connect) decodePassword(buf []byte) (int, error) {
	if !p.Flags.Password() {
		return 0, nil
	}
	if len(buf) == 0 {
		return 0, fmt.Errorf("%w: missing password", ErrMalformedPacket)
	}

	password, n, err := decodeBinary(buf)
	if err != nil {
		return n, fmt.Errorf("%w: invalid password", ErrMalformedPacket)
	}

	p.Password = make([]byte, len(password))
	copy(p.Password, password)

	return n, nil
}

// ConnectFlags represents the Connect Flags from the MQTT specification. The Connect Flags contains a number of
// parameters specifying the behavior of the MQTT connection. It also indicates the presence or absence of fields
// in the payload.
type ConnectFlags byte

// Username returns whether the username flag is set or not.
func (f ConnectFlags) Username() bool {
	return (f & connectFlagUsernameFlag >> connectFlagShiftUsernameFlag) > 0
}

// Password returns whether the password flag is set or not.
func (f ConnectFlags) Password() bool {
	return (f & connectFlagPasswordFlag >> connectFlagShiftPasswordFlag) > 0
}

// WillRetain returns whether the Will Retain flag is set or not.
func (f ConnectFlags) WillRetain() bool {
	return (f & connectFlagWillRetain >> connectFlagShiftWillRetain) > 0
}

// WillQoS returns the Will QoS value in the flags.
func (f ConnectFlags) WillQoS() QoS {
	return QoS(f & connectFlagWillQoS >> connectFlagShiftWillQoS)
}

// WillFlag returns whether the Will Flag is set or not.
func (f ConnectFlags) WillFlag() bool {
	return (f & connectFlagWillFlag >> connectFlagShiftWillFlag) > 0
}

// CleanSession returns whether the Clean Session flag is set or not.
func (f ConnectFlags) CleanSession() bool {
	return (f & connectFlagCleanSession >> connectFlagShiftCleanSession) > 0
}

// Reserved returns whether the reserved flag is set or not.
func (f ConnectFlags) Reserved() bool {
	return (f & connectFlagReserved >> connectFlagShiftReserved) > 0
}

// ConnectProperties contains the properties of the CONNECT packet.
type ConnectProperties struct {
	// Flags indicates which properties are present.
	Flags PropertyFlags `json:"flags"`

	// UserProperties is a list of user properties.
	UserProperties []UserProperty `json:"user_properties"`

	// AuthenticationMethod contains the name of the authentication method.
	AuthenticationMethod []byte `json:"authentication_method"`

	// AuthenticationData contains the authentication data.
	AuthenticationData []byte `json:"authentication_data"`

	// SessionExpiryInterval represents the time, in seconds, which the server must store the Session State after
	// the network connection is closed.
	SessionExpiryInterval uint32 `json:"session_expiry_interval"`

	// MaximumPacketSize represents the maximum packet size, in bytes, the client is willing to accept.
	MaximumPacketSize uint32 `json:"maximum_packet_size"`

	// ReceiveMaximum represents the maximum number of inflight messages with QoS > 0.
	ReceiveMaximum uint16 `json:"receive_maximum"`

	// TopicAliasMaximum represents the highest number of Topic Alias that the client accepts.
	TopicAliasMaximum uint16 `json:"topic_alias_maximum"`

	// RequestResponseInfo indicates if the server can send Response Information with the CONNACK Packet.
	RequestResponseInfo bool `json:"request_response_info"`

	// RequestProblemInfo indicates whether the Reason String or User Properties can be sent to the client in case
	// of failures on any packet.
	RequestProblemInfo bool `json:"request_problem_info"`
}

// Has returns whether the property is present or not.
func (p *ConnectProperties) Has(id PropertyID) bool {
	if p == nil {
		return false
	}
	return p.Flags.Has(id)
}

// Set sets the property indicating that it's present.
func (p *ConnectProperties) Set(id PropertyID) {
	if p == nil {
		return
	}
	p.Flags = p.Flags.Set(id)
}

func (p *ConnectProperties) size() int {
	var size int

	if p != nil {
		size += sizePropSessionExpiryInterval(p.Flags)
		size += sizePropReceiveMaximum(p.Flags)
		size += sizePropMaxPacketSize(p.Flags)
		size += sizePropTopicAliasMaximum(p.Flags)
		size += sizePropRequestProblemInfo(p.Flags)
		size += sizePropRequestResponseInfo(p.Flags)
		size += sizePropUserProperties(p.Flags, p.UserProperties)
		size += sizePropAuthenticationMethod(p.Flags, p.AuthenticationMethod)
		size += sizePropAuthenticationData(p.Flags, p.AuthenticationData)
	}

	return size
}

func (p *ConnectProperties) decode(buf []byte, remaining int) (n int, err error) {
	for remaining > 0 {
		var b byte
		var size int

		if n >= len(buf) {
			return n, fmt.Errorf("%w: missing connect properties", ErrMalformedPacket)
		}

		b = buf[n]
		n++
		remaining--

		size, err = p.decodeProperty(PropertyID(b), buf[n:])
		n += size
		remaining -= size

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (p *ConnectProperties) decodeProperty(id PropertyID, buf []byte) (n int, err error) {
	switch id {
	case PropertySessionExpiryInterval:
		p.SessionExpiryInterval, n, err = decodePropSessionExpiryInterval(buf, p)
	case PropertyReceiveMaximum:
		p.ReceiveMaximum, n, err = decodePropReceiveMaximum(buf, p)
	case PropertyMaximumPacketSize:
		p.MaximumPacketSize, n, err = decodePropMaxPacketSize(buf, p)
	case PropertyTopicAliasMaximum:
		p.TopicAliasMaximum, n, err = decodePropTopicAliasMaximum(buf, p)
	case PropertyRequestResponseInfo:
		p.RequestResponseInfo, n, err = decodePropRequestResponseInfo(buf, p)
	case PropertyRequestProblemInfo:
		p.RequestProblemInfo, n, err = decodePropRequestProblemInfo(buf, p)
	case PropertyUserProperty:
		var user UserProperty
		user, n, err = decodePropUserProperty(buf, p)
		if err == nil {
			p.UserProperties = append(p.UserProperties, user)
		}
	case PropertyAuthenticationMethod:
		p.AuthenticationMethod, n, err = decodePropAuthenticationMethod(buf, p)
	case PropertyAuthenticationData:
		p.AuthenticationData, n, err = decodePropAuthenticationData(buf, p)
	default:
		err = fmt.Errorf("%w: invalid connect property: %v", ErrMalformedPacket, id)
	}

	return n, err
}

// WillProperties defines the properties to be sent with the Will message when it is published, and properties
// which define when to publish the Will message.
type WillProperties struct {
	// Flags indicates which properties are present.
	Flags PropertyFlags `json:"flags"`

	// UserProperties is a list of user properties.
	UserProperties []UserProperty `json:"user_properties"`

	// CorrelationData is used to correlate a response message with a request message.
	CorrelationData []byte `json:"correlation_data"`

	// ContentType describes the content type of the Will Payload.
	ContentType []byte `json:"content_type"`

	// ResponseTopic indicates the topic name for response message.
	ResponseTopic []byte `json:"response_topic"`

	// WillDelayInterval represents the number of seconds which the server must delay before publish the Will
	// message.
	WillDelayInterval uint32 `json:"will_delay_interval"`

	// MessageExpiryInterval represents the lifetime, in seconds, of the Will message.
	MessageExpiryInterval uint32 `json:"message_expiry_interval"`

	// PayloadFormatIndicator indicates whether the Will message is a UTF-8 string or not.
	PayloadFormatIndicator bool `json:"payload_format_indicator"`
}

// Has returns whether the property is present or not.
func (p *WillProperties) Has(id PropertyID) bool {
	if p == nil {
		return false
	}
	return p.Flags.Has(id)
}

// Set sets the property indicating that it's present.
func (p *WillProperties) Set(id PropertyID) {
	if p == nil {
		return
	}
	p.Flags = p.Flags.Set(id)
}

func (p *WillProperties) size() int {
	var size int

	if p != nil {
		size += sizePropWillDelayInterval(p.Flags)
		size += sizePropPayloadFormatIndicator(p.Flags)
		size += sizePropMessageExpiryInterval(p.Flags)
		size += sizePropContentType(p.Flags, p.ContentType)
		size += sizePropResponseTopic(p.Flags, p.ResponseTopic)
		size += sizePropCorrelationData(p.Flags, p.CorrelationData)
		size += sizePropUserProperties(p.Flags, p.UserProperties)
	}

	return size
}

func (p *WillProperties) decode(buf []byte, remaining int) (n int, err error) {
	for remaining > 0 {
		var b byte
		var size int

		if n >= len(buf) {
			return n, fmt.Errorf("%w: missing will properties", ErrMalformedPacket)
		}

		b = buf[n]
		n++
		remaining--

		size, err = p.decodeProperty(PropertyID(b), buf[n:])
		n += size
		remaining -= size

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (p *WillProperties) decodeProperty(id PropertyID, buf []byte) (n int, err error) {
	switch id {
	case PropertyWillDelayInterval:
		p.WillDelayInterval, n, err = decodePropWillDelayInterval(buf, p)
	case PropertyPayloadFormatIndicator:
		p.PayloadFormatIndicator, n, err = decodePropPayloadFormatIndicator(buf, p)
	case PropertyMessageExpiryInterval:
		p.MessageExpiryInterval, n, err = decodePropMessageExpiryInterval(buf, p)
	case PropertyContentType:
		p.ContentType, n, err = decodePropContentType(buf, p)
	case PropertyResponseTopic:
		p.ResponseTopic, n, err = decodePropResponseTopic(buf, p)
	case PropertyCorrelationData:
		p.CorrelationData, n, err = decodePropCorrelationData(buf, p)
	case PropertyUserProperty:
		var user UserProperty
		user, n, err = decodePropUserProperty(buf, p)
		if err == nil {
			p.UserProperties = append(p.UserProperties, user)
		}
	default:
		err = fmt.Errorf("%w: invalid will property: %v", ErrMalformedPacket, id)
	}

	return n, err
}
