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
	"errors"
	"net"
	"sync"
)

// ErrListenerAlreadyExists indicates that the Listener already exists based on its name.
var ErrListenerAlreadyExists = errors.New("listener already exists")

// OnConnectionFunc is the function which the Listener must call when a new connection has been opened.
type OnConnectionFunc func(Listener, net.Conn)

// Listener is an interface which all network listeners must implement. A network Listener is responsible for listen
// for network connections and notify any incoming connection.
type Listener interface {
	// Name returns the name of the Listener. The Server supports only one Listener for each name. If a Listener is
	// added into the Server with the same name of another listener, the ErrListenerAlreadyExists is returned.
	Name() string

	// Address returns the network address of the Listener.
	Address() string

	// Protocol returns the network protocol of the Listener.
	Protocol() string

	// Listen starts the Listener. The Listener must call the OnConnectionFunc for any incoming connection received.
	// This function must not block the caller and returns a channel, which an event is sent when the Listener is
	// ready for accept incoming requests, and closed when the Listener is closed.
	Listen(OnConnectionFunc) <-chan bool

	// Close closes the Listener. When the Listener is closed, the channel returned by the Listen method must be
	// closed, and the Listener must not accept any other incoming connection.
	Close()

	// Listening returns whether the Listener is listening for incoming connection or not.
	Listening() bool
}

type listeners struct {
	sync.RWMutex
	internal map[string]Listener
	wg       sync.WaitGroup
}

func newListeners() *listeners {
	l := listeners{
		internal: map[string]Listener{},
	}
	return &l
}

func (l *listeners) add(lsn Listener) {
	l.Lock()
	defer l.Unlock()
	l.internal[lsn.Name()] = lsn
}

func (l *listeners) get(name string) (lsn Listener, ok bool) {
	l.RLock()
	defer l.RUnlock()
	lsn, ok = l.internal[name]
	return lsn, ok
}

func (l *listeners) delete(name string) {
	l.Lock()
	defer l.Unlock()
	delete(l.internal, name)
}

func (l *listeners) len() int {
	l.RLock()
	defer l.RUnlock()

	return len(l.internal)
}

func (l *listeners) listen(lsn Listener, f OnConnectionFunc) {
	l.RLock()
	defer l.RUnlock()

	ready := make(chan struct{})
	l.wg.Add(1)

	go func(lsn Listener) {
		defer l.wg.Done()

		listening := lsn.Listen(f)

		// Wait for Listener starts to listen.
		<-listening
		close(ready)

		// Wait while Listener is listening.
		<-listening
	}(lsn)

	<-ready
}

func (l *listeners) listenAll(f OnConnectionFunc) {
	l.RLock()
	defer l.RUnlock()

	ready := make(chan struct{}, len(l.internal))
	l.wg.Add(len(l.internal))

	for _, lsn := range l.internal {
		go func(lsn Listener) {
			defer l.wg.Done()

			listening := lsn.Listen(f)

			// Wait for Listener starts to listen.
			<-listening
			ready <- struct{}{}

			// Wait while Listener is listening.
			<-listening
		}(lsn)
	}

	for i := 0; i < len(l.internal); i++ {
		<-ready
	}
	close(ready)
}

func (l *listeners) closeAll() {
	l.RLock()
	defer l.RUnlock()

	for _, lsn := range l.internal {
		lsn.Close()
	}
}

func (l *listeners) wait() {
	l.wg.Wait()
}
