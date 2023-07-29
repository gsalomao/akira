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

// Represents the MQTT property identifiers.
const (
	PropertyPayloadFormatIndicator        PropertyID = 0x01
	PropertyMessageExpiryInterval         PropertyID = 0x02
	PropertyContentType                   PropertyID = 0x03
	PropertyResponseTopic                 PropertyID = 0x08
	PropertyCorrelationData               PropertyID = 0x09
	PropertySessionExpiryInterval         PropertyID = 0x11
	PropertyAssignedClientID              PropertyID = 0x12
	PropertyServerKeepAlive               PropertyID = 0x13
	PropertyAuthenticationMethod          PropertyID = 0x15
	PropertyAuthenticationData            PropertyID = 0x16
	PropertyRequestProblemInfo            PropertyID = 0x17
	PropertyWillDelayInterval             PropertyID = 0x18
	PropertyRequestResponseInfo           PropertyID = 0x19
	PropertyResponseInfo                  PropertyID = 0x1a
	PropertyServerReference               PropertyID = 0x1c
	PropertyReasonString                  PropertyID = 0x1f
	PropertyReceiveMaximum                PropertyID = 0x21
	PropertyTopicAliasMaximum             PropertyID = 0x22
	PropertyMaximumQoS                    PropertyID = 0x24
	PropertyRetainAvailable               PropertyID = 0x25
	PropertyUserProperty                  PropertyID = 0x26
	PropertyMaximumPacketSize             PropertyID = 0x27
	PropertyWildcardSubscriptionAvailable PropertyID = 0x28
	PropertySubscriptionIDAvailable       PropertyID = 0x29
	PropertySharedSubscriptionAvailable   PropertyID = 0x2a
)

// PropertyID represents the MQTT property.
type PropertyID byte

// PropertyFlags holds the flags for each property.
type PropertyFlags uint64

// Has returns true if the flag assigned to id is set. Otherwise, it returns false.
func (f PropertyFlags) Has(id PropertyID) bool {
	return f&(1<<id) > 0
}

// Set sets the flag assigned to id.
func (f PropertyFlags) Set(id PropertyID) PropertyFlags {
	return f | 1<<id
}

// Properties is the interface which all MQTT properties must implement.
type Properties interface {
	// Has returns whether the property is present or not.
	Has(id PropertyID) bool

	// Set sets the property indicating that it's present.
	Set(id PropertyID)
}

type propertiesDecoder interface {
	Properties
	decode(buf []byte, remaining int) (n int, err error)
}

func decodeProperties[T any](buf []byte) (p *T, n int, err error) {
	var remaining int

	n, err = decodeVarInteger(buf, &remaining)
	if err != nil {
		return nil, n, fmt.Errorf("%w: invalid property length: %s", ErrMalformedPacket, err.Error())
	}
	if remaining == 0 {
		return nil, n, nil
	}

	var size int
	p = new(T)

	// It must be a propertiesDecoder. If it's not, this is an integrity issue and let it panic.
	props := any(p).(propertiesDecoder)

	size, err = props.decode(buf[n:], remaining)
	n += size
	return p, n, err
}

// UserProperty contains the key/value pair for a user property.
type UserProperty struct {
	// Key represents the key of the key/value pair to the property.
	Key []byte `json:"key"`

	// Value represents the value of the key/value pair to the property.
	Value []byte `json:"value"`
}

func sizePropSessionExpiryInterval(flags PropertyFlags) int {
	if !flags.Has(PropertySessionExpiryInterval) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 5
}

func sizePropReceiveMaximum(flags PropertyFlags) int {
	if !flags.Has(PropertyReceiveMaximum) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 3
}

func sizePropMaxPacketSize(flags PropertyFlags) int {
	if !flags.Has(PropertyMaximumPacketSize) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 5
}

func sizePropTopicAliasMaximum(flags PropertyFlags) int {
	if !flags.Has(PropertyTopicAliasMaximum) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 3
}

