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

import (
	"golang.org/x/exp/constraints"
)

// Represents the MQTT Property.
const (
	PropertyPayloadFormatIndicator Property = 0x01
	PropertyMessageExpiryInterval  Property = 0x02
	PropertyContentType            Property = 0x03
	PropertyResponseTopic          Property = 0x08
	PropertyCorrelationData        Property = 0x09
	PropertySessionExpiryInterval  Property = 0x11
	PropertyAuthenticationMethod   Property = 0x15
	PropertyAuthenticationData     Property = 0x16
	PropertyRequestProblemInfo     Property = 0x17
	PropertyWillDelayInterval      Property = 0x18
	PropertyRequestResponseInfo    Property = 0x19
	PropertyReceiveMaximum         Property = 0x21
	PropertyTopicAliasMaximum      Property = 0x22
	PropertyUserProperty           Property = 0x26
	PropertyMaximumPacketSize      Property = 0x27
)

// Property represents the MQTT property.
type Property byte

type propertyFlags uint64

func (f propertyFlags) has(p Property) bool {
	return f&(1<<p) > 0
}

func (f propertyFlags) set(p Property) propertyFlags {
	return f | 1<<p
}

type properties interface {
	Has(prop Property) bool
	Set(prop Property)
	decode(buf []byte, remaining int) (n int, err error)
}

func decodeProperties[T any](buf []byte) (p *T, n int, err error) {
	var remaining int

	n, err = getVarInteger(buf, &remaining)
	if err != nil {
		return nil, n, ErrMalformedPropertyLength
	}
	if remaining == 0 {
		return nil, n, nil
	}

	var size int
	p = new(T)

	props, ok := any(p).(properties)
	if !ok {
		return nil, n, ErrMalformedPropertyInvalid
	}

	size, err = props.decode(buf[n:], remaining)
	n += size

	return p, n, err
}

// UserProperty contains the key/value pair to a user property.
type UserProperty struct {
	// Key represents the key of the key/value pair to the property.
	Key []byte `json:"key"`

	// Value represents the value of the key/value pair to the property.
	Value []byte `json:"value"`
}

// PropertiesConnect contains the properties of the CONNECT packet.
type PropertiesConnect struct {
	// Flags indicates which properties are present.
	Flags propertyFlags `json:"flags"`

	// UserProperties is a list of user properties.
	UserProperties []UserProperty `json:"user_properties"`

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

	// RequestResponseInfo indicates if the server can send Response Information with the CONNACK Packet.
	RequestResponseInfo bool `json:"request_response_info"`

	// RequestProblemInfo indicates whether the Reason String or User Properties can be sent to the client in case of
	// failures on any packet.
	RequestProblemInfo bool `json:"request_problem_info"`
}

// Has returns whether the property is present or not.
func (p *PropertiesConnect) Has(prop Property) bool {
	if p == nil {
		return false
	}
	return p.Flags.has(prop)
}

// Set sets the property indicating that it's present.
func (p *PropertiesConnect) Set(prop Property) {
	if p == nil {
		return
	}
	p.Flags = p.Flags.set(prop)
}

func (p *PropertiesConnect) size() int {
	if p == nil {
		// Even if it is nil, 1 byte is required for the property length.
		return 1
	}

	var size int

	size += sizePropSessionExpiryInterval(p.Flags)
	size += sizePropReceiveMaximum(p.Flags)
	size += sizePropMaxPacketSize(p.Flags)
	size += sizePropTopicAliasMaximum(p.Flags)
	size += sizePropRequestProblemInfo(p.Flags)
	size += sizePropRequestResponseInfo(p.Flags)
	size += sizePropUserProperties(p.Flags, p.UserProperties)
	size += sizePropAuthenticationMethod(p.Flags, p.AuthenticationMethod)
	size += sizePropAuthenticationData(p.Flags, p.AuthenticationData)

	size += sizeVarInteger(size)
	return size
}

