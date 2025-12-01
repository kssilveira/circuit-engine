// Package bit encapsulates individual bits.
package bit

import (
	"github.com/kssilveira/circuit-engine/component"
)

// Bit contains a single bit.
type Bit struct {
	bit     bool
	readers []component.Component
}

// Get returns the bit and updates the list of readers.
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

// Set sets the bit and updates the readers.
func (b *Bit) Set(v bool) {
	if v != b.bit {
		b.bit = v
		for _, reader := range b.readers {
			reader.Update()
		}
	}
}

// SilentSet sets the bit without updating the readers.
func (b *Bit) SilentSet(v bool) {
	b.bit = v
}
