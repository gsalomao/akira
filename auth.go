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
	"context"
	"errors"
	"sync"

	"github.com/gsalomao/akira/packet"
)

// ErrEnhancedAuthAlreadyExists indicates that the EnhancedAuth already exists based on its name.
var ErrEnhancedAuthAlreadyExists = errors.New("enhanced auth already exists")

// EnhancedAuth is the interface to implement enhanced authentication.
//
// An enhanced authentication is identified by its name, which must match with the name of the authentication method
// sent by clients. If the client tries to connect with the server using an authentication method which does not match
// with the name of any registered EnhancedAuth, the server rejects the connection with the reason code Bad
// Authentication Method.
type EnhancedAuth interface {
	// Name returns the name of the enhanced authentication. This name must match with the name of the authentication
	// method sent by clients.
	Name() string

	// Authenticate authenticates the client.
	//
	// When the client sends a CONNECT packet, or an AUTH packet, which the authentication method matches with the
	// name of the EnhancedAuth, this method is called. The corresponding packet received from client is passed as p.
	//
	// If the authentication fails, this method must return the error. If the error is a valid packet.Error, a CONNACK
	// packet or an AUTH packet is sent to the client before the connection is closed. If this method doesn't return
	// any error, the process that triggered this authentication continues.
	//
	// When this enhanced authentication involves the exchange of AUTH packets, this method must return the AUTH packet
	// to be sent to client and no error.
	//
	// When this method is called due to a CONNECT packet and this method returns a CONNACK packet, the returned
	// CONNACK packet is not sent directly to the client, but its reason code, authentication data, and the reason
	// string are used to create the final CONNACK packet which will be sent to the client.
	//
	// If this method returns any packet but CONNACK and AUTH packets, no packet is sent to the client and the
	// connection is closed.
	Authenticate(ctx context.Context, c *Client, p Packet) (PacketEncodable, error)
}

type auths struct {
	mutex    sync.RWMutex
	enhanced map[string]EnhancedAuth
}

func newAuths() *auths {
	return &auths{enhanced: make(map[string]EnhancedAuth)}
}

func (a *auths) addEnhancedAuth(e EnhancedAuth) error {
	name := e.Name()

	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, ok := a.enhanced[name]; ok {
		return ErrEnhancedAuthAlreadyExists
	}

	a.enhanced[name] = e
	return nil
}

func (a *auths) authenticateEnhanced(ctx context.Context, c *Client, p Packet, method string) (PacketEncodable, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	auth, ok := a.enhanced[method]
	if !ok {
		return nil, packet.ErrBadAuthenticationMethod
	}

	return auth.Authenticate(ctx, c, p)
}
