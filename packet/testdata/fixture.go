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

package testdata

import (
	"encoding/json"
	"errors"
	"os"
)

const (
	ConnectFlagReserved   = "Reserved"
	ConnectFlagCleanStart = "CleanStart"
	ConnectFlagWillFlag   = "WillFlag"
	ConnectFlagWillQoS0   = "WillQoS0"
	ConnectFlagWillQoS1   = "WillQoS1"
	ConnectFlagWillQoS2   = "WillQoS2"
	ConnectFlagWillRetain = "WillRetain"
	ConnectFlagPassword   = "Password"
	ConnectFlagUsername   = "Username"
)

type Fixture struct {
	Name   string   `json:"name"`
	Packet []byte   `json:"packet"`
	Flags  []string `json:"flags"`
}

func ReadFixtures(path string) ([]Fixture, error) {
	data, err := os.ReadFile("testdata/" + path)
	if err != nil {
		return nil, err
	}

	var fixtures []Fixture
	err = json.Unmarshal(data, &fixtures)
	if err != nil {
		return nil, err
	}

	return fixtures, nil
}

func ReadFixture(path, name string) (Fixture, error) {
	fixtures, err := ReadFixtures(path)
	if err != nil {
		return Fixture{}, err
	}

	for i := range fixtures {
		if fixtures[i].Name == name {
			return fixtures[i], nil
		}
	}

	return Fixture{}, errors.New("fixture not found")
}
