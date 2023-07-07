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

import (
	"golang.org/x/exp/constraints"
)

// Represents the MQTT Property.
const (
	PropertyPayloadFormatIndicator        Property = 0x01
	PropertyMessageExpiryInterval         Property = 0x02
	PropertyContentType                   Property = 0x03
	PropertyResponseTopic                 Property = 0x08
	PropertyCorrelationData               Property = 0x09
	PropertySessionExpiryInterval         Property = 0x11
	PropertyAssignedClientID              Property = 0x12
	PropertyServerKeepAlive               Property = 0x13
	PropertyAuthenticationMethod          Property = 0x15
	PropertyAuthenticationData            Property = 0x16
	PropertyRequestProblemInfo            Property = 0x17
	PropertyWillDelayInterval             Property = 0x18
	PropertyRequestResponseInfo           Property = 0x19
	PropertyResponseInfo                  Property = 0x1a
	PropertyServerReference               Property = 0x1c
	PropertyReasonString                  Property = 0x1f
	PropertyReceiveMaximum                Property = 0x21
	PropertyTopicAliasMaximum             Property = 0x22
	PropertyMaximumQoS                    Property = 0x24
	PropertyRetainAvailable               Property = 0x25
	PropertyUserProperty                  Property = 0x26
	PropertyMaximumPacketSize             Property = 0x27
	PropertyWildcardSubscriptionAvailable Property = 0x28
	PropertySubscriptionIDAvailable       Property = 0x29
	PropertySharedSubscriptionAvailable   Property = 0x2a
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
			size++
			size += sizeBinary(p.Key)
			size += sizeBinary(p.Value)
		}

		return size
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

func sizePropMaxQoS(flags propertyFlags) int {
	if flags.has(PropertyMaximumQoS) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropRetainAvailable(flags propertyFlags) int {
	if flags.has(PropertyRetainAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropAssignedClientID(flags propertyFlags, id []byte) int {
	if flags.has(PropertyAssignedClientID) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(id) + 1
	}
	return 0
}

func sizePropReasonString(flags propertyFlags, reasonString []byte) int {
	if flags.has(PropertyReasonString) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(reasonString) + 1
	}
	return 0
}

func sizePropWildcardSubscriptionAvailable(flags propertyFlags) int {
	if flags.has(PropertyWildcardSubscriptionAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropSubscriptionIDAvailable(flags propertyFlags) int {
	if flags.has(PropertySubscriptionIDAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropSharedSubscriptionAvailable(flags propertyFlags) int {
	if flags.has(PropertySharedSubscriptionAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropServerKeepAlive(flags propertyFlags) int {
	if flags.has(PropertyServerKeepAlive) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropResponseInfo(flags propertyFlags, info []byte) int {
	if flags.has(PropertyResponseInfo) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(info) + 1
	}
	return 0
}

func sizePropServerReference(flags propertyFlags, reference []byte) int {
	if flags.has(PropertyServerReference) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(reference) + 1
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

	n, err = decodePropBool(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyRequestResponseInfo
	}

	p.Set(PropertyRequestResponseInfo)
	return v, n, nil
}

func decodePropRequestProblemInfo(buf []byte, p properties) (v bool, n int, err error) {
	if p.Has(PropertyRequestProblemInfo) {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	n, err = decodePropBool(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	p.Set(PropertyRequestProblemInfo)
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

	n, err = decodePropBool(buf, &v)
	if err != nil {
		return v, n, ErrMalformedPropertyPayloadFormatIndicator
	}

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

func encodePropUserProperties(buf []byte, flags propertyFlags, props []UserProperty, err error) (int, error) {
	var n int
	if err == nil && flags.has(PropertyUserProperty) {
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
	}
	return n, err
}

func encodePropUint[T constraints.Unsigned](buf []byte, f propertyFlags, p Property, v T) int {
	if f.has(p) {
		var n int

		buf[0] = byte(p)
		n = encodeUint(buf[1:], v)

		return n + 1
	}
	return 0
}

func encodePropBool(buf []byte, f propertyFlags, p Property, v bool) int {
	if f.has(p) {
		var n int

		buf[0] = byte(p)
		n = encodeBool(buf[1:], v)

		return n + 1
	}
	return 0
}

func encodePropString(buf []byte, f propertyFlags, p Property, str []byte) (int, error) {
	if f.has(p) {
		buf[0] = byte(p)
		n, err := encodeString(buf[1:], str)

		return n + 1, err
	}
	return 0, nil
}
