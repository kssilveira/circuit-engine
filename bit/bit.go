// Package bit encapsulates individual bits.
package bit

import (
	"github.com/kssilveira/circuit-engine/component"
	"github.com/kssilveira/circuit-engine/config"
	"github.com/kssilveira/circuit-engine/sfmt"
)

// Bit contains a single bit.
type Bit struct {
	bit     bool
	readers []component.Component
	writer  component.Component
}

// Get returns the bit, updates the writer and stores the reader.
func (b *Bit) Get(reader component.Component) bool {
	if reader == nil {
		return b.bit
	}
	found := false
	// maybe use a set to avoid this linear search
	for _, r := range b.readers {
		if r == reader {
			found = true
			break
		}
	}
	if !found {
		b.readers = append(b.readers, reader)
	}
	return b.bit
}

// SilentGet returns the bit without  updating the writer and storing the reader.
func (b *Bit) SilentGet() bool {
	return b.bit
}

// Set sets the bit, stores the writer and updates the readers.
func (b *Bit) Set(v bool, writer component.Component, updateReaders bool) {
	if v != b.bit {
		b.bit = v
		if updateReaders {
			for _, reader := range b.readers {
				reader.Update(updateReaders)
			}
		}
	}
	if writer == nil {
		return
	}
	if b.writer != nil && writer != b.writer {
		panic(sfmt.Sprintf("multiple writers %#v %#v", b.writer.String(0, config.Config{}), writer.String(0, config.Config{})))
	}
	b.writer = writer
}

// SilentSet sets the bit without updating the readers.
func (b *Bit) SilentSet(v bool) {
	b.bit = v
}
