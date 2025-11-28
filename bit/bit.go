package bit

import (
	"github.com/kssilveira/circuit-engine/component"
)

type Bit struct {
	bit     bool
	readers []component.Component
}

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

func (b *Bit) Set(v bool) {
	if v != b.bit {
		b.bit = v
		for _, reader := range b.readers {
			reader.Update()
		}
	}
}

func (b *Bit) SilentSet(v bool) {
	b.bit = v
}
