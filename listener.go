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
	"errors"
	"net"
	"sync"
)

// ErrListenerAlreadyExists indicates that the Listener already exists based on its name.
var ErrListenerAlreadyExists = errors.New("listener already exists")

// OnConnectionFunc is the function which the Listener must call when a new connection has been opened.
type OnConnectionFunc func(Listener, net.Conn)

// Listener is an interface which all network listeners must implement. A network listener is responsible for listen
// for network connections and notify any incoming connection.
type Listener interface {
	// Name returns the name of the listener. The Server supports only one listener for each name. If a listener is
	// added into the server with the same name of another listener, the ErrListenerAlreadyExists is returned.
	Name() string

	// Address returns the network address of the listener.
	Address() string

	// Protocol returns the network protocol of the listener.
	Protocol() string

	// Listen starts the listener. The listener calls the OnConnectionFunc for any received incoming connection.
	// This function does not block the caller and returns a channel, which an event is sent when the listener is
	// ready for accept incoming connection, and closed when the listener has stopped.
	// It returns an error if the Listener fails to start.
	Listen(OnConnectionFunc) (<-chan bool, error)

	// Stop stops the listener. When the listener is stopped, the channel returned by the Listen method is closed,
	// and the listener does not accept any other incoming connection.
	Stop()

	// Listening returns whether the listener is listening for incoming connection or not.
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

func (l *listeners) listen(lsn Listener, f OnConnectionFunc) error {
	l.RLock()
	defer l.RUnlock()

	listening, err := lsn.Listen(f)
	if err != nil {
		return err
	}

	<-listening
	l.wg.Add(1)

	go func() {
		defer l.wg.Done()

		// Wait while Listener is listening.
		<-listening
	}()

	return nil
}

func (l *listeners) listenAll(f OnConnectionFunc) error {
	l.RLock()
	defer l.RUnlock()

	for _, lsn := range l.internal {
		err := l.listen(lsn, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *listeners) stopAll() {
	l.RLock()
	defer l.RUnlock()

	for _, lsn := range l.internal {
		lsn.Stop()
	}
}

func (l *listeners) wait() {
	l.wg.Wait()
}