func sizePropRequestResponseInfo(flags PropertyFlags) int {
	if !flags.Has(PropertyRequestResponseInfo) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropRequestProblemInfo(flags PropertyFlags) int {
	if !flags.Has(PropertyRequestProblemInfo) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropUserProperties(flags PropertyFlags, val []UserProperty) int {
	if !flags.Has(PropertyUserProperty) {
		return 0
	}

	var size int
	for _, p := range val {
		size++
		size += sizeBinary(p.Key)
		size += sizeBinary(p.Value)
	}
	return size
}

func sizePropAuthenticationMethod(flags PropertyFlags, val []byte) int {
	if !flags.Has(PropertyAuthenticationMethod) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(val) + 1
}

func sizePropAuthenticationData(flags PropertyFlags, val []byte) int {
	if !flags.Has(PropertyAuthenticationData) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(val) + 1
}

func sizePropWillDelayInterval(flags PropertyFlags) int {
	if !flags.Has(PropertyWillDelayInterval) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 5
}

func sizePropPayloadFormatIndicator(flags PropertyFlags) int {
	if !flags.Has(PropertyPayloadFormatIndicator) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropMessageExpiryInterval(flags PropertyFlags) int {
	if !flags.Has(PropertyMessageExpiryInterval) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 5
}

func sizePropContentType(flags PropertyFlags, val []byte) int {
	if !flags.Has(PropertyContentType) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(val) + 1
}

func sizePropResponseTopic(flags PropertyFlags, val []byte) int {
	if !flags.Has(PropertyResponseTopic) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(val) + 1
}

func sizePropCorrelationData(flags PropertyFlags, val []byte) int {
	if !flags.Has(PropertyCorrelationData) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(val) + 1
}

func sizePropMaxQoS(flags PropertyFlags) int {
	if !flags.Has(PropertyMaximumQoS) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropRetainAvailable(flags PropertyFlags) int {
	if !flags.Has(PropertyRetainAvailable) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropAssignedClientID(flags PropertyFlags, id []byte) int {
	if !flags.Has(PropertyAssignedClientID) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(id) + 1
}

func sizePropReasonString(flags PropertyFlags, reasonString []byte) int {
	if !flags.Has(PropertyReasonString) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(reasonString) + 1
}

func sizePropWildcardSubscriptionAvailable(flags PropertyFlags) int {
	if !flags.Has(PropertyWildcardSubscriptionAvailable) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropSubscriptionIDAvailable(flags PropertyFlags) int {
	if !flags.Has(PropertySubscriptionIDAvailable) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropSharedSubscriptionAvailable(flags PropertyFlags) int {
	if !flags.Has(PropertySharedSubscriptionAvailable) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 2
}

func sizePropServerKeepAlive(flags PropertyFlags) int {
	if !flags.Has(PropertyServerKeepAlive) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return 3
}

func sizePropResponseInfo(flags PropertyFlags, info []byte) int {
	if !flags.Has(PropertyResponseInfo) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(info) + 1
}

func sizePropServerReference(flags PropertyFlags, reference []byte) int {
	if !flags.Has(PropertyServerReference) {
		return 0
	}
	// Size of the field + 1 byte for the property identifier.
	return sizeBinary(reference) + 1
}

func decodePropSessionExpiryInterval(buf []byte, p Properties) (v uint32, n int, err error) {
	if p.Has(PropertySessionExpiryInterval) {
		return v, n, fmt.Errorf("%w: duplicated session expiry interval", ErrMalformedPacket)
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid session expiry interval", ErrMalformedPacket)
	}

	p.Set(PropertySessionExpiryInterval)
	return v, n, nil
}

func decodePropReceiveMaximum(buf []byte, p Properties) (v uint16, n int, err error) {
	if p.Has(PropertyReceiveMaximum) {
		return v, n, fmt.Errorf("%w: duplicated recive maximum", ErrMalformedPacket)
	}

	n, err = decodePropUint[uint16](buf, &v, func(val uint16) bool { return val != 0 })
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid receive maximum", ErrMalformedPacket)
	}

	p.Set(PropertyReceiveMaximum)
	return v, n, nil
}

func decodePropMaxPacketSize(buf []byte, p Properties) (v uint32, n int, err error) {
	if p.Has(PropertyMaximumPacketSize) {
		return v, n, fmt.Errorf("%w: duplicated maximum packet size", ErrMalformedPacket)
	}

	n, err = decodePropUint[uint32](buf, &v, func(val uint32) bool { return val != 0 })
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid maximum packet size", ErrMalformedPacket)
	}

	p.Set(PropertyMaximumPacketSize)
	return v, n, nil
}

func decodePropTopicAliasMaximum(buf []byte, p Properties) (v uint16, n int, err error) {
	if p.Has(PropertyTopicAliasMaximum) {
		return v, n, fmt.Errorf("%w: duplicated topic alias maximum", ErrMalformedPacket)
	}

	n, err = decodePropUint[uint16](buf, &v, nil)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid topic alias maximum", ErrMalformedPacket)
	}

	p.Set(PropertyTopicAliasMaximum)
	return v, n, nil
}

func decodePropRequestResponseInfo(buf []byte, p Properties) (v bool, n int, err error) {
	if p.Has(PropertyRequestResponseInfo) {
		return v, n, fmt.Errorf("%w: duplicated request response info", ErrMalformedPacket)
	}

	v, err = decodeBool(buf)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid request response info", ErrMalformedPacket)
	}

	n++
	p.Set(PropertyRequestResponseInfo)
	return v, n, nil
}

func decodePropRequestProblemInfo(buf []byte, p Properties) (v bool, n int, err error) {
	if p.Has(PropertyRequestProblemInfo) {
		return v, n, fmt.Errorf("%w: duplicated request problem info", ErrMalformedPacket)
	}

	v, err = decodeBool(buf)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid request problem info", ErrMalformedPacket)
	}

	n++
	p.Set(PropertyRequestProblemInfo)
	return v, n, nil
}

func decodePropUserProperty(buf []byte, p Properties) (v UserProperty, n int, err error) {
	var (
		size  int
		key   []byte
		value []byte
	)

	key, size, err = decodeString(buf)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid user property key", ErrMalformedPacket)
	}

	v.Key = make([]byte, len(key))
	copy(v.Key, key)
	n += size

	value, size, err = decodeString(buf[n:])
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid user property value", ErrMalformedPacket)
	}

	v.Value = make([]byte, len(value))
	copy(v.Value, value)
	n += size

	p.Set(PropertyUserProperty)
	return v, n, nil
}

