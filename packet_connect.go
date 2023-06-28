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

package melitte

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

var protocolNames = []string{"MQIsdp", "MQTT", "MQTT"}

// PacketConnect represents the CONNECT Packet from MQTT specifications.
type PacketConnect struct {
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
	Properties *PropertiesConnect `json:"properties"`

	// WillProperties contains the Will properties.
	WillProperties *PropertiesWill `json:"will_properties"`

	// KeepAlive is a time interval, measured in seconds, that is permitted to elapse between the point at which the
	// client finishes transmitting one control packet and the point it starts sending the next.
	KeepAlive uint16 `json:"keep_alive"`

	// Version represents the MQTT version.
	Version MQTTVersion `json:"version"`

	// Flags represents the Connect flags.
	Flags ConnectFlags `json:"flags"`
}

// Type returns the packet type.
func (p *PacketConnect) Type() PacketType {
	return PacketTypeConnect
}

// Size returns the CONNECT packet size.
func (p *PacketConnect) Size() int {
	// Compute the size of the protocol name, +1 byte for the protocol version, +1 byte for the Connect flags,
	// and +2 bytes for the Keep Alive.
	size := sizeString(protocolNames[p.Version-MQTT31]) + 1 + 1 + 2

	if p.Version == MQTT50 {
		size += p.Properties.size()
	}

	size += sizeBinary(p.ClientID)

	if p.Flags.WillFlag() {
		if p.Version == MQTT50 {
			size += p.WillProperties.size()
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

	header := FixedHeader{
		PacketType:      PacketTypeConnect,
		RemainingLength: size,
	}

	size += header.size()
	return size
}

// Decode decodes the CONNECT Packet from buf and header. This method returns the number of bytes read
// from buf and the error, if it fails to read the packet correctly.
func (p *PacketConnect) Decode(buf []byte, h FixedHeader) (n int, err error) {
	if h.PacketType != PacketTypeConnect {
		return 0, ErrMalformedPacketType
	}

	if h.Flags != 0 {
		return 0, ErrMalformedFlags
	}

	var cnt int

	cnt, err = p.decodeVersion(buf)
	n += cnt
	if err != nil {
		return n, err
	}

	cnt, err = p.decodeFlags(buf[n:])
	n += cnt
	if err != nil {
		return n, err
	}

	err = getUint[uint16](buf[n:], &p.KeepAlive)
	n += 2
	if err != nil {
		return n, ErrMalformedKeepAlive
	}

	cnt, err = p.decodeProperties(buf[n:])
	n += cnt
	if err != nil {
		return n, err
	}

	cnt, err = p.decodeClientID(buf[n:])
	n += cnt
	if err != nil {
		return n, err
	}

	cnt, err = p.decodeWillProperties(buf[n:])
	n += cnt
	if err != nil {
		return n, err
	}

	cnt, err = p.decodeWill(buf[n:])
	n += cnt
	if err != nil {
		return n, err
	}

	cnt, err = p.decodeUsername(buf[n:])
	n += cnt
	if err != nil {
		return n, err
	}

	cnt, err = p.decodePassword(buf[n:])
	n += cnt
	if err != nil {
		return n, err
	}

	return n, nil
}

func (p *PacketConnect) decodeVersion(buf []byte) (int, error) {
	name, n, err := getString(buf)
	if err != nil {
		return n, ErrMalformedProtocolName
	}
	if n >= len(buf) {
		return n, ErrMalformedProtocolVersion
	}

	p.Version = MQTTVersion(buf[n])
	n++

	if p.Version < MQTT31 || p.Version > MQTT50 {
		return n, ErrMalformedProtocolVersion
	}

	if string(name) != protocolNames[p.Version-MQTT31] {
		return n, ErrMalformedProtocolName
	}

	return n, nil
}

func (p *PacketConnect) decodeFlags(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, ErrMalformedConnectFlags
	}

	p.Flags = ConnectFlags(buf[0])
	n := 1

	if p.Flags.Reserved() {
		return n, ErrMalformedConnectFlags
	}

	willFlags := p.Flags.WillFlag()
	willQos := p.Flags.WillQoS()

	if !willFlags && willQos != QoS0 {
		return n, ErrMalformedConnectFlags
	}

	if willQos > QoS2 {
		return n, ErrMalformedConnectFlags
	}

	if !willFlags && p.Flags.WillRetain() {
		return n, ErrMalformedConnectFlags
	}

	if p.Version != MQTT50 && p.Flags.Password() && !p.Flags.Username() {
		return n, ErrMalformedConnectFlags
	}

	return n, nil
}

func (p *PacketConnect) decodeProperties(buf []byte) (int, error) {
	if p.Version != MQTT50 {
		return 0, nil
	}

	var n int
	var err error

	p.Properties, n, err = decodeProperties[PropertiesConnect](buf)
	return n, err
}

func (p *PacketConnect) decodeClientID(buf []byte) (int, error) {
	id, n, err := getString(buf)
	if err != nil {
		return n, ErrMalformedClientID
	}

	if len(id) == 0 {
		if p.Version != MQTT50 && !p.Flags.CleanSession() {
			return n, ErrV3ClientIDRejected
		}
	}

	p.ClientID = id
	return n, nil
}

func (p *PacketConnect) decodeWillProperties(buf []byte) (int, error) {
	if !p.Flags.WillFlag() || p.Version != MQTT50 {
		return 0, nil
	}

	var n int
	var err error

	p.WillProperties, n, err = decodeProperties[PropertiesWill](buf)
	return n, err
}

func (p *PacketConnect) decodeWill(buf []byte) (int, error) {
	if !p.Flags.WillFlag() {
		return 0, nil
	}

	topic, n, err := getString(buf)
	if err != nil {
		return n, ErrMalformedWillTopic
	}
	if !isValidTopicName(string(topic)) {
		return n, ErrMalformedWillTopic
	}

	var payload []byte
	var size int

	payload, size, err = getString(buf[n:])
	n += size
	if err != nil {
		return n, ErrMalformedWillPayload
	}

	p.WillTopic = topic
	p.WillPayload = payload
	return n, nil
}

func (p *PacketConnect) decodeUsername(buf []byte) (int, error) {
	if !p.Flags.Username() {
		return 0, nil
	}

	username, n, err := getString(buf)
	if err != nil {
		return n, ErrMalformedUsername
	}

	p.Username = username
	return n, nil
}

func (p *PacketConnect) decodePassword(buf []byte) (int, error) {
	if !p.Flags.Password() {
		return 0, nil
	}

	password, n, err := getBinary(buf)
	if err != nil {
		return n, ErrMalformedPassword
	}

	p.Password = password
	return n, nil
}
