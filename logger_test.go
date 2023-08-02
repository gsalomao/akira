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

import "testing"

func TestNoOpLoggerLogWithoutFieldsDontPanic(t *testing.T) {
	l := &noOpLogger{}
	l.Log("No panic!")
}

func TestNoOpLoggerLogWithEvenNumberOfFieldsDontPanic(t *testing.T) {
	l := &noOpLogger{}
	l.Log("No panic!", "hello", "world")
}

func TestNoOpLoggerLogWithOddNumberOfFieldsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Logger did not panic")
		}
	}()

	l := &noOpLogger{}
	l.Log("Panic!", "hello")
}
