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
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Log level color.
const (
	NoColor LevelColor = "\x1b[0m"
	Red     LevelColor = "\x1b[31m"
	Green   LevelColor = "\x1b[32m"
	Yellow  LevelColor = "\x1b[33m"
	Blue    LevelColor = "\x1b[34m"
	White   LevelColor = "\x1b[37m"
	Cyan    LevelColor = "\x1b[36m"
	Gray    LevelColor = "\x1b[90m"
)

// LevelColor indicates the color to be used when logging the level.
type LevelColor string

// Options is the options to create the StdLog.
type Options struct {
	// Writer is the writer where the logs will be written to.
	Writer io.Writer

	// Name is the name of the logger.
	Name string

	// Level is the level to be used when creating the logs.
	Level string

	// LevelColor is the color to be used when logging the level.
	LevelColor LevelColor

	// Separator is a string to be used to separate different sections of the log.
	Separator string

	// WithFields indicates whether the logs should be created with the fields or not.
	WithFields bool

	// WithColors indicates whether the logs should be created with colors or not.
	WithColors bool
}

// OptionFunc is a function responsible to set an option.
type OptionFunc func(opts *Options)

// WithName sets the logger name into the Options.
func WithName(n string) OptionFunc {
	return func(opts *Options) {
		opts.Name = n
	}
}

// WithLevel sets the level to be used when creating the logs into the Options.
func WithLevel(l string) OptionFunc {
	return func(opts *Options) {
		opts.Level = l
	}
}

// WithFields sets to log the fields into the Options.
func WithFields() OptionFunc {
	return func(opts *Options) {
		opts.WithFields = true
	}
}

// WithColors sets to log using colors into the Options.
func WithColors() OptionFunc {
	return func(opts *Options) {
		opts.WithColors = true
	}
}

// WithLevelColor sets the color to be used when logging the level into the Options.
func WithLevelColor(c LevelColor) OptionFunc {
	return func(opts *Options) {
		opts.LevelColor = c
	}
}

// WithSeparator sets the separator to be used when creating the logs into the Options.
func WithSeparator(s string) OptionFunc {
	return func(opts *Options) {
		opts.Separator = s
	}
}

// WithWriter sets the writer where the logs should be written to into the Options.
func WithWriter(w io.Writer) OptionFunc {
	return func(opts *Options) {
		opts.Writer = w
	}
}

// StdLog is a logger responsible for logging the messages using the standard library.
type StdLog struct {
	logger     *log.Logger
	name       string
	level      string
	levelColor LevelColor
	separator  string
	withColors bool
	withFields bool
}

// New returns a new StdLog.
func New(opts ...OptionFunc) *StdLog {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return NewWithOptions(options)
}

// NewWithOptions returns a new StdLog with the provided Options.
func NewWithOptions(opts *Options) *StdLog {
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}
	return &StdLog{
		logger:     log.New(opts.Writer, "", 0),
		name:       opts.Name,
		level:      opts.Level,
		withFields: opts.WithFields,
		withColors: opts.WithColors,
		levelColor: opts.LevelColor,
		separator:  opts.Separator,
	}
}

// Log logs the given message and variable number of fields using the standard library.
func (s *StdLog) Log(msg string, fields ...any) {
	e := logEvent{
		timestamp:  time.Now(),
		name:       s.name,
		level:      s.level,
		levelColor: s.levelColor,
		msg:        msg,
		fields:     fields,
		withColors: s.withColors,
		withFields: s.withFields,
		separator:  s.separator,
	}
	s.logger.Println(e.string())
}

type logEvent struct {
	timestamp  time.Time
	name       string
	level      string
	levelColor LevelColor
	msg        string
	separator  string
	fields     []any
	withColors bool
	withFields bool
}

func (e *logEvent) string() string {
	b := strings.Builder{}

	e.writeColor(&b, White)
	b.WriteString(e.timestamp.Format("2006-01-02 15:04:05.000000 -0700"))
	e.writeColor(&b, NoColor)
	e.writeSeparator(&b)

	if e.level != "" {
		e.writeColor(&b, e.levelColor)
		b.WriteString(e.level)
		e.writeColor(&b, NoColor)
		e.writeSeparator(&b)
	}

	e.writeColor(&b, Cyan)
	if e.name != "" {
		b.WriteString("(")
		b.WriteString(e.name)
		b.WriteString(") ")
	}

	b.WriteString(e.msg)

	if e.withFields {
		e.writeColor(&b, Gray)

		if len(e.fields) > 0 {
			b.WriteString(" {")
		}

		for i := 0; i < len(e.fields); i += 2 {
			k := e.fields[i].(string)

			b.WriteString(k)
			b.WriteString("=")

			switch v := e.fields[i+1].(type) {
			case string:
				b.WriteString(v)
			case error:
				b.WriteString(v.Error())
			default:
				b.WriteString(fmt.Sprintf("%+v", v))
			}

			if i < len(e.fields)-2 {
				b.WriteString(", ")
			}
		}

		if len(e.fields) > 0 {
			b.WriteString("}")
		}
	}

	e.writeColor(&b, NoColor)
	return b.String()
}

func (e *logEvent) writeColor(b *strings.Builder, color LevelColor) {
	if !e.withColors {
		return
	}
	b.WriteString(string(color))
}

func (e *logEvent) writeSeparator(b *strings.Builder) {
	b.WriteString(" ")
	if e.separator == "" {
		return
	}
	b.WriteString(e.separator)
	b.WriteString(" ")
}
