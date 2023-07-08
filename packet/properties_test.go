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
	"testing"

	"github.com/stretchr/testify/suite"
)

type PropertiesTestSuite struct {
	suite.Suite
}

func (s *PropertiesTestSuite) TestDecodePropertiesNoProperties() {
	data := []byte{0}

	p, n, err := decodeProperties[PropertiesConnect](data)
	s.Require().NoError(err)
	s.Assert().Nil(p)
	s.Assert().Equal(1, n)
}

func (s *PropertiesTestSuite) TestDecodePropertiesErrorPropertyLength() {
	data := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	_, _, err := decodeProperties[PropertiesConnect](data)
	s.Require().Error(err)
}

func (s *PropertiesTestSuite) TestDecodePropertiesInvalidPropertyType() {
	data := []byte{1}

	p, n, err := decodeProperties[int](data)
	s.Require().Error(err)
	s.Assert().Nil(p)
	s.Assert().Equal(1, n)
}

func TestPropertiesTestSuite(t *testing.T) {
	suite.Run(t, new(PropertiesTestSuite))
}
