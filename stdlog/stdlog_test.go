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

package stdlog

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewSetsDefaultWriter(t *testing.T) {
	l := New()
	if l == nil {
		t.Fatal("A logger was expected")
	}
	if l.logger.Writer() != os.Stderr {
		t.Error("The default writer should be the Stderr")
	}
}

func TestStdLogLogDefaultOptions(t *testing.T) {
	out := bytes.NewBufferString("")
	l := New(WithWriter(out))
	msg := "This is a message"

	l.Log(msg)
	if !strings.Contains(out.String(), msg) {
		t.Errorf("Log %s should have '%s'", out.String(), msg)
	}
}

func TestStdLogLogWithLevel(t *testing.T) {
	out := bytes.NewBufferString("")
	level := "DEBUG"
	l := New(
		WithWriter(out),
		WithLevel(level),
	)

	l.Log("This is a message")
	if !strings.Contains(out.String(), level) {
		t.Errorf("Log %s should have '%s'", out.String(), level)
	}
}

func TestStdLogLogWithName(t *testing.T) {
	out := bytes.NewBufferString("")
	name := "akira"
	l := New(
		WithWriter(out),
		WithName(name),
	)

	l.Log("This is a message")
	if !strings.Contains(out.String(), name) {
		t.Errorf("Log %s should have '%s'", out.String(), name)
	}
}

func TestStdLogLogWithFields(t *testing.T) {
	testCases := []struct {
		name     string
		value    any
		expected string
	}{
		{"String", "hello", "{field=hello}"},
		{"Integer", 10, "{field=10}"},
		{"Error", errors.New("failed"), "{field=failed}"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := bytes.NewBufferString("")
			l := New(
				WithWriter(out),
				WithFields(),
			)

			l.Log("This is a message", "field", tc.value)
			if !strings.Contains(out.String(), tc.expected) {
				t.Errorf("Log %s should have '%s'", out.String(), tc.expected)
			}
		})
	}
}

func TestStdLogLogWithMultipleFields(t *testing.T) {
	out := bytes.NewBufferString("")
	l := New(
		WithWriter(out),
		WithFields(),
	)

	l.Log("This is a message", "field1", "hello", "field2", "world")

	expected := "{field1=hello, field2=world}"
	if !strings.Contains(out.String(), expected) {
		t.Errorf("Log %s should have '%s'", out.String(), expected)
	}
}

func TestStdLogLogWithColors(t *testing.T) {
	out := bytes.NewBufferString("")
	l := New(
		WithWriter(out),
		WithColors(),
	)

	l.Log("This is a message")
	if !strings.Contains(out.String(), string(Cyan)) {
		t.Errorf("Log %s should have colors", out.String())
	}
}

func TestStdLogLogWithLevelColor(t *testing.T) {
	out := bytes.NewBufferString("")
	color := Blue
	l := New(
		WithWriter(out),
		WithColors(),
		WithLevel("DEBUG"),
		WithLevelColor(color),
	)

	l.Log("This is a message")
	if !strings.Contains(out.String(), string(color)) {
		t.Errorf("Log %s should have level with %scolor%s", out.String(), string(color), string(NoColor))
	}
}

func TestStdLogLogWithSeparator(t *testing.T) {
	out := bytes.NewBufferString("")
	separator := "|"
	l := New(
		WithWriter(out),
		WithSeparator(separator),
	)

	l.Log("This is a message")
	if !strings.Contains(out.String(), separator) {
		t.Errorf("Log %s should have '%s'", out.String(), separator)
	}
}

func BenchmarkStdLogLog(b *testing.B) {
	b.Run("Default Options", func(b *testing.B) {
		l := New(WithWriter(io.Discard))
		for i := 0; i < b.N; i++ {
			l.Log("This is a message", "field1", "hello", "field2", "world")
		}
	})

	b.Run("Full Options", func(b *testing.B) {
		l := New(
			WithWriter(io.Discard),
			WithColors(),
			WithFields(),
			WithName("akira"),
			WithLevel("DEBUG"),
			WithLevelColor(Blue),
			WithSeparator("|"),
		)
		for i := 0; i < b.N; i++ {
			l.Log("This is a message", "field1", "hello", "field2", "world")
		}
	})
}
