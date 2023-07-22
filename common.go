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

import "github.com/gsalomao/akira/packet"

func newIfNotExists[T any](v *T) *T {
	if v != nil {
		return v
	}
	return new(T)
}

func maxIfExceeded[T ~int | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64](val, max T) T {
	if max > 0 && val > max {
		return max
	}
	return val
}

func isClientIDValid(version packet.Version, clientIDSize int, config *Config) bool {
	if clientIDSize == 0 {
		return false
	}
	if version == packet.MQTT31 && clientIDSize > 23 {
		return false
	}
	if clientIDSize > config.MaxClientIDSize {
		return false
	}
	return true
}

func isKeepAliveValid(version packet.Version, keepAlive, maxKeepAlive uint16) bool {
	// For V5.0 clients, the Keep Alive sent in the CONNECT Packet can be overwritten in the CONNACK Packet.
	if version == packet.MQTT50 || maxKeepAlive == 0 {
		return true
	}
	if keepAlive > 0 && keepAlive <= maxKeepAlive {
		return true
	}
	return false
}

func sessionKeepAlive(keepAlive, maxKeepAlive uint16) uint16 {
	if maxKeepAlive > 0 && (keepAlive == 0 || keepAlive > maxKeepAlive) {
		return maxKeepAlive
	}
	return keepAlive
}

func hasSessionExpiryInterval(s *Session, connect *packet.Connect) bool {
	if s != nil && s.Properties.Has(packet.PropertySessionExpiryInterval) {
		if connect != nil && s.Properties.SessionExpiryInterval != connect.Properties.SessionExpiryInterval {
			return true
		}
	}
	return false
}

func sessionProperties(props *packet.ConnectProperties, conf *Config) *SessionProperties {
	if props == nil {
		return nil
	}

	p := &SessionProperties{}

	if props.Has(packet.PropertySessionExpiryInterval) {
		interval := props.SessionExpiryInterval
		p.SessionExpiryInterval = maxIfExceeded(interval, conf.MaxSessionExpiryInterval)
		p.Set(packet.PropertySessionExpiryInterval)
	}
	if props.Has(packet.PropertyMaximumPacketSize) {
		size := props.MaximumPacketSize
		p.MaximumPacketSize = maxIfExceeded(size, conf.MaxPacketSize)
		p.Set(packet.PropertyMaximumPacketSize)
	}
	if props.Has(packet.PropertyReceiveMaximum) {
		maximum := props.ReceiveMaximum
		p.ReceiveMaximum = maxIfExceeded(maximum, conf.MaxInflightMessages)
		p.Set(packet.PropertyReceiveMaximum)
	}
	if props.Has(packet.PropertyTopicAliasMaximum) {
		maximum := props.TopicAliasMaximum
		p.TopicAliasMaximum = maxIfExceeded(maximum, conf.TopicAliasMax)
		p.Set(packet.PropertyTopicAliasMaximum)
	}
	if props.Has(packet.PropertyRequestResponseInfo) {
		p.RequestResponseInfo = props.RequestResponseInfo
		p.Set(packet.PropertyRequestResponseInfo)
	}
	if props.Has(packet.PropertyRequestProblemInfo) {
		p.RequestProblemInfo = props.RequestProblemInfo
		p.Set(packet.PropertyRequestProblemInfo)
	}
	if props.Has(packet.PropertyUserProperty) {
		p.UserProperties = props.UserProperties
		p.Set(packet.PropertyUserProperty)
	}

	return p
}

func lastWill(connect *packet.Connect) *LastWill {
	if !connect.Flags.WillFlag() {
		return nil
	}

	will := &LastWill{
		Topic:   connect.WillTopic,
		Payload: connect.WillPayload,
		QoS:     connect.Flags.WillQoS(),
		Retain:  connect.Flags.WillRetain(),
	}

	if connect.WillProperties == nil {
		return nil
	}

	props := connect.WillProperties
	will.Properties = &packet.WillProperties{Flags: props.Flags}

	if props.Has(packet.PropertyWillDelayInterval) {
		will.Properties.WillDelayInterval = props.WillDelayInterval
		will.Properties.Set(packet.PropertyWillDelayInterval)
	}
	if props.Has(packet.PropertyPayloadFormatIndicator) {
		will.Properties.PayloadFormatIndicator = props.PayloadFormatIndicator
		will.Properties.Set(packet.PropertyPayloadFormatIndicator)
	}
	if props.Has(packet.PropertyMessageExpiryInterval) {
		will.Properties.MessageExpiryInterval = props.MessageExpiryInterval
		will.Properties.Set(packet.PropertyMessageExpiryInterval)
	}
	if props.Has(packet.PropertyContentType) {
		will.Properties.ContentType = props.ContentType
		will.Properties.Set(packet.PropertyContentType)
	}
	if props.Has(packet.PropertyResponseTopic) {
		will.Properties.ResponseTopic = props.ResponseTopic
		will.Properties.Set(packet.PropertyResponseTopic)
	}
	if props.Has(packet.PropertyCorrelationData) {
		will.Properties.CorrelationData = props.CorrelationData
		will.Properties.Set(packet.PropertyCorrelationData)
	}
	if props.Has(packet.PropertyUserProperty) {
		will.Properties.UserProperties = props.UserProperties
		will.Properties.Set(packet.PropertyUserProperty)
	}

	return will
}

func connAckProperties(c *Client, conf *Config, connect *packet.Connect) *packet.ConnAckProperties {
	var props *packet.ConnAckProperties

	if hasSessionExpiryInterval(c.Session, connect) {
		props = newIfNotExists(props)
		props.SessionExpiryInterval = c.Session.Properties.SessionExpiryInterval
		props.Set(packet.PropertySessionExpiryInterval)
	}
	if connect != nil && connect.KeepAlive != c.Connection.KeepAlive {
		props = newIfNotExists(props)
		props.ServerKeepAlive = c.Connection.KeepAlive
		props.Set(packet.PropertyServerKeepAlive)
	}
	if conf.MaxInflightMessages > 0 {
		props = newIfNotExists(props)
		props.ReceiveMaximum = conf.MaxInflightMessages
		props.Set(packet.PropertyReceiveMaximum)
	}
	if conf.MaxPacketSize > 0 {
		props = newIfNotExists(props)
		props.MaximumPacketSize = conf.MaxPacketSize
		props.Set(packet.PropertyMaximumPacketSize)
	}
	if conf.MaxQoS < byte(packet.QoS2) {
		props = newIfNotExists(props)
		props.MaximumQoS = conf.MaxQoS
		props.Set(packet.PropertyMaximumQoS)
	}
	if conf.TopicAliasMax > 0 {
		props = newIfNotExists(props)
		props.TopicAliasMaximum = conf.TopicAliasMax
		props.Set(packet.PropertyTopicAliasMaximum)
	}
	if !conf.RetainAvailable {
		props = newIfNotExists(props)
		props.RetainAvailable = false
		props.Set(packet.PropertyRetainAvailable)
	}
	if !conf.WildcardSubscriptionAvailable {
		props = newIfNotExists(props)
		props.WildcardSubscriptionAvailable = false
		props.Set(packet.PropertyWildcardSubscriptionAvailable)
	}
	if !conf.SubscriptionIDAvailable {
		props = newIfNotExists(props)
		props.SubscriptionIDAvailable = false
		props.Set(packet.PropertySubscriptionIDAvailable)
	}
	if !conf.SharedSubscriptionAvailable {
		props = newIfNotExists(props)
		props.SharedSubscriptionAvailable = false
		props.Set(packet.PropertySharedSubscriptionAvailable)
	}
	return props
}
