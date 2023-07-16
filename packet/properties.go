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

// UserProperty contains the key/value pair for a user property.
type UserProperty struct {
	// Key represents the key of the key/value pair to the property.
	Key []byte `json:"key"`

	// Value represents the value of the key/value pair to the property.
	Value []byte `json:"value"`
}

func sizePropSessionExpiryInterval(flags PropertyFlags) int {
	if flags.Has(PropertySessionExpiryInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropReceiveMaximum(flags PropertyFlags) int {
	if flags.Has(PropertyReceiveMaximum) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropMaxPacketSize(flags PropertyFlags) int {
	if flags.Has(PropertyMaximumPacketSize) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropTopicAliasMaximum(flags PropertyFlags) int {
	if flags.Has(PropertyTopicAliasMaximum) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropRequestResponseInfo(flags PropertyFlags) int {
	if flags.Has(PropertyRequestResponseInfo) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropRequestProblemInfo(flags PropertyFlags) int {
	if flags.Has(PropertyRequestProblemInfo) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropUserProperties(flags PropertyFlags, val []UserProperty) int {
	if flags.Has(PropertyUserProperty) {
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
	if flags.Has(PropertyAuthenticationMethod) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropAuthenticationData(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyAuthenticationData) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropWillDelayInterval(flags PropertyFlags) int {
	if flags.Has(PropertyWillDelayInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropPayloadFormatIndicator(flags PropertyFlags) int {
	if flags.Has(PropertyPayloadFormatIndicator) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropMessageExpiryInterval(flags PropertyFlags) int {
	if flags.Has(PropertyMessageExpiryInterval) {
		// Size of the field + 1 byte for the property identifier
		return 5
	}
	return 0
}

func sizePropContentType(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyContentType) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropResponseTopic(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyResponseTopic) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropCorrelationData(flags PropertyFlags, val []byte) int {
	if flags.Has(PropertyCorrelationData) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(val) + 1
	}
	return 0
}

func sizePropMaxQoS(flags PropertyFlags) int {
	if flags.Has(PropertyMaximumQoS) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropRetainAvailable(flags PropertyFlags) int {
	if flags.Has(PropertyRetainAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropAssignedClientID(flags PropertyFlags, id []byte) int {
	if flags.Has(PropertyAssignedClientID) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(id) + 1
	}
	return 0
}

func sizePropReasonString(flags PropertyFlags, reasonString []byte) int {
	if flags.Has(PropertyReasonString) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(reasonString) + 1
	}
	return 0
}

func sizePropWildcardSubscriptionAvailable(flags PropertyFlags) int {
	if flags.Has(PropertyWildcardSubscriptionAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropSubscriptionIDAvailable(flags PropertyFlags) int {
	if flags.Has(PropertySubscriptionIDAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropSharedSubscriptionAvailable(flags PropertyFlags) int {
	if flags.Has(PropertySharedSubscriptionAvailable) {
		// Size of the field + 1 byte for the property identifier
		return 2
	}
	return 0
}

func sizePropServerKeepAlive(flags PropertyFlags) int {
	if flags.Has(PropertyServerKeepAlive) {
		// Size of the field + 1 byte for the property identifier
		return 3
	}
	return 0
}

func sizePropResponseInfo(flags PropertyFlags, info []byte) int {
	if flags.Has(PropertyResponseInfo) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(info) + 1
	}
	return 0
}

func sizePropServerReference(flags PropertyFlags, reference []byte) int {
	if flags.Has(PropertyServerReference) {
		// Size of the field + 1 byte for the property identifier
		return sizeBinary(reference) + 1
	}
	return 0
}

func decodePropSessionExpiryInterval(buf []byte, p Properties) (v uint32, n int, err error) {
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

func decodePropReceiveMaximum(buf []byte, p Properties) (v uint16, n int, err error) {
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

func decodePropMaxPacketSize(buf []byte, p Properties) (v uint32, n int, err error) {
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

func decodePropTopicAliasMaximum(buf []byte, p Properties) (v uint16, n int, err error) {
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

func decodePropRequestResponseInfo(buf []byte, p Properties) (v bool, n int, err error) {
	if p.Has(PropertyRequestResponseInfo) {
		return v, n, ErrMalformedPropertyRequestResponseInfo
	}

	v, err = decodeBool(buf)
	if err != nil {
		return v, n, ErrMalformedPropertyRequestResponseInfo
	}

	n++
	p.Set(PropertyRequestResponseInfo)
	return v, n, nil
}

func decodePropRequestProblemInfo(buf []byte, p Properties) (v bool, n int, err error) {
	if p.Has(PropertyRequestProblemInfo) {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	v, err = decodeBool(buf)
	if err != nil {
		return v, n, ErrMalformedPropertyRequestProblemInfo
	}

	n++
	p.Set(PropertyRequestProblemInfo)
	return v, n, nil
}

func decodePropUserProperty(buf []byte, p Properties) (v UserProperty, n int, err error) {
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

func decodePropAuthenticationMethod(buf []byte, p Properties) (v []byte, n int, err error) {
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

func decodePropAuthenticationData(buf []byte, p Properties) (v []byte, n int, err error) {
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

func decodePropWillDelayInterval(buf []byte, p Properties) (v uint32, n int, err error) {
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

func decodePropPayloadFormatIndicator(buf []byte, p Properties) (v bool, n int, err error) {
	if p.Has(PropertyPayloadFormatIndicator) {
		return v, n, ErrMalformedPropertyPayloadFormatIndicator
	}

	v, err = decodeBool(buf)
	if err != nil {
		return v, n, ErrMalformedPropertyPayloadFormatIndicator
	}

	n++
	p.Set(PropertyPayloadFormatIndicator)
	return v, n, nil
}

func decodePropMessageExpiryInterval(buf []byte, p Properties) (v uint32, n int, err error) {
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

func decodePropContentType(buf []byte, p Properties) (v []byte, n int, err error) {
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

func decodePropResponseTopic(buf []byte, p Properties) (v []byte, n int, err error) {
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

func decodePropCorrelationData(buf []byte, p Properties) (v []byte, n int, err error) {
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

func decodePropUint[T integer](buf []byte, v *T, validator func(T) bool) (n int, err error) {
	var prop T

	prop, err = decodeUint[T](buf)
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
	if err == nil && flags.Has(PropertyUserProperty) {
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

func encodePropUint[T integer](buf []byte, f PropertyFlags, id PropertyID, v T) int {
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