func decodePropAuthenticationMethod(buf []byte, p Properties) (v []byte, n int, err error) {
	if p.Has(PropertyAuthenticationMethod) {
		return v, n, fmt.Errorf("%w: duplicated authentication method", ErrMalformedPacket)
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid authentication method", ErrMalformedPacket)
	}

	p.Set(PropertyAuthenticationMethod)
	return v, n, nil
}

func decodePropAuthenticationData(buf []byte, p Properties) (v []byte, n int, err error) {
	if p.Has(PropertyAuthenticationData) {
		return v, n, fmt.Errorf("%w: duplicated authentication data ", ErrMalformedPacket)
	}

	n, err = decodePropBinary(buf, &v)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid authentication data", ErrMalformedPacket)
	}

	p.Set(PropertyAuthenticationData)
	return v, n, nil
}

func decodePropWillDelayInterval(buf []byte, p Properties) (v uint32, n int, err error) {
	if p.Has(PropertyWillDelayInterval) {
		return v, n, fmt.Errorf("%w: duplicated will delay interval", ErrMalformedPacket)
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid will delay interval", ErrMalformedPacket)
	}

	p.Set(PropertyWillDelayInterval)
	return v, n, nil
}

func decodePropPayloadFormatIndicator(buf []byte, p Properties) (v bool, n int, err error) {
	if p.Has(PropertyPayloadFormatIndicator) {
		return v, n, fmt.Errorf("%w: duplicated payload format indicator", ErrMalformedPacket)
	}

	v, err = decodeBool(buf)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid payload format indicator", ErrMalformedPacket)
	}

	n++
	p.Set(PropertyPayloadFormatIndicator)
	return v, n, nil
}

func decodePropMessageExpiryInterval(buf []byte, p Properties) (v uint32, n int, err error) {
	if p.Has(PropertyMessageExpiryInterval) {
		return v, n, fmt.Errorf("%w: duplicated message expiry interval", ErrMalformedPacket)
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid message expiry interval", ErrMalformedPacket)
	}

	p.Set(PropertyMessageExpiryInterval)
	return v, n, nil
}

