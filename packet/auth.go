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

// Auth represents the AUTH packet from MQTT specifications.
type Auth struct {
	// Properties contains the properties of the AUTH packet.
	Properties *AuthProperties `json:"properties"`

	// Code represents the reason code based on the MQTT specifications.
	Code ReasonCode `json:"code"`
}

// Type returns the packet type.
func (p *Auth) Type() Type {
	return TypeAuth
}

// Size returns the AUTH packet size.
func (p *Auth) Size() int {
	size := p.remainingLength()
	header := FixedHeader{PacketType: TypeAuth, RemainingLength: size}

	size += header.Size()
	return size
}

// Decode decodes the AUTH packet from buf and header. This method returns the number of bytes read
// from buf and the error, if it fails to read the packet correctly.
func (p *Auth) Decode(buf []byte, h FixedHeader) (n int, err error) {
	if h.PacketType != TypeAuth {
		return 0, fmt.Errorf("%w: invalid packet type", ErrMalformedPacket)
	}
	if h.Flags != 0 {
		return 0, fmt.Errorf("%w: invalid control flags", ErrMalformedPacket)
	}
	if len(buf) != h.RemainingLength {
		return 0, fmt.Errorf("%w: buffer length does not match remaining length", ErrMalformedPacket)
	}

	if h.RemainingLength == 0 {
		p.Code = ReasonCodeSuccess
		p.Properties = nil
		return 0, nil
	}

	p.Code = ReasonCode(buf[n])
	n++

	var size int

	p.Properties, size, err = decodeProperties[AuthProperties](buf[n:])
	if err != nil {
		return 0, fmt.Errorf("%w: auth properties: %s", ErrMalformedPacket, err.Error())
	}
	n += size

	return n, p.Validate()
}

// Encode encodes the AUTH packet into buf and returns the number of bytes encoded. The buffer must have the
// length greater than or equals to the packet size, otherwise this method returns an error.
func (p *Auth) Encode(buf []byte) (n int, err error) {
	if len(buf) < p.Size() {
		return 0, errors.New("buffer too small")
	}

	err = p.Validate()
	if err != nil {
		return 0, err
	}

	header := FixedHeader{PacketType: TypeAuth, RemainingLength: p.remainingLength()}
	n = header.encode(buf)

	if header.RemainingLength == 0 {
		return n, nil
	}

	buf[n] = byte(p.Code)
	n++

	var size int

	size, err = p.Properties.encode(buf[n:])
	if err != nil {
		return 0, fmt.Errorf("%w: auth properties: %s", ErrMalformedPacket, err.Error())
	}
	n += size

	return n, err
}

// Validate validates if the CONNACK packet is valid.
func (p *Auth) Validate() error {
	var valid bool
	validCodes := [3]ReasonCode{ReasonCodeSuccess, ReasonCodeContinueAuthentication, ReasonCodeReAuthenticate}

	for _, code := range validCodes {
		if p.Code == code {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("%w: invalid reason code", ErrMalformedPacket)
	}

	err := p.Properties.Validate()
	if err != nil {
		return fmt.Errorf("%w: invalid auth properties: %s ", ErrMalformedPacket, err.Error())
	}

	return nil
}

func (p *Auth) remainingLength() int {
	size := p.Properties.size()
	if size == 0 && p.Code == ReasonCodeSuccess {
		return 0
	}
	return size + sizeVarInteger(size) + 1 // +1 for the reason code.
}

// AuthProperties contains the properties of the AUTH packet.
type AuthProperties struct {
	// Flags indicates which properties are present.
	Flags PropertyFlags `json:"flags"`

	// UserProperties is a list of user properties.
	UserProperties []UserProperty `json:"user_properties"`

	// AuthenticationMethod contains the name of the authentication method.
	AuthenticationMethod []byte `json:"authentication_method"`

	// AuthenticationData contains the authentication data.
	AuthenticationData []byte `json:"authentication_data"`

	// ReasonString represents the reason associated with the response.
	ReasonString []byte `json:"reason_string"`
}

// Has returns whether the property is present or not.
func (p *AuthProperties) Has(id PropertyID) bool {
	if p == nil {
		return false
	}
	return p.Flags.Has(id)
}

// Set sets the property indicating that it's present.
func (p *AuthProperties) Set(id PropertyID) {
	if p == nil {
		return
	}
	p.Flags = p.Flags.Set(id)
}

// Validate validates if the properties are valid.
func (p *AuthProperties) Validate() error {
	if p == nil {
		return nil
	}

	if err := validatePropString(p.Flags, PropertyAuthenticationMethod, p.AuthenticationMethod); err != nil {
		return errors.New("missing authentication method")
	}
	if err := validatePropString(p.Flags, PropertyAuthenticationData, p.AuthenticationData); err != nil {
		return errors.New("missing authentication data")
	}
	if err := validatePropString(p.Flags, PropertyReasonString, p.ReasonString); err != nil {
		return errors.New("missing reason string")
	}
	if err := validatePropUserProperty(p.Flags, p.UserProperties); err != nil {
		return err
	}

	return nil
}

func (p *AuthProperties) size() int {
	if p == nil {
		return 0
	}

	size := sizePropAuthenticationMethod(p.Flags, p.AuthenticationMethod)
	size += sizePropAuthenticationData(p.Flags, p.AuthenticationData)
	size += sizePropReasonString(p.Flags, p.ReasonString)
	size += sizePropUserProperties(p.Flags, p.UserProperties)
	return size
}

func (p *AuthProperties) decodeProperty(id PropertyID, buf []byte) (n int, err error) {
	switch id {
	case PropertyAuthenticationMethod:
		p.AuthenticationMethod, n, err = decodePropAuthenticationMethod(buf, p)
	case PropertyAuthenticationData:
		p.AuthenticationData, n, err = decodePropAuthenticationData(buf, p)
	case PropertyReasonString:
		p.ReasonString, n, err = decodePropReasonString(buf, p)
	case PropertyUserProperty:
		var user UserProperty
		user, n, err = decodePropUserProperty(buf, p)
		if err == nil {
			p.UserProperties = append(p.UserProperties, user)
		}
	default:
		err = fmt.Errorf("%w: invalid auth property: %v", ErrMalformedPacket, id)
	}

	return n, err
}

func (p *AuthProperties) encode(buf []byte) (n int, err error) {
	n = encodeVarInteger(buf, p.size())

	if p == nil {
		return n, nil
	}

	var size int

	size, err = encodePropString(buf[n:], p.Flags, PropertyAuthenticationMethod, p.AuthenticationMethod)
	if err != nil {
		return 0, fmt.Errorf("authentication method: %w", err)
	}
	n += size

	size = encodePropBinary(buf[n:], p.Flags, PropertyAuthenticationData, p.AuthenticationData)
	n += size

	size, err = encodePropString(buf[n:], p.Flags, PropertyReasonString, p.ReasonString)
	if err != nil {
		return 0, fmt.Errorf("reason string: %w", err)
	}
	n += size

	size, err = encodePropUserProperties(buf[n:], p.Flags, p.UserProperties, err)
	if err != nil {
		return 0, fmt.Errorf("user properties: %w", err)
	}
	n += size

	return n, nil
}