func (p *PropertiesConnect) decode(buf []byte, remaining int) (n int, err error) {
	for remaining > 0 {
		var b byte
		var size int

		if n >= len(buf) {
			return n, ErrMalformedPropertyConnect
		}

		b = buf[n]
		n++
		remaining--

		size, err = p.decodeProperty(Property(b), buf[n:])
		n += size
		remaining -= size

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (p *PropertiesConnect) decodeProperty(prop Property, buf []byte) (n int, err error) {
	switch prop {
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
		err = ErrMalformedPropertyInvalid
	}

	return n, err
}

// PropertiesWill defines the properties to be sent with the Will message when it is published, and properties
// which define when to publish the Will message.
type PropertiesWill struct {
	// Flags indicates which properties are present.
	Flags propertyFlags `json:"flags"`

	// UserProperties is a list of user properties.
	UserProperties []UserProperty `json:"user_properties"`

	// CorrelationData is used to correlate a response message with a request message.
	CorrelationData []byte `json:"correlation_data"`

	// ContentType describes the content type of the Will Payload.
	ContentType []byte `json:"content_type"`

	// ResponseTopic indicates the topic name for response message.
	ResponseTopic []byte `json:"response_topic"`

	// WillDelayInterval represents the number of seconds which the server must delay before publish the Will message.
	WillDelayInterval uint32 `json:"will_delay_interval"`

	// MessageExpiryInterval represents the lifetime, in seconds, of the Will message.
	MessageExpiryInterval uint32 `json:"message_expiry_interval"`

	// PayloadFormatIndicator indicates whether the Will message is a UTF-8 string or not.
	PayloadFormatIndicator bool `json:"payload_format_indicator"`
}

// Has returns whether the property is present or not.
func (p *PropertiesWill) Has(prop Property) bool {
	if p == nil {
		return false
	}
	return p.Flags.has(prop)
}

// Set sets the property indicating that it's present.
func (p *PropertiesWill) Set(prop Property) {
	if p == nil {
		return
	}
	p.Flags = p.Flags.set(prop)
}

func (p *PropertiesWill) size() int {
	if p == nil {
		// Even if it is nil, 1 byte is required for the property length.
		return 1
	}

	var size int

	size += sizePropWillDelayInterval(p.Flags)
	size += sizePropPayloadFormatIndicator(p.Flags)
	size += sizePropMessageExpiryInterval(p.Flags)
	size += sizePropContentType(p.Flags, p.ContentType)
	size += sizePropResponseTopic(p.Flags, p.ResponseTopic)
	size += sizePropCorrelationData(p.Flags, p.CorrelationData)
	size += sizePropUserProperties(p.Flags, p.UserProperties)

	size += sizeVarInteger(size)
	return size
}

func (p *PropertiesWill) decode(buf []byte, remaining int) (n int, err error) {
	for remaining > 0 {
		var b byte
		var size int

		if n >= len(buf) {
			return n, ErrMalformedPropertyWill
		}

		b = buf[n]
		n++
		remaining--

		size, err = p.decodeProperty(Property(b), buf[n:])
		n += size
		remaining -= size

		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (p *PropertiesWill) decodeProperty(prop Property, buf []byte) (n int, err error) {
	switch prop {
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
		err = ErrMalformedPropertyInvalid
	}

	return n, err
}

func sizePropSessionExpiryInterval(flags propertyFlags) int {
	if flags.has(PropertySessionExpiryInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropReceiveMaximum(flags propertyFlags) int {
	if flags.has(PropertyReceiveMaximum) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropMaxPacketSize(flags propertyFlags) int {
	if flags.has(PropertyMaximumPacketSize) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropTopicAliasMaximum(flags propertyFlags) int {
	if flags.has(PropertyTopicAliasMaximum) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropRequestResponseInfo(flags propertyFlags) int {
	if flags.has(PropertyRequestResponseInfo) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropRequestProblemInfo(flags propertyFlags) int {
	if flags.has(PropertyRequestProblemInfo) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropUserProperties(flags propertyFlags, val []UserProperty) int {
	if flags.has(PropertyUserProperty) {
		var size int

		for _, p := range val {
			size += sizeBinary(p.Key)
			size += sizeBinary(p.Value)
		}

		// Size of the field + 1 byte for the property identifier
		return size + 1
	}

	return 0
}

func sizePropAuthenticationMethod(flags propertyFlags, val []byte) int {
	if flags.has(PropertyAuthenticationMethod) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropAuthenticationData(flags propertyFlags, val []byte) int {
	if flags.has(PropertyAuthenticationData) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropWillDelayInterval(flags propertyFlags) int {
	if flags.has(PropertyWillDelayInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropPayloadFormatIndicator(flags propertyFlags) int {
	if flags.has(PropertyPayloadFormatIndicator) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropMessageExpiryInterval(flags propertyFlags) int {
	if flags.has(PropertyMessageExpiryInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropContentType(flags propertyFlags, val []byte) int {
	if flags.has(PropertyContentType) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropResponseTopic(flags propertyFlags, val []byte) int {
	if flags.has(PropertyResponseTopic) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropCorrelationData(flags propertyFlags, val []byte) int {
	if flags.has(PropertyCorrelationData) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func decodePropSessionExpiryInterval(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertySessionExpiryInterval) {
		return v, n, ErrMalformedPropertySessionExpiryInterval
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertySessionExpiryInterval
	}

	p.Set(PropertySessionExpiryInterval)
	return v, n, nil
}

func decodePropReceiveMaximum(buf []byte, p properties) (v uint16, n int, err error) {
	if p.Has(PropertyReceiveMaximum) {
		return v, n, ErrMalformedPropertyReceiveMaximum
	}

	n, err = decodePropUint[uint16](buf, &v, func(val uint16) bool { return val != 0 })
	if err != nil {
		return v, n, ErrMalformedPropertyReceiveMaximum
	}

	p.Set(PropertyReceiveMaximum)
	return v, n, nil
}

func decodePropMaxPacketSize(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertyMaximumPacketSize) {
		return v, n, ErrMalformedPropertyMaxPacketSize
	}

	n, err = decodePropUint[uint32](buf, &v, func(val uint32) bool { return val != 0 })
	if err != nil {
		return v, n, ErrMalformedPropertyMaxPacketSize
	}

	p.Set(PropertyMaximumPacketSize)
	return v, n, nil
}

func decodePropTopicAliasMaximum(buf []byte, p properties) (v uint16, n int, err error) {
	if p.Has(PropertyTopicAliasMaximum) {
		return v, n, ErrMalformedPropertyTopicAliasMaximum
	}

	n, err = decodePropUint[uint16](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertyTopicAliasMaximum
	}

	p.Set(PropertyTopicAliasMaximum)
	return v, n, nil
}

func decodePropRequestResponseInfo(buf []byte, p properties) (v bool, n int, err error) {
	if p.Has(PropertyRequestResponseInfo) {
		return v, n, ErrMalformedPropertyRequestResponseInfo
	}

	var b byte
	n, err = decodePropUint[uint8](buf, &b, func(val byte) bool { return val == 0 || val == 1 })
	if err != nil {
		return v, n, ErrMalformedPropertyRequestResponseInfo
	}

	v = b == 1
	p.Set(PropertyRequestResponseInfo)
	return v, n, nil
}

func decodePropRequestProblemInfo(buf []byte, p properties) (v bool, n int, err error) {
	if p.Has(PropertyRequestProblemInfo) {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	var b byte
	n, err = decodePropUint[uint8](buf, &b, func(val byte) bool { return val == 0 || val == 1 })
	if err != nil {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	v = b == 1
	p.Set(PropertyRequestProblemInfo)
	return v, n, nil
}

func decodePropUserProperty(buf []byte, p properties) (v UserProperty, n int, err error) {
	var size int

	v.Key, size, err = getString(buf)
	if err != nil {
		return v, n, ErrMalformedPropertyUserProperty
	}
	n += size

	v.Value, size, err = getString(buf[n:])
	if err != nil {
		return v, n, ErrMalformedPropertyUserProperty
	}
	n += size

	p.Set(PropertyUserProperty)
	return v, n, nil
}

func decodePropAuthenticationMethod(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyAuthenticationMethod) {
		return v, n, ErrMalformedPropertyAuthenticationMethod
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyAuthenticationMethod
	}

	p.Set(PropertyAuthenticationMethod)
	return v, n, nil
}

func decodePropAuthenticationData(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyAuthenticationData) {
		return v, n, ErrMalformedPropertyAuthenticationData
	}

	n, err = decodePropBinary(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyAuthenticationData
	}

	p.Set(PropertyAuthenticationData)
	return v, n, nil
}

func decodePropWillDelayInterval(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertyWillDelayInterval) {
		return v, n, ErrMalformedPropertyWillDelayInterval
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertyWillDelayInterval
	}

	p.Set(PropertyWillDelayInterval)
	return v, n, nil
}

func decodePropPayloadFormatIndicator(buf []byte, p properties) (v bool, n int, err error) {
	if p.Has(PropertyPayloadFormatIndicator) {
		return v, n, ErrMalformedPropertyPayloadFormatIndicator
	}

	var b byte
	n, err = decodePropUint[byte](buf, &b, func(val byte) bool { return val == 0 || val == 1 })
	if err != nil {
		return v, n, ErrMalformedPropertyPayloadFormatIndicator
	}

	v = b == 1
	p.Set(PropertyPayloadFormatIndicator)
	return v, n, nil
}

func decodePropMessageExpiryInterval(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertyMessageExpiryInterval) {
		return v, n, ErrMalformedPropertyMessageExpiryInterval
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertyMessageExpiryInterval
	}

	p.Set(PropertyMessageExpiryInterval)
	return v, n, nil
}

func decodePropContentType(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyContentType) {
		return v, n, ErrMalformedPropertyContentType
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyContentType
	}

	p.Set(PropertyContentType)
	return v, n, nil
}

func decodePropResponseTopic(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyResponseTopic) {
		return v, n, ErrMalformedPropertyResponseTopic
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyResponseTopic
	}

	if !isValidTopicName(string(v)) {
		return v, n, ErrMalformedPropertyResponseTopic
	}

	p.Set(PropertyResponseTopic)
	return v, n, nil
}

func decodePropCorrelationData(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyCorrelationData) {
		return v, n, ErrMalformedPropertyCorrelationData
	}

	n, err = decodePropBinary(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyCorrelationData
	}

	p.Set(PropertyCorrelationData)
	return v, n, nil
}

func decodePropUint[T constraints.Unsigned](buf []byte, v *T, validator func(T) bool) (n int, err error) {
	var prop T

	err = getUint[T](buf, &prop)
	if err != nil {
		return 0, err
	}

	if validator != nil && !validator(prop) {
		return n, ErrMalformedPropertyInvalid
	}

	*v = prop
	return sizeUint(prop), nil
}

func decodePropString(buf []byte, str *[]byte) (n int, err error) {
	var data []byte

	data, n, err = getString(buf)
	if err != nil {
		return n, err
	}

	*str = data
	return n, nil
}

func decodePropBinary(buf []byte, bin *[]byte) (n int, err error) {
	var data []byte

	data, n, err = getBinary(buf)
	if err != nil {
		return n, err
	}

	*bin = data
	return n, nil
}
