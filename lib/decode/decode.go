// Package decode defines decoder components
package decode

import (
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// Decode decodes an N-bit address into 2^N invidual bits.
func Decode(group *group.Group, a []*wire.Wire) []*wire.Wire {
	var s []*wire.Wire
	for address := 0; address < 1<<len(a); address++ {
		si := group.True()
		for i, ai := range a {
			if address>>i&1 == 1 {
				si = gate.And(group, si, ai)
			} else {
				si = gate.And(group, si, gate.Not(group, ai))
			}
		}
		si.Name = sfmt.Sprintf("%s-s%d", group.Name, address)
		s = append(s, si)
	}
	return s
}
