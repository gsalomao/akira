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

type PacketFixture struct {
	Name   string `json:"name"`
	Packet []byte `json:"packet"`
}

func ReadPacketFixture(file, name string) (PacketFixture, error) {
	data, err := os.ReadFile("testdata/" + file)
	if err != nil {
		return PacketFixture{}, err
	}

	var fixtures []PacketFixture
	err = json.Unmarshal(data, &fixtures)
	if err != nil {
		return PacketFixture{}, err
	}

	for i := range fixtures {
		if fixtures[i].Name == name {
			return fixtures[i], nil
		}
	}

	return PacketFixture{}, errors.New("packet fixture not found")
}
