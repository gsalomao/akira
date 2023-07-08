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
	"golang.org/x/exp/constraints"
)

// Represents the MQTT property identifiers.
const (
	PropertyIDPayloadFormatIndicator        PropertyID = 0x01
	PropertyIDMessageExpiryInterval         PropertyID = 0x02
	PropertyIDContentType                   PropertyID = 0x03
	PropertyIDResponseTopic                 PropertyID = 0x08
	PropertyIDCorrelationData               PropertyID = 0x09
	PropertyIDSessionExpiryInterval         PropertyID = 0x11
	PropertyIDAssignedClientID              PropertyID = 0x12
	PropertyIDServerKeepAlive               PropertyID = 0x13
	PropertyIDAuthenticationMethod          PropertyID = 0x15
	PropertyIDAuthenticationData            PropertyID = 0x16
	PropertyIDRequestProblemInfo            PropertyID = 0x17
	PropertyIDWillDelayInterval             PropertyID = 0x18
	PropertyIDRequestResponseInfo           PropertyID = 0x19
	PropertyIDResponseInfo                  PropertyID = 0x1a
	PropertyIDServerReference               PropertyID = 0x1c
	PropertyIDReasonString                  PropertyID = 0x1f
	PropertyIDReceiveMaximum                PropertyID = 0x21
	PropertyIDTopicAliasMaximum             PropertyID = 0x22
	PropertyIDMaximumQoS                    PropertyID = 0x24
	PropertyIDRetainAvailable               PropertyID = 0x25
	PropertyIDUserProperty                  PropertyID = 0x26
	PropertyIDMaximumPacketSize             PropertyID = 0x27
	PropertyIDWildcardSubscriptionAvailable PropertyID = 0x28
	PropertyIDSubscriptionIDAvailable       PropertyID = 0x29
	PropertyIDSharedSubscriptionAvailable   PropertyID = 0x2a
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

type properties interface {
	Has(id PropertyID) bool
	Set(id PropertyID)
}

type propertiesDecoder interface {
	properties
	decode(buf []byte, remaining int) (n int, err error)
}

func decodeProperties[T any](buf []byte) (p *T, n int, err error) {
	var remaining int

	n, err = decodeVarInteger(buf, &remaining)
	if err != nil {
		return nil, n, ErrMalformedPropertyLength
	}
	if remaining == 0 {
		return nil, n, nil
	}

	var size int
	p = new(T)

	props, ok := any(p).(propertiesDecoder)
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

func sizePropSessionExpiryInterval(flags PropertyFlags) int {
	if flags.Has(PropertyIDSessionExpiryInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropReceiveMaximum(flags PropertyFlags) int {
	if flags.Has(PropertyIDReceiveMaximum) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropMaxPacketSize(flags PropertyFlags) int {
	if flags.Has(PropertyIDMaximumPacketSize) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropTopicAliasMaximum(flags PropertyFlags) int {
	if flags.Has(PropertyIDTopicAliasMaximum) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropRequestResponseInfo(flags PropertyFlags) int {
	if flags.Has(PropertyIDRequestResponseInfo) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropRequestProblemInfo(flags PropertyFlags) int {
	if flags.Has(PropertyIDRequestProblemInfo) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropUserProperties(flags PropertyFlags, val []UserProperty) int {
	if flags.Has(PropertyIDUserProperty) {
		var size int

		for _, p := range val {
			size++
			size += sizeBinary(p.Key)
			size += sizeBinary(p.Value)
		}

		return size
	}

	return 0
}

func sizePropAuthenticationMethod(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyIDAuthenticationMethod) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropAuthenticationData(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyIDAuthenticationData) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropWillDelayInterval(flags PropertyFlags) int {
	if flags.Has(PropertyIDWillDelayInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropPayloadFormatIndicator(flags PropertyFlags) int {
	if flags.Has(PropertyIDPayloadFormatIndicator) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropMessageExpiryInterval(flags PropertyFlags) int {
	if flags.Has(PropertyIDMessageExpiryInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropContentType(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyIDContentType) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropResponseTopic(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyIDResponseTopic) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropCorrelationData(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyIDCorrelationData) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropMaxQoS(flags PropertyFlags) int {
	if flags.Has(PropertyIDMaximumQoS) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropRetainAvailable(flags PropertyFlags) int {
	if flags.Has(PropertyIDRetainAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropAssignedClientID(flags PropertyFlags, id []byte) int {
	if flags.Has(PropertyIDAssignedClientID) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(id) + 1
	}
	return 0
}

func sizePropReasonString(flags PropertyFlags, reasonString []byte) int {
	if flags.Has(PropertyIDReasonString) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(reasonString) + 1
	}
	return 0
}

func sizePropWildcardSubscriptionAvailable(flags PropertyFlags) int {
	if flags.Has(PropertyIDWildcardSubscriptionAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropSubscriptionIDAvailable(flags PropertyFlags) int {
	if flags.Has(PropertyIDSubscriptionIDAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropSharedSubscriptionAvailable(flags PropertyFlags) int {
	if flags.Has(PropertyIDSharedSubscriptionAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropServerKeepAlive(flags PropertyFlags) int {
	if flags.Has(PropertyIDServerKeepAlive) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropResponseInfo(flags PropertyFlags, info []byte) int {
	if flags.Has(PropertyIDResponseInfo) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(info) + 1
	}
	return 0
}

func sizePropServerReference(flags PropertyFlags, reference []byte) int {
	if flags.Has(PropertyIDServerReference) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(reference) + 1
	}
	return 0
}

func decodePropSessionExpiryInterval(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertyIDSessionExpiryInterval) {
		return v, n, ErrMalformedPropertySessionExpiryInterval
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertySessionExpiryInterval
	}

	p.Set(PropertyIDSessionExpiryInterval)
	return v, n, nil
}

func decodePropReceiveMaximum(buf []byte, p properties) (v uint16, n int, err error) {
	if p.Has(PropertyIDReceiveMaximum) {
		return v, n, ErrMalformedPropertyReceiveMaximum
	}

	n, err = decodePropUint[uint16](buf, &v, func(val uint16) bool { return val != 0 })
	if err != nil {
		return v, n, ErrMalformedPropertyReceiveMaximum
	}

	p.Set(PropertyIDReceiveMaximum)
	return v, n, nil
}

func decodePropMaxPacketSize(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertyIDMaximumPacketSize) {
		return v, n, ErrMalformedPropertyMaxPacketSize
	}

	n, err = decodePropUint[uint32](buf, &v, func(val uint32) bool { return val != 0 })
	if err != nil {
		return v, n, ErrMalformedPropertyMaxPacketSize
	}

	p.Set(PropertyIDMaximumPacketSize)
	return v, n, nil
}

func decodePropTopicAliasMaximum(buf []byte, p properties) (v uint16, n int, err error) {
	if p.Has(PropertyIDTopicAliasMaximum) {
		return v, n, ErrMalformedPropertyTopicAliasMaximum
	}

	n, err = decodePropUint[uint16](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertyTopicAliasMaximum
	}

	p.Set(PropertyIDTopicAliasMaximum)
	return v, n, nil
}

func decodePropRequestResponseInfo(buf []byte, p properties) (v bool, n int, err error) {
	if p.Has(PropertyIDRequestResponseInfo) {
		return v, n, ErrMalformedPropertyRequestResponseInfo
	}

	n, err = decodePropBool(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyRequestResponseInfo
	}

	p.Set(PropertyIDRequestResponseInfo)
	return v, n, nil
}

func decodePropRequestProblemInfo(buf []byte, p properties) (v bool, n int, err error) {
	if p.Has(PropertyIDRequestProblemInfo) {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	n, err = decodePropBool(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	p.Set(PropertyIDRequestProblemInfo)
	return v, n, nil
}

func decodePropUserProperty(buf []byte, p properties) (v UserProperty, n int, err error) {
	var size int

	v.Key, size, err = decodeString(buf)
	if err != nil {
		return v, n, ErrMalformedPropertyUserProperty
	}
	n += size

	v.Value, size, err = decodeString(buf[n:])
	if err != nil {
		return v, n, ErrMalformedPropertyUserProperty
	}
	n += size

	p.Set(PropertyIDUserProperty)
	return v, n, nil
}

func decodePropAuthenticationMethod(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyIDAuthenticationMethod) {
		return v, n, ErrMalformedPropertyAuthenticationMethod
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyAuthenticationMethod
	}

	p.Set(PropertyIDAuthenticationMethod)
	return v, n, nil
}

func decodePropAuthenticationData(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyIDAuthenticationData) {
		return v, n, ErrMalformedPropertyAuthenticationData
	}

	n, err = decodePropBinary(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyAuthenticationData
	}

	p.Set(PropertyIDAuthenticationData)
	return v, n, nil
}

func decodePropWillDelayInterval(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertyIDWillDelayInterval) {
		return v, n, ErrMalformedPropertyWillDelayInterval
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertyWillDelayInterval
	}

	p.Set(PropertyIDWillDelayInterval)
	return v, n, nil
}

func decodePropPayloadFormatIndicator(buf []byte, p properties) (v bool, n int, err error) {
	if p.Has(PropertyIDPayloadFormatIndicator) {
		return v, n, ErrMalformedPropertyPayloadFormatIndicator
	}

	n, err = decodePropBool(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyPayloadFormatIndicator
	}

	p.Set(PropertyIDPayloadFormatIndicator)
	return v, n, nil
}

func decodePropMessageExpiryInterval(buf []byte, p properties) (v uint32, n int, err error) {
	if p.Has(PropertyIDMessageExpiryInterval) {
		return v, n, ErrMalformedPropertyMessageExpiryInterval
	}

	n, err = decodePropUint[uint32](buf, &v, nil)
	if err != nil {
		return v, n, ErrMalformedPropertyMessageExpiryInterval
	}

	p.Set(PropertyIDMessageExpiryInterval)
	return v, n, nil
}

func decodePropContentType(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyIDContentType) {
		return v, n, ErrMalformedPropertyContentType
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyContentType
	}

	p.Set(PropertyIDContentType)
	return v, n, nil
}

func decodePropResponseTopic(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyIDResponseTopic) {
		return v, n, ErrMalformedPropertyResponseTopic
	}

	n, err = decodePropString(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyResponseTopic
	}

	if !isValidTopicName(string(v)) {
		return v, n, ErrMalformedPropertyResponseTopic
	}

	p.Set(PropertyIDResponseTopic)
	return v, n, nil
}

func decodePropCorrelationData(buf []byte, p properties) (v []byte, n int, err error) {
	if p.Has(PropertyIDCorrelationData) {
		return v, n, ErrMalformedPropertyCorrelationData
	}

	n, err = decodePropBinary(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyCorrelationData
	}

	p.Set(PropertyIDCorrelationData)
	return v, n, nil
}

func decodePropUint[T constraints.Unsigned](buf []byte, v *T, validator func(T) bool) (n int, err error) {
	var prop T

	err = decodeUint[T](buf, &prop)
	if err != nil {
		return 0, err
	}

	if validator != nil && !validator(prop) {
		return n, ErrMalformedPropertyInvalid
	}

	*v = prop
	return sizeUint(prop), nil
}

func decodePropBool(buf []byte, v *bool) (n int, err error) {
	err = decodeBool(buf, v)
	if err != nil {
		return 0, err
	}

	return 1, nil
}

func decodePropString(buf []byte, str *[]byte) (n int, err error) {
	var data []byte

	data, n, err = decodeString(buf)
	if err != nil {
		return n, err
	}

	*str = data
	return n, nil
}

func decodePropBinary(buf []byte, bin *[]byte) (n int, err error) {
	var data []byte

	data, n, err = decodeBinary(buf)
	if err != nil {
		return n, err
	}

	*bin = data
	return n, nil
}

func encodePropUserProperties(buf []byte, flags PropertyFlags, props []UserProperty, err error) (int, error) {
	var n int
	if err == nil && flags.Has(PropertyIDUserProperty) {
		for _, p := range props {
			var size int

			buf[n] = byte(PropertyIDUserProperty)
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
	}
	return n, err
}

func encodePropUint[T constraints.Unsigned](buf []byte, f PropertyFlags, id PropertyID, v T) int {
	if f.Has(id) {
		var n int

		buf[0] = byte(id)
		n = encodeUint(buf[1:], v)

		return n + 1
	}
	return 0
}

func encodePropBool(buf []byte, f PropertyFlags, id PropertyID, v bool) int {
	if f.Has(id) {
		var n int

		buf[0] = byte(id)
		n = encodeBool(buf[1:], v)

		return n + 1
	}
	return 0
}

func encodePropString(buf []byte, f PropertyFlags, id PropertyID, str []byte) (int, error) {
	if f.Has(id) {
		buf[0] = byte(id)
		n, err := encodeString(buf[1:], str)

		return n + 1, err
	}
	return 0, nil
}
