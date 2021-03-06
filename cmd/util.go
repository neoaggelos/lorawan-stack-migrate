// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"go.thethings.network/lorawan-stack-migrate/pkg/source"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func toJSON(dev *ttnpb.EndDevice) ([]byte, error) {
	return jsonpb.TTN().Marshal(dev)
}

var (
	errExport = errors.Define("export", "export device `{device}`")
	errFormat = errors.DefineCorruption("format", "format device `{device}`")
)

func exportDev(s source.Source, devEui string) error {
	dev, err := s.ExportDevice(devEui)
	if err != nil {
		return errExport.WithAttributes("device", devEui).WithCause(err)
	}
	b, err := toJSON(dev)
	if err != nil {
		return errFormat.WithAttributes("device", devEui).WithCause(err)
	}
	_, err = fmt.Fprintln(os.Stdout, string(b))
	return err
}

// Iterator returns items
type Iterator interface {
	// Next returns the next item from the iterator. io.EOF is returned when no more items are left.
	Next() (item string, err error)
}

type listIterator struct {
	items []string
	index int
}

type readerIterator struct {
	rd  *bufio.Reader
	sep byte
}

// NewListIterator returns a new iterator from a list of items.
func NewListIterator(items []string) Iterator {
	return &listIterator{items: items}
}

func (l *listIterator) Next() (string, error) {
	if l.index < len(l.items) {
		l.index++
		return l.items[l.index-1], nil
	}
	return "", io.EOF
}

// NewReaderIterator returns a new iterator from a reader.
func NewReaderIterator(rd io.Reader, sep byte) Iterator {
	return &readerIterator{rd: bufio.NewReader(rd), sep: sep}
}

func (r *readerIterator) Next() (string, error) {
	s, err := r.rd.ReadString(r.sep)
	if err == io.EOF && s != "" {
		return s, nil
	}
	return strings.Trim(s, string(r.sep)), err
}
