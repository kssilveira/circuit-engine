// Package ram defines random access memory components.
package ram

import (
	"github.com/kssilveira/circuit-engine/group"
	"github.com/kssilveira/circuit-engine/lib/gate"
	"github.com/kssilveira/circuit-engine/lib/reg"
	"github.com/kssilveira/circuit-engine/sfmt"
	"github.com/kssilveira/circuit-engine/wire"
)

// RAM adds a random access memory.
func RAM(parent *group.Group, a, d []*wire.Wire, ei, eo *wire.Wire) []*wire.Wire {
	group := parent.Group("RAM")
	s := ramAddress(group, a)
	rei, reo := ramEnable(group, s, ei, eo)
	return ramRegisters(group, d, rei, reo)
}

func ramAddress(group *group.Group, a []*wire.Wire) []*wire.Wire {
	var s []*wire.Wire
	for address := 0; address < 1<<len(a); address++ {
		si := &wire.Wire{}
		si.Bit.Set(true)
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

func ramEnable(group *group.Group, s []*wire.Wire, ei, eo *wire.Wire) ([]*wire.Wire, []*wire.Wire) {
	var rei, reo []*wire.Wire
	for i, si := range s {
		reii := gate.And(group, ei, si)
		reii.Name = sfmt.Sprintf("i%d", i)
		rei = append(rei, reii)

		reoi := gate.And(group, eo, si)
		reoi.Name = sfmt.Sprintf("o%d", i)
		reo = append(reo, reoi)
	}
	return rei, reo
}

func ramRegisters(group *group.Group, d, ei, eo []*wire.Wire) []*wire.Wire {
	var all []*wire.Wire
	for i, eii := range ei {
		ri := reg.N(group, d, eii, eo[i])
		for i := range d {
			all = append(all, ri[i])
		}
	}
	return all
}