func decodePropContentType(buf []byte, p Properties) (v []byte, n int, err error) {
	if p.Has(PropertyContentType) {
		return v, n, fmt.Errorf("%w: duplicated content type", ErrMalformedPacket)
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid content type", ErrMalformedPacket)
	}

	p.Set(PropertyContentType)
	return v, n, nil
}

func decodePropResponseTopic(buf []byte, p Properties) (v []byte, n int, err error) {
	if p.Has(PropertyResponseTopic) {
		return v, n, fmt.Errorf("%w: duplicated response topic", ErrMalformedPacket)
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid response topic", ErrMalformedPacket)
	}
	if !isValidTopicName(string(v)) {
		return v, n, fmt.Errorf("%w: invalid response topic", ErrMalformedPacket)
	}

	p.Set(PropertyResponseTopic)
	return v, n, nil
}

func decodePropCorrelationData(buf []byte, p Properties) (v []byte, n int, err error) {
	if p.Has(PropertyCorrelationData) {
		return v, n, fmt.Errorf("%w: duplicated correlation data", ErrMalformedPacket)
	}

	n, err = decodePropBinary(buf, &v)
	if err != nil {
		return v, n, fmt.Errorf("%w: invalid correlation data", ErrMalformedPacket)
	}

	p.Set(PropertyCorrelationData)
	return v, n, nil
}

func decodePropUint[T integer](buf []byte, v *T, validator func(T) bool) (n int, err error) {
	var prop T

	prop, err = decodeUint[T](buf)
	if err != nil {
		return 0, err
	}

	if validator != nil && !validator(prop) {
		return n, errors.New("invalid integer")
	}

	*v = prop
	return sizeUint(prop), nil
}

func decodePropString(buf []byte, str *[]byte) (n int, err error) {
	var data []byte

	data, n, err = decodeString(buf)
	if err != nil {
		return n, err
	}

	*str = make([]byte, len(data))
	copy(*str, data)

	return n, nil
}

func decodePropBinary(buf []byte, bin *[]byte) (n int, err error) {
	var data []byte

	data, n, err = decodeBinary(buf)
	if err != nil {
		return n, err
	}

	*bin = make([]byte, len(data))
	copy(*bin, data)

	return n, nil
}

func encodePropUserProperties(buf []byte, flags PropertyFlags, props []UserProperty, err error) (int, error) {
	if err != nil || !flags.Has(PropertyUserProperty) {
		return 0, err
	}

	var n int
	for _, p := range props {
		var size int

		buf[n] = byte(PropertyUserProperty)
		n++

		size, err = encodeString(buf[n:], p.Key)
		n += size
		if err != nil {
			return n, err
		}

		size, err = encodeString(buf[n:], p.Value)
		n += size
		if err != nil {
			return n, err
		}
	}
	return n, err
}

func encodePropUint[T integer](buf []byte, f PropertyFlags, id PropertyID, v T) int {
	if !f.Has(id) {
		return 0
	}

	buf[0] = byte(id)
	return encodeUint(buf[1:], v) + 1
}

func encodePropBool(buf []byte, f PropertyFlags, id PropertyID, v bool) int {
	if !f.Has(id) {
		return 0
	}

	buf[0] = byte(id)
	return encodeBool(buf[1:], v) + 1
}

func encodePropString(buf []byte, f PropertyFlags, id PropertyID, str []byte) (int, error) {
	if !f.Has(id) {
		return 0, nil
	}

	buf[0] = byte(id)
	n, err := encodeString(buf[1:], str)
	return n + 1, err
}

func validatePropUserProperty(f PropertyFlags, props []UserProperty) error {
	if !f.Has(PropertyUserProperty) {
		return nil
	}
	if len(props) == 0 {
		return errors.New("missing user property")
	}
	for _, p := range props {
		if len(p.Key) == 0 {
			return errors.New("missing user property key")
		}
	}
	return nil
}

func validatePropString(f PropertyFlags, id PropertyID, str []byte) error {
	if f.Has(id) && len(str) == 0 {
		return errors.New("missing string")
	}
	return nil
}
